package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"malFuse/internal/engine"
	"malFuse/internal/logger"
)

type Checker interface {
	Check(ctx context.Context, req engine.Request) engine.Result
}

type Handler struct {
	engine        Checker
	streamChecker engine.StreamChecker
	routes        map[string]RouteEntry
	dbPinger      DBPinger
	startTime     time.Time
}

type DBPinger interface {
	Ping() error
}

type RouteEntry struct {
	Upstream  *url.URL
	Ecosystem string
}

func New(eng Checker, routes map[string]RouteEntry) *Handler {
	return &Handler{
		engine:    eng,
		routes:    routes,
		startTime: time.Now(),
	}
}

func (h *Handler) SetDBPinger(p DBPinger) {
	h.dbPinger = p
}

func (h *Handler) SetStreamChecker(sc engine.StreamChecker) {
	h.streamChecker = sc
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" {
		h.handleHealth(w)
		return
	}

	matched, entry, prefix := h.matchRoute(r.URL.Path)
	if !matched {
		logger.Warn("no route matched", "path", r.URL.Path)
		http.Error(w, "no route matched", http.StatusBadGateway)
		return
	}

	strippedPath := strings.TrimPrefix(r.URL.Path, prefix)
	pkgName, version := h.extractPackageInfo(strippedPath, entry.Ecosystem)

	logger.Debug("incoming request",
		"method", r.Method,
		"path", r.URL.Path,
		"route", prefix,
		"upstream", entry.Upstream.String(),
		"package", pkgName,
		"ecosystem", entry.Ecosystem,
	)

	result := h.engine.Check(r.Context(), engine.Request{
		Name:      pkgName,
		Version:   version,
		Ecosystem: entry.Ecosystem,
		RawPath:   r.URL.Path,
	})

	if result.Block {
		logger.Warn("package blocked", "package", pkgName, "reason", result.Reason)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(result.Reason))
		return
	}

	if result.Skip {
		logger.Info("package whitelisted", "package", pkgName, "ecosystem", entry.Ecosystem)
	}

	h.forward(w, r, entry, prefix)
}

func (h *Handler) matchRoute(path string) (bool, RouteEntry, string) {
	for prefix, entry := range h.routes {
		if strings.HasPrefix(path, prefix) {
			return true, entry, prefix
		}
	}
	return false, RouteEntry{}, ""
}

func (h *Handler) extractPackageInfo(path, ecosystem string) (string, string) {
	switch ecosystem {
	case "pypi":
		name, _ := extractPyPIPackageName(path)
		return name, extractPyPIVersion(path, name)
	case "npm":
		name, _ := extractNPMPackageName(path)
		return name, extractNPMVersion(path, name)
	}
	return "", ""
}

func (h *Handler) forward(w http.ResponseWriter, r *http.Request, entry RouteEntry, prefix string) {
	proxy := httputil.NewSingleHostReverseProxy(entry.Upstream)

	proxy.ModifyResponse = func(resp *http.Response) error {
		logger.Debug("upstream response",
			"status", resp.StatusCode,
			"content_type", resp.Header.Get("Content-Type"),
			"path", r.URL.Path,
		)

		// Rewrite download URLs in Simple API response to route through proxy
		if isHTML(resp.Header.Get("Content-Type")) && strings.Contains(r.URL.Path, "/simple/") {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err == nil {
				rewritten := rewriteURLs(string(body), entry.Upstream.String())
				resp.Body = io.NopCloser(strings.NewReader(rewritten))
				resp.ContentLength = int64(len(rewritten))
				resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(rewritten)))
			}
		}

		if h.streamChecker != nil && isArchive(resp.Header.Get("Content-Type")) {
			ctx, cancel := context.WithCancel(resp.Request.Context())
			pr, pw := io.Pipe()
			resp.Body = &teeReadCloser{
				reader: io.TeeReader(resp.Body, pw),
				closer: resp.Body,
			}

			go func() {
				defer pr.Close()
				result := h.streamChecker.StreamCheck(engine.Request{}, pr)
				if result.Block {
					logger.Warn("script scan blocked", "reason", result.Reason)
					cancel()
				}
			}()

			resp.Request = resp.Request.WithContext(ctx)
		}
		return nil
	}

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = entry.Upstream.Host
		req.URL.Scheme = entry.Upstream.Scheme
		req.URL.Host = entry.Upstream.Host

		req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
		if !strings.HasPrefix(req.URL.Path, "/") {
			req.URL.Path = "/" + req.URL.Path
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("upstream error", "error", err)
		http.Error(w, "upstream unreachable", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}

type teeReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func (t *teeReadCloser) Read(p []byte) (int, error) { return t.reader.Read(p) }
func (t *teeReadCloser) Close() error               { return t.closer.Close() }

func isArchive(ct string) bool {
	if ct == "" {
		return false
	}
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "application/x-tar") ||
		strings.Contains(ct, "application/gzip") ||
		strings.Contains(ct, "application/zip") ||
		strings.Contains(ct, "application/octet-stream")
}

func (h *Handler) matchingPrefix(path string) string {
	for prefix := range h.routes {
		if strings.HasPrefix(path, prefix) {
			return prefix
		}
	}
	return ""
}

func extractPyPIPackageName(path string) (string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// Simple API: /simple/<package>/
	if len(parts) >= 2 && parts[0] == "simple" {
		name := strings.ToLower(parts[1])
		return name, name != ""
	}
	// Download URL: /packages/.../<package>-<version>.tar.gz
	if len(parts) >= 2 && parts[0] == "packages" {
		// Get the last part which is <package>-<version>.ext
		last := parts[len(parts)-1]
		// Extract package name: everything before the last version-like segment
		if idx := strings.LastIndex(last, "-"); idx >= 0 {
			name := strings.ToLower(last[:idx])
			return name, name != ""
		}
	}
	return "", false
}

func extractPyPIVersion(path, pkgName string) string {
	// Pattern 2: <pkgName>-<version>.ext in filename
	base := path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			base = path[i+1:]
			break
		}
	}
	prefix := pkgName + "-"
	if strings.HasPrefix(base, prefix) {
		rest := strings.TrimPrefix(base, prefix)
		for _, ext := range []string{".tar.gz", ".whl", ".zip", ".tgz", ".tar"} {
			if strings.HasSuffix(rest, ext) {
				rest = rest[:len(rest)-len(ext)]
			}
		}
		if rest != "" && len(rest) < 30 {
			return rest
		}
	}
	// Pattern 1: /<pkgName>/<version>/ like /pandas/1.1.3/
	if idx := strings.Index(path, "/"+pkgName+"/"); idx >= 0 {
		rest := path[idx+len(pkgName)+2:]
		if slash := strings.IndexByte(rest, '/'); slash >= 0 {
			rest = rest[:slash]
		}
		rest = strings.TrimRight(rest, "/")
		if len(rest) > 0 && len(rest) < 20 {
			return rest
		}
	}
	return ""
}

func extractNPMPackageName(path string) (string, bool) {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "", false
	}
	if strings.HasPrefix(trimmed, "@") {
		parts := strings.SplitN(trimmed, "/", 3)
		if len(parts) >= 2 {
			return "@" + parts[0][1:] + "/" + parts[1], true
		}
		return "", false
	}
	return strings.Split(trimmed, "/")[0], true
}

func extractNPMVersion(path, pkgName string) string {
	trimmed := strings.Trim(path, "/")

	// pattern: /<pkg>/<version> or /@scope/pkg/<version>
	nameLen := len(pkgName)
	if len(trimmed) > nameLen+1 && strings.HasPrefix(trimmed, pkgName+"/") {
		rest := trimmed[nameLen+1:]
		// Skip tgz tarball paths like /left-pad/-/left-pad-1.0.0.tgz
		if strings.HasPrefix(rest, "-/") {
			localName := pkgName
			if idx := strings.LastIndex(pkgName, "/"); idx >= 0 {
				localName = pkgName[idx+1:]
			}
			rest = rest[2:] // skip "-/"
			prefix := localName + "-"
			if strings.HasPrefix(rest, prefix) {
				rest = strings.TrimPrefix(rest, prefix)
				for _, ext := range []string{".tgz", ".tar.gz", ".tar"} {
					if strings.HasSuffix(rest, ext) {
						rest = rest[:len(rest)-len(ext)]
						break
					}
				}
				if rest != "" && len(rest) < 30 {
					return rest
				}
			}
			return ""
		}
		rest = strings.TrimRight(rest, "/")
		if len(rest) > 0 && len(rest) < 20 {
			return rest
		}
	}

	return ""
}


func (h *Handler) handleHealth(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	status := "ok"
	dbOK := true
	if h.dbPinger != nil {
		if err := h.dbPinger.Ping(); err != nil {
			dbOK = false
			status = "degraded"
		}
	}

	resp := map[string]interface{}{
		"status": status,
		"db":     dbOK,
		"uptime": time.Since(h.startTime).String(),
	}

	json.NewEncoder(w).Encode(resp)
}

func isHTML(ct string) bool {
	return strings.Contains(strings.ToLower(ct), "text/html")
}

func rewriteURLs(body, upstream string) string {
	for _, old := range []string{
		upstream + "/packages/",
		strings.TrimSuffix(upstream, "/") + "/packages/",
	} {
		body = string(bytes.ReplaceAll([]byte(body), []byte(old), []byte("/pypi/packages/")))
	}
	return body
}
