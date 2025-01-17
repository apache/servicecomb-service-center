package protect

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chassis/foundation/gopool"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/service/registry"
)

/**
for restart service center, set a restartProtectInterval time window to return RestartProtectHttpCode on discovery apis,
indicating that sdk not need to clear cache
*/

var (
	enableInstanceNullProtect bool
	restartProtectInterval    time.Duration
	RestartProtectHttpCode    int
	validProtectCode          = map[int]struct{}{http.StatusNotModified: {}, http.StatusUnprocessableEntity: {}, http.StatusInternalServerError: {}}
	globalProtectionChecker   ProtectionChecker
)

const (
	maxInterval                   = 60 * 60 * 24 * time.Second
	minInterval                   = 0 * time.Second
	defaultRestartProtectInterval = 120 * time.Second
	defaultReadinessCheckInterval = 5 * time.Second // the null instance protection should start again when server is unready
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

	StartProtectionAndStopDelayed()
	gopool.Go(watch)

}

func watch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadinessCheckInterval):
			err := registry.Readiness(ctx)
			if err != nil {
				AlwaysProtection()
				continue
			}
			StartProtectionAndStopDelayed()
		}
	}
}

func ShouldProtectOnNullInstance() bool {
	if !enableInstanceNullProtect {
		return false
	}

	if globalProtectionChecker == nil { // protect by default
		return true
	}

	return GetGlobalProtectionChecker().CheckProtection()
}

func StartProtectionAndStopDelayed() {
	if _, ok := globalProtectionChecker.(*DelayedStopProtectChecker); ok {
		return
	}
	globalProtectionChecker = NewDelayedSuccessChecker(restartProtectInterval)
	log.Info("start protection and stop delayed on null instance")
}

func AlwaysProtection() {
	if globalProtectionChecker == nil {
		return
	}
	globalProtectionChecker = nil
	log.Info("always protect on null instance")
}
