package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"malFuse/internal/config"
	"malFuse/internal/db/schema"
	"malFuse/internal/engine"
	"malFuse/internal/logger"
	"malFuse/internal/osv"
	"malFuse/internal/proxy"
	"malFuse/internal/scanner"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		logger.Fatal("failed to read config", "error", err)
	}

	cfg, err := config.Load(data)
	if err != nil {
		logger.Fatal("failed to load config", "error", err)
	}

	logger.Init(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	})

	routes := make(map[string]proxy.RouteEntry)
	routesForEngine := make([]engine.RouteConfig, 0, len(cfg.Routing))
	for _, r := range cfg.Routing {
		upstreamURL, err := url.Parse(r.Upstream)
		if err != nil {
			logger.Fatal("invalid upstream URL", "url", r.Upstream, "error", err)
		}
		routes[r.Prefix] = proxy.RouteEntry{
			Upstream:  upstreamURL,
			Ecosystem: r.Ecosystem,
		}
		routesForEngine = append(routesForEngine, engine.RouteConfig{
			Prefix:    r.Prefix,
			Upstream:  r.Upstream,
			Ecosystem: r.Ecosystem,
		})
	}

	var malDB *sql.DB
	if cfg.DBPath != "" {
		malDB, err = schema.OpenReadOnly(cfg.DBPath)
		if err != nil {
			logger.Warn("failed to open malicious database", "error", err)
			malDB = nil
		} else {
			defer malDB.Close()
		}
	}

	checks := []engine.CheckFunc{}
	checks = append(checks, engine.MaliciousDBCheck(malDB))

	if cfg.Cooldown.Enabled {
		metadataFetcher := engine.NewRegistryMetadataFetcher(routesForEngine)
		checks = append(checks, engine.CooldownCheck(metadataFetcher, cfg.Cooldown.Duration))
	}

	if cfg.Typo.Enabled {
		checks = append(checks, engine.TypoCheck(nil, cfg.Typo.Threshold))
	}

	if cfg.OSV.Enabled {
		osvClient := osv.NewClient(cfg.OSV.BaseURL, cfg.OSV.TTL)
		checks = append(checks, engine.OSVCheck(osvClient))
	}

	eng := engine.New(checks...)
	handler := proxy.New(eng, routes)

	if cfg.ScriptScan.Enabled {
		sc := &scanner.StreamChecker{
			Config: scanner.ScanConfig{
				MaxFileSize:        cfg.ScriptScan.MaxFileSize,
				MaxTotalSize:       cfg.ScriptScan.MaxTotalSize,
				EntropyEnabled:     cfg.ScriptScan.Entropy.Enabled,
				EntropyThreshold:   cfg.ScriptScan.Entropy.Threshold,
				ObfuscationEnabled: cfg.ScriptScan.Obfuscation.Enabled,
				ObfuscationMinB64:  cfg.ScriptScan.Obfuscation.Base64MinLength,
				ObfuscationMinHex:  cfg.ScriptScan.Obfuscation.HexMinLength,
				NetworkEnabled:     cfg.ScriptScan.Network.Enabled,
				AllowPrivateIPs:    cfg.ScriptScan.Network.AllowPrivateIPs,
			},
		}
		handler.SetStreamChecker(sc)
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("server shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	logger.Info("malFuse listening", "addr", addr)
	for prefix := range routes {
		logger.Info("route configured", "prefix", prefix, "upstream", "upstream")
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("server error", "error", err)
	}
	logger.Info("server stopped")
}
