package core

import (
	"flag"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/grace"
	"github.com/ServiceComb/service-center/pkg/lager"
	"github.com/ServiceComb/service-center/pkg/logrotate"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/version"
	"github.com/astaxie/beego"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

var printVer bool

func init() {
	Initialize()
}

func Initialize() {
	initCommandLine()

	initLogger()

	printVersion()

	go handleSignals()

	initLogRotate()

	grace.Init()
}

func initCommandLine() {
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
	logFormatText, err := beego.AppConfig.Bool("LogFormatText")
	loggerFile := os.ExpandEnv(beego.AppConfig.String("logfile"))
	loggerName := beego.AppConfig.String("ComponentName")
	enableRsyslog, err := beego.AppConfig.Bool("EnableRsyslog")
	if err != nil {
		enableRsyslog = false
	}

	util.InitLogger(loggerName, &lager.Config{
		LoggerLevel:   beego.AppConfig.String("loglevel"),
		LoggerFile:    loggerFile,
		EnableRsyslog: enableRsyslog,
		LogFormatText: logFormatText,
		EnableStdOut:  version.Ver().RunMode == "dev",
	})

	// custom loggers
	util.CustomLogger("Heartbeat", "heartbeat")
	util.CustomLogger("HeartbeatSet", "heartbeat")

	util.CustomLogger("github.com/ServiceComb/service-center/server/service/event", "event")
	util.CustomLogger("github.com/ServiceComb/service-center/server/service/notification", "event")

	util.CustomLogger("github.com/ServiceComb/service-center/server/core/backend", "registry")
}

func initLogRotate() {
	logDir := os.ExpandEnv(beego.AppConfig.String("logfile"))
	rotatePeriod := 30 * time.Second
	maxFileSize := beego.AppConfig.DefaultInt("log_rotate_size", 20)
	if maxFileSize <= 0 || maxFileSize > 50 {
		maxFileSize = 20
	}
	maxBackupCount := beego.AppConfig.DefaultInt("log_backup_count", 5)
	if maxBackupCount < 0 || maxBackupCount > 100 {
		maxBackupCount = 5
	}
	traceutils.RunLogRotate(&traceutils.LogRotateConfig{
		Dir:         filepath.Dir(logDir),
		MaxFileSize: maxFileSize,
		BackupCount: maxBackupCount,
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
			util.Logger().Warnf(nil, "Clean up resources timed out(%s), force shutdown.", wait)
			os.Exit(1)
		}
	}
}
