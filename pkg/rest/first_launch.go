package rest

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

/**
for restart service center, set a 2-minute time window to return http status 304 on discovery apis,
indicating that sdk not need to clear cache
*/

var (
	isWithinProtection        bool
	startupTimestamp          int64
	firstLaunchFlagPath       string
	enableInstanceNullProtect bool
	restartProtectInterval    time.Duration
	RestartProtectHttpCode    int
)

func Init() {
	enableInstanceNullProtect = config.GetBool("instance_null_protect.enable", true)
	restartProtectInterval = time.Duration(config.GetInt("instance_null_protect.restart_protect_interval", 120))
	RestartProtectHttpCode = config.GetInt("instance_null_protect.http_status", 304)
	firstLaunchFlagPath = filepath.Join(util.GetAppRoot(), "first_launch.flag")

	_, err := os.Stat(firstLaunchFlagPath)
	// first launch, need not instance null protection
	if errors.Is(err, fs.ErrNotExist) {
		file, err := os.Create(firstLaunchFlagPath)
		if err != nil {
			log.Error(firstLaunchFlagPath, errors.New("failed to create file"))
			os.Exit(1)
		}
		file.Close()
	} else if err != nil {
		log.Info("failed to stat flag file")
		os.Exit(1)
	}
	// file exist, not first launch, set protection time window of restartProtectInterval
	startupTimestamp = time.Now().UnixNano()
	isWithinProtection = true
}

func IsWithinRestartProtection() bool {
	if !enableInstanceNullProtect {
		return false
	}

	if !isWithinProtection {
		return false
	}

	if time.Now().Add(-restartProtectInterval).UnixNano() > startupTimestamp {
		log.Info("restart protection stop")
		isWithinProtection = false
		return false
	}
	log.Info("within restart protection")
	return true
}
