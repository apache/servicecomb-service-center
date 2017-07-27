package common

import (
	"github.com/astaxie/beego"
	"github.com/ServiceComb/service-center/pkg/common/logrotate"
	"github.com/ServiceComb/service-center/pkg/lager"
	"github.com/ServiceComb/service-center/util"
	"os"
	"path/filepath"
	"time"
)

func init() {
	initLogger()
	loadServerSSLConfig()
	loadClientSSLConfig()
	initLogRotate()
}

func initLogger() {
	logFormatText, err := beego.AppConfig.Bool("LogFormatText")
	loggerFile := os.ExpandEnv(beego.AppConfig.String("logfile"))
	loggerName := beego.AppConfig.String("ComponentName")
	enableRsyslog, err := beego.AppConfig.Bool("EnableRsyslog")
	if err != nil {
		enableRsyslog = false
	}

	enableStdOut := beego.AppConfig.DefaultString("runmode", "prod") == "dev"
	util.InitLogger(loggerName, &lager.Config{
		LoggerLevel:   beego.AppConfig.String("loglevel"),
		LoggerFile:    loggerFile,
		EnableRsyslog: enableRsyslog,
		LogFormatText: logFormatText,
		EnableStdOut:  enableStdOut,
	})
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
