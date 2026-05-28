package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"malFuse/internal/config"
	"malFuse/internal/daemon"
	"malFuse/internal/db/schema"
	"malFuse/internal/engine"
	"malFuse/internal/logger"
	"malFuse/internal/osv"
	"malFuse/internal/proxy"
	"malFuse/internal/scanner"
)

var configPath string

func main() {
	rootCmd := &cobra.Command{Use: "malfuse", Short: "malFuse — malicious package firewall"}
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.json", "config file path")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start proxy (daemon mode)",
		RunE:  startCmd,
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop running daemon",
		RunE:  stopCmd,
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		RunE:  statusCmd,
	})

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runProxy()
	}

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("fatal", "error", err)
	}
}

func startCmd(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if os.Getenv("MALFUSE_DAEMON") == "1" {
		if err := daemon.WritePID(cfg.PIDFile, os.Getpid()); err != nil {
			return fmt.Errorf("write pid: %w", err)
		}
		defer daemon.RemovePID(cfg.PIDFile)
		return runProxyWithConfig(cfg)
	}

	exe, _ := os.Executable()
	proc, err := os.StartProcess(exe, append(os.Args, "-c", configPath), &os.ProcAttr{
		Env: append(os.Environ(), "MALFUSE_DAEMON=1"),
	})
	if err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}
	fmt.Printf("Daemon started, PID: %d\n", proc.Pid)
	return nil
}

func stopCmd(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pid, err := daemon.ReadPID(cfg.PIDFile)
	if err != nil {
		return fmt.Errorf("not running: %w", err)
	}

	fmt.Printf("Stopping PID %d...\n", pid)
	if err := daemon.SendSignal(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM: %w", err)
	}

	time.Sleep(5 * time.Second)
	if daemon.IsRunning(pid) {
		fmt.Println("Force killing...")
		daemon.SendSignal(pid, syscall.SIGKILL)
	}
	daemon.RemovePID(cfg.PIDFile)
	fmt.Println("Stopped")
	return nil
}

func statusCmd(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pid, err := daemon.ReadPID(cfg.PIDFile)
	if err != nil {
		fmt.Println("not running")
		return nil
	}
	if daemon.IsRunning(pid) {
		fmt.Printf("running, PID: %d\n", pid)
	} else {
		fmt.Println("not running (stale pid file)")
	}
	return nil
}

func runProxy() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	logger.Init(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	})
	return runProxyWithConfig(cfg)
}

func loadConfig() (*config.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return config.Load(data)
}

func runProxyWithConfig(cfg *config.Config) error {
	routes := make(map[string]proxy.RouteEntry)
	routesForEngine := make([]engine.RouteConfig, 0, len(cfg.Routing))
	for _, r := range cfg.Routing {
		upstreamURL, err := url.Parse(r.Upstream)
		if err != nil {
			return fmt.Errorf("invalid upstream %s: %w", r.Upstream, err)
		}
		routes[r.Prefix] = proxy.RouteEntry{Upstream: upstreamURL, Ecosystem: r.Ecosystem}
		routesForEngine = append(routesForEngine, engine.RouteConfig{
			Prefix: r.Prefix, Upstream: r.Upstream, Ecosystem: r.Ecosystem,
		})
	}

	var malDB *sql.DB
	var err error
	if cfg.DBPath != "" {
		malDB, err = schema.OpenReadOnly(cfg.DBPath)
		if err != nil {
			logger.Warn("failed to open malware database", "error", err)
			malDB = nil
		} else {
			defer malDB.Close()
		}
	}

	checks := []engine.CheckFunc{engine.MaliciousDBCheck(malDB)}

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

	if malDB != nil {
		handler.SetDBPinger(malDB)
	}

	if cfg.ScriptScan.Enabled {
		handler.SetStreamChecker(&scanner.StreamChecker{
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
		})
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{Addr: addr, Handler: handler}

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
		logger.Info("route", "prefix", prefix)
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
