package proxy

import (
	"context"
	"encoding/json"
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
		return name, ""
	case "npm":
		name, _ := extractNPMPackageName(path)
		return name, ""
	}
	return "", ""
}

func (h *Handler) forward(w http.ResponseWriter, r *http.Request, entry RouteEntry, prefix string) {
	proxy := httputil.NewSingleHostReverseProxy(entry.Upstream)

	if h.streamChecker != nil {
		proxy.ModifyResponse = func(resp *http.Response) error {
			ct := resp.Header.Get("Content-Type")
			if !isArchive(ct) {
				return nil
			}

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
			return nil
		}
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
	if len(parts) >= 2 && parts[0] == "simple" {
		name := strings.ToLower(parts[1])
		return name, name != ""
	}
	return "", false
}

func extractNPMPackageName(path string) (string, bool) {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "", false
	}
	if strings.HasPrefix(trimmed, "@") {
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) == 2 {
			return "@" + parts[0][1:] + "/" + parts[1], true
		}
		return "", false
	}
	return strings.Split(trimmed, "/")[0], true
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
