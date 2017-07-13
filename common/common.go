package common

import (
	"github.com/astaxie/beego"
	"github.com/servicecomb/service-center/common/logrotate"
	"github.com/servicecomb/service-center/lager"
	"github.com/servicecomb/service-center/util"
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

	util.InitLogger(loggerName, &lager.Config{
		LoggerLevel:   beego.AppConfig.String("loglevel"),
		LoggerFile:    loggerFile,
		EnableRsyslog: enableRsyslog,
		LogFormatText: logFormatText,
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
