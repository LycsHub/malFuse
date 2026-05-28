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
	"malFuse/internal/linker"
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

	linkCmd := &cobra.Command{Use: "link", Short: "Configure pip/npm to use malFuse proxy", RunE: linkCmdFn}
	linkCmd.Flags().String("target", "", "package manager: pip, npm, or empty for both")
	rootCmd.AddCommand(linkCmd)

	unlinkCmd := &cobra.Command{Use: "unlink", Short: "Restore original pip/npm configuration", RunE: unlinkCmdFn}
	unlinkCmd.Flags().String("target", "", "package manager: pip, npm, or empty for both")
	rootCmd.AddCommand(unlinkCmd)

	allowCmd := &cobra.Command{Use: "allow", Short: "Manage whitelist (allow package through firewall)"}
	allowAdd := &cobra.Command{Use: "add <name>", Short: "Add package to whitelist", Args: cobra.ExactArgs(1), RunE: allowAddCmd}
	allowAdd.Flags().String("ecosystem", "pypi", "ecosystem: pypi or npm")
	allowAdd.Flags().String("version", "", "version constraint (empty = all versions)")
	allowCmd.AddCommand(allowAdd)

	allowRemove := &cobra.Command{Use: "remove <name>", Short: "Remove package from whitelist", Args: cobra.ExactArgs(1), RunE: allowRemoveCmd}
	allowRemove.Flags().String("ecosystem", "pypi", "ecosystem: pypi or npm")
	allowRemove.Flags().String("version", "", "version constraint")
	allowCmd.AddCommand(allowRemove)

	allowList := &cobra.Command{Use: "list", Short: "List whitelisted packages", RunE: allowListCmd}
	allowList.Flags().String("ecosystem", "", "filter by ecosystem (empty = all)")
	allowCmd.AddCommand(allowList)

	rootCmd.AddCommand(allowCmd)

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

	eng := engine.New()
	eng.AddNamed("whitelist", engine.WhitelistCheck(malDB))
	eng.AddNamed("malicious-db", engine.MaliciousDBCheck(malDB))

	if cfg.Cooldown.Enabled {
		metadataFetcher := engine.NewRegistryMetadataFetcher(routesForEngine)
		eng.AddNamed("cooldown", engine.CooldownCheck(metadataFetcher, cfg.Cooldown.Duration))
	}
	if cfg.Typo.Enabled {
		eng.AddNamed("typo-squatting", engine.TypoCheck(nil, cfg.Typo.Threshold))
	}
	if cfg.OSV.Enabled {
		osvClient := osv.NewClient(cfg.OSV.BaseURL, cfg.OSV.TTL)
		eng.AddNamed("osv-api", engine.OSVCheck(osvClient, cfg.OSV.BlockOnVuln))
	}

	handler := proxy.New(eng, routes)

	if malDB != nil {
		handler.SetDBPinger(malDB)
		handler.SetDBFilter(malDB)
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

func linkCmdFn(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	target, _ := cmd.Flags().GetString("target")
	return linker.Link(linker.LinkConfig{
		ProxyHost:    cfg.Host,
		ProxyPort:    cfg.Port,
		PypiUpstream: "https://pypi.org/simple/",
		NpmUpstream:  "https://registry.npmjs.org/",
	}, target)
}

func unlinkCmdFn(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")
	return linker.Unlink(target)
}

func allowAddCmd(cmd *cobra.Command, args []string) error {
	cfg, _ := loadConfig()
	db, err := schema.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	eco, _ := cmd.Flags().GetString("ecosystem")
	ver, _ := cmd.Flags().GetString("version")
	name := args[0]

	if err := schema.InsertWhitelist(db, name, eco, ver); err != nil {
		return fmt.Errorf("add whitelist: %w", err)
	}
	fmt.Printf("Added %s (%s) to whitelist\n", name, eco)
	return nil
}

func allowRemoveCmd(cmd *cobra.Command, args []string) error {
	cfg, _ := loadConfig()
	db, err := schema.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	eco, _ := cmd.Flags().GetString("ecosystem")
	ver, _ := cmd.Flags().GetString("version")
	name := args[0]

	if err := schema.DeleteWhitelist(db, name, eco, ver); err != nil {
		return fmt.Errorf("remove whitelist: %w", err)
	}
	fmt.Printf("Removed %s (%s) from whitelist\n", name, eco)
	return nil
}

func allowListCmd(cmd *cobra.Command, args []string) error {
	cfg, _ := loadConfig()
	db, err := schema.OpenReadOnly(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	eco, _ := cmd.Flags().GetString("ecosystem")
	query := "SELECT name, COALESCE(version, '(all)'), ecosystem FROM whitelist"
	args2 := []interface{}{}
	if eco != "" {
		query += " WHERE ecosystem=?"
		args2 = append(args2, eco)
	}
	query += " ORDER BY ecosystem, name"

	rows, err := db.Query(query, args2...)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, version, ecosystem string
		rows.Scan(&name, &version, &ecosystem)
		fmt.Printf("  %s@%s [%s]\n", name, version, ecosystem)
	}
	return nil
}
