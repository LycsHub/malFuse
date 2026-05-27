package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"malFuse/internal/config"
	"malFuse/internal/db/schema"
	"malFuse/internal/engine"
	"malFuse/internal/osv"
	"malFuse/internal/proxy"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	cfg, err := config.Load(data)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	routes := make(map[string]proxy.RouteEntry)
	routesForEngine := make([]engine.RouteConfig, 0, len(cfg.Routing))
	for _, r := range cfg.Routing {
		upstreamURL, err := url.Parse(r.Upstream)
		if err != nil {
			log.Fatalf("Invalid upstream URL %s: %v", r.Upstream, err)
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
			log.Printf("[WARN] Failed to open malicious database: %v", err)
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

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("malFuse listening on %s", addr)
	for prefix := range routes {
		log.Printf("  %s -> %s", prefix, "upstream")
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped")
}
