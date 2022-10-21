package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/YouDecideIt/auto-index/config"
	"github.com/YouDecideIt/auto-index/utils/printer"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pingcap/log"
	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
)

const (
	nmConfigFilePath = "config.file"
	nmAddr           = "address"
	nmTiDBAddr       = "tidb.address"
	nmLogLevel       = "log.level"
	nmLogFile        = "log.file"
	nmCleanup        = "cleanup"
)

var (
	cfgFilePath = flag.String(nmConfigFilePath, "", "YAML config file path for autoIndex.")
	cleanup     = flag.Bool(nmCleanup, false, "Whether to cleanup data during shutting down, set for debug")
	tidbAddr    = flag.String(nmTiDBAddr, config.DefaultAutoIndexConfig.TiDBConfig.Address, "The address of TiDB")
	listenAddr  = flag.String(nmAddr, config.DefaultAutoIndexConfig.WebConfig.Address, "TCP address to listen for http connections")
	logLevel    = flag.String(nmLogLevel, config.DefaultAutoIndexConfig.LogConfig.LogLevel, "Log level")
	logFile     = flag.String(nmLogFile, config.DefaultAutoIndexConfig.LogConfig.LogFile, "Log file")
)

func overrideConfig(config *config.AutoIndexConfig) {
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case nmAddr:
			config.WebConfig.Address = *listenAddr
		case nmTiDBAddr:
			config.TiDBConfig.Address = *tidbAddr
		case nmLogFile:
			config.LogConfig.LogFile = *logFile
		case nmLogLevel:
			config.LogConfig.LogLevel = *logLevel
		}
	})
}

func initLogger(cfg *config.AutoIndexConfig) error {
	logCfg := &log.Config{
		Level: cfg.LogConfig.LogLevel,
		File:  log.FileLogConfig{Filename: cfg.LogConfig.LogFile},
	}

	logger, p, err := log.InitLogger(logCfg)
	if err != nil {
		return err
	}
	log.ReplaceGlobals(logger, p)
	return nil
}

func initDatabase(cfg *config.AutoIndexConfig) *sql.DB {
	now := time.Now()
	log.Info("setting up database")
	defer func() {
		log.Info("init database done", zap.Duration("in", time.Since(now)))
	}()

	db, err := sql.Open("mysql", fmt.Sprintf("root@(%s)/test", cfg.TiDBConfig.Address))
	if err != nil {
		log.Fatal("failed to open db", zap.Error(err))
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}

func closeDatabase(db *sql.DB) {
	now := time.Now()
	log.Info("closing database")
	defer func() {
		log.Info("close database done", zap.Duration("in", time.Since(now)))
	}()

	if err := db.Close(); err != nil {
		log.Warn("failed to close database", zap.Error(err))
	}
}

func waitForSigterm() os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	for {
		sig := <-ch
		if sig == syscall.SIGHUP {
			// Prevent from the program stop on SIGHUP
			continue
		}
		return sig
	}
}

func main() {
	flag.Parse()

	autoIndexConfig, err := config.LoadConfig(*cfgFilePath, overrideConfig)
	if err != nil {
		// logger isn't initialized, need to use stdlog
		stdlog.Fatalf("failed to load config file, config.file: %s", *cfgFilePath)
	}
	err = initLogger(autoIndexConfig)
	if err != nil {
		// failed to initialize logger, need to use stdlog
		stdlog.Fatalf("failed to init logger, err: %s", err.Error())
	}

	printer.PrintAutoIndexInfo()

	if len(autoIndexConfig.WebConfig.Address) == 0 {
		log.Fatal("empty listen address", zap.String("listen-address", autoIndexConfig.WebConfig.Address))
	}

	db := initDatabase(autoIndexConfig)
	defer closeDatabase(db)

	//storage := store.NewDefaultMetricStorage(db)
	//defer storage.Close()

	//service.Init(autoIndexConfig, storage)
	//defer service.Stop()

	//scrape.Init(autoIndexConfig, storage)
	//defer scrape.Stop()

	sig := waitForSigterm()
	log.Info("received signal", zap.String("sig", sig.String()))
}
