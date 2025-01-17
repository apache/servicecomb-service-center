package protect

import (
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

// ProtectionChecker checks whether to do the protection on null instance
type ProtectionChecker interface {
	CheckProtection() bool
}

// DelayedStopProtectChecker means start the protection and close it after some time
type DelayedStopProtectChecker struct {
	delay       time.Duration
	start       time.Time
	logWhenStop sync.Once
}

func NewDelayedSuccessChecker(delay time.Duration) *DelayedStopProtectChecker {
	return &DelayedStopProtectChecker{
		delay:       delay,
		start:       time.Now(),
		logWhenStop: sync.Once{},
	}
}

func (d *DelayedStopProtectChecker) CheckProtection() bool {
	if time.Since(d.start) > d.delay {
		d.logWhenStop.Do(func() {
			log.Info("null instance protection stopped")
		})
		return false
	}
	return true
}

func GetGlobalProtectionChecker() ProtectionChecker {
	return globalProtectionChecker
}
