package protect

import (
	"fmt"
	"net/http"
	"time"

	"github.com/apache/servicecomb-service-center/server/config"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

/**
for restart service center, set a restartProtectInterval time window to return RestartProtectHttpCode on discovery apis,
indicating that sdk not need to clear cache
*/

var (
	isWithinProtection        bool
	startupTimestamp          int64
	enableInstanceNullProtect bool
	restartProtectInterval    time.Duration
	RestartProtectHttpCode    int
	validProtectCode          = map[int]struct{}{http.StatusNotModified: {}, http.StatusUnprocessableEntity: {}, http.StatusInternalServerError: {}}
)

const (
	maxInterval                   = 120 * time.Second
	minInterval                   = 0 * time.Second
	defaultRestartProtectInterval = 120 * time.Second
)

func Init() {
	enableInstanceNullProtect = config.GetBool("instance_null_protect.enable", false)
	if !enableInstanceNullProtect {
		return
	}
	restartProtectInterval = time.Duration(config.GetInt("instance_null_protect.restart_protect_interval", 120)) * time.Second
	if restartProtectInterval > maxInterval || restartProtectInterval < minInterval {
		log.Warn(fmt.Sprintf("invalid instance_null_protect.restart_protect_interval: %d,"+
			" must between %d-%ds inclusively", restartProtectInterval, minInterval, maxInterval))
		restartProtectInterval = defaultRestartProtectInterval
	}
	RestartProtectHttpCode = config.GetInt("instance_null_protect.http_status", http.StatusNotModified)
	if _, ok := validProtectCode[RestartProtectHttpCode]; !ok {
		log.Warn(fmt.Sprintf("invalid instance_null_protect.http_status: %d, must be %v", RestartProtectHttpCode, validProtectCode))
		RestartProtectHttpCode = http.StatusNotModified
	}

	log.Info(fmt.Sprintf("instance_null_protect.enable: %t", enableInstanceNullProtect))
	log.Info(fmt.Sprintf("instance_null_protect.restart_protect_interval: %d", restartProtectInterval))
	log.Info(fmt.Sprintf("instance_null_protect.http_status: %d", RestartProtectHttpCode))
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
