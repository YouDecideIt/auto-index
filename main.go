package main

import (
	sysctx "context"
	"encoding/json"
	"flag"
	"github.com/YouDecideIt/auto-index/config"
	"github.com/YouDecideIt/auto-index/context"
	"github.com/YouDecideIt/auto-index/experiment"
	"github.com/YouDecideIt/auto-index/operations"
	"github.com/YouDecideIt/auto-index/request"
	"github.com/YouDecideIt/auto-index/study"
	"github.com/YouDecideIt/auto-index/utils"
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

func Process(ctx context.Context) {
	now := time.Now()
	log.Info("processing")
	defer func() {
		log.Info("process done", zap.Duration("in", time.Since(now)))
	}()

	items, err := study.Study(ctx.Cfg.NgMonitorConfig.Address)
	if err != nil {
		log.Error("failed to study", zap.Error(err))
		return
	}

	for i, item := range items {
		log.Info("study", zap.Int("index", i))
		if item.SQLText == "" {
			continue
		}
		indexes, estRatio, err := request.WhatIf(ctx, item.SQLText)
		if err != nil {
			log.Error("failed to request what if", zap.Error(err))
			return
		}
		if estRatio < ctx.Cfg.EvaluateConfig.EstRatioThreshold {
			log.Info("optimization estRatio is lower than threshold, skip",
				zap.Float64("estRatio", estRatio),
				zap.Float64("threshold", ctx.Cfg.EvaluateConfig.EstRatioThreshold))
			continue
		}

		// Start a B instance
		cluster := operations.New()
		err = cluster.StartBCluster(sysctx.Background())
		if err != nil {
			return
		}
		endpoint, err := cluster.WaitBClusterStartedAndMirrored(sysctx.Background())
		if err != nil {
			log.Error("failed to wait for B cluster", zap.Error(err))
		}
		defer cluster.DestroyBCluster(sysctx.Background())

		// experiment
		actRatio, err := experiment.Experiment(ctx, endpoint, item, indexes)
		if err != nil {
			log.Error("failed to experiment", zap.Error(err))
			return
		}
		if actRatio < ctx.Cfg.EvaluateConfig.ActRatioThreshold {
			log.Info("optimization actRatio is lower than threshold, skip",
				zap.Float64("astRatio", actRatio),
				zap.Float64("threshold", ctx.Cfg.EvaluateConfig.ActRatioThreshold))
			continue
		}

		// ApplyIndex(ctx,)
		err = request.ApplyIndex(ctx.DB, indexes)
		if err != nil {
			log.Error("failed to apply index", zap.Error(err))
			return
		}
	}
}

func main() {
	flag.Parse()

	ctx := context.Context{}
	{
		loadConfig, err := config.LoadConfig(*cfgFilePath, overrideConfig)
		if err != nil {
			// logger isn't initialized, need to use stdlog
			stdlog.Fatalf("failed to load config file, config.file: %s", *cfgFilePath)
		}
		ctx.Cfg = loadConfig
	}

	err := initLogger(ctx.Cfg)
	if err != nil {
		// failed to initialize logger, need to use stdlog
		stdlog.Fatalf("failed to init logger, err: %s", err.Error())
	}

	printer.PrintAutoIndexInfo()

	str, err := json.Marshal(ctx.Cfg)
	log.Info("config", zap.String("config", string(str)))

	//if len(context.Cfg.WebConfig.Address) == 0 {
	//	log.Fatal("empty listen address", zap.String("listen-address", context.Cfg.WebConfig.Address))
	//}

	ctx.DB = utils.OpenDatabase(ctx.Cfg.TiDBConfig.Address)
	defer utils.CloseDatabase(ctx.DB)

	//service.Init(AutoIndexConfig, storage)
	//defer service.Stop()

	//scrape.Init(AutoIndexConfig, storage)
	//defer scrape.Stop()

	//ticker := time.NewTicker(ctx.Cfg.EvaluateConfig.Interval)
	go func() {
		Process(ctx)
		//for _ = range ticker.C {
		//	Process(ctx)
		//}
	}()

	sig := waitForSigterm()
	log.Info("received signal", zap.String("sig", sig.String()))
}
