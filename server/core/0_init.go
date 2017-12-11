package core

import (
	"flag"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/grace"
	"github.com/ServiceComb/service-center/pkg/lager"
	"github.com/ServiceComb/service-center/pkg/logrotate"
	"github.com/ServiceComb/service-center/pkg/plugin"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/version"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func init() {
	Initialize()
}

func Initialize() {
	initCommandLine()

	plugin.SetPluginDir(ServerInfo.Config.PluginsDir)

	initLogger()

	printVersion()

	go handleSignals()

	grace.Init()
}

func initCommandLine() {
	var printVer bool
	flag.BoolVar(&printVer, "v", false, "Print the version and exit.")
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.Parse(os.Args[1:])

	if printVer {
		fmt.Printf("ServiceCenter version: %s\n", version.Ver().Version)
		fmt.Printf("Build tag: %s\n", version.Ver().BuildTag)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

func printVersion() {
	util.Logger().Infof("service center version: %s", version.Ver().Version)
	util.Logger().Infof("Build tag: %s", version.Ver().BuildTag)
	util.Logger().Infof("Go version: %s", runtime.Version())
	util.Logger().Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)

	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	util.Logger().Infof("service center is running simultaneously with %d CPU cores", cores)
}

func initLogger() {
	util.InitLogger(ServerInfo.Config.LoggerName,
		&lager.Config{
			LoggerLevel:   ServerInfo.Config.LogLevel,
			LoggerFile:    os.ExpandEnv(ServerInfo.Config.LogFilePath),
			EnableRsyslog: ServerInfo.Config.LogSys,
			LogFormatText: ServerInfo.Config.LogFormat == "text",
			EnableStdOut:  version.Ver().RunMode == "dev",
		})

	// custom loggers
	util.CustomLogger("Heartbeat", "heartbeat")
	util.CustomLogger("HeartbeatSet", "heartbeat")

	util.CustomLogger("github.com/ServiceComb/service-center/server/service/event", "event")
	util.CustomLogger("github.com/ServiceComb/service-center/server/service/notification", "event")

	util.CustomLogger("github.com/ServiceComb/service-center/server/core/backend", "registry")

	initLogRotate()
}

func initLogRotate() {
	rotatePeriod := 30 * time.Second
	traceutils.RunLogRotate(&traceutils.LogRotateConfig{
		Dir:         filepath.Dir(os.ExpandEnv(ServerInfo.Config.LogFilePath)),
		MaxFileSize: int(ServerInfo.Config.LogRotateSize),
		BackupCount: int(ServerInfo.Config.LogBackupCount),
		Period:      rotatePeriod,
	})
}

func handleSignals() {
	var sig os.Signal
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
	)
	wait := 5 * time.Second
	for {
		sig = <-sigCh
		switch sig {
		case syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM:
			<-time.After(wait)
			util.Logger().Warnf(nil, "Waiting for server response timed out(%s), force shutdown.", wait)
			os.Exit(1)
		}
	}
}
