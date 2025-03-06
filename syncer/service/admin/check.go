package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chassis/foundation/gopool"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/rpc"
)

func ShouldTrustPeerServer() bool {
	return globalHealthChecker.ShouldTrustPeerServer()
}

type HealthChecker struct {
	checkIntervalBySecond uint
	latestCheckResult     *Resp
	latestCheckErr        error

	// 为了容忍网络抖动，使用滑动窗口判断状态
	failureWindow *HealthCheckWindow
	// 恢复窗口，成功率要求更高，用于数据同步的恢复
	recoveryWindow        *HealthCheckWindow
	shouldTrustPeerServer bool
}

func (h *HealthChecker) check() {
	h.latestCheckResult, h.latestCheckErr = checkPeerStatus()
	if h.latestCheckErr != nil || h.latestCheckResult == nil || len(h.latestCheckResult.Peers) == 0 {
		h.AddResult(false)
		return
	}
	if h.latestCheckResult.Peers[0].Status != rpc.HealthStatusConnected {
		h.AddResult(false)
		return
	}
	h.AddResult(true)
}

func (h *HealthChecker) ShouldTrustPeerServer() bool {
	return h.shouldTrustPeerServer
}

func (h *HealthChecker) AddResult(pass bool) {
	h.failureWindow.AddResult(pass)
	h.recoveryWindow.AddResult(pass)

	shouldTrustPeerServerNew := true
	if h.shouldTrustPeerServer {
		// 健康 > 不健康
		shouldTrustPeerServerNew = h.failureWindow.IsHealthy()
	} else {
		// 不健康 > 健康
		shouldTrustPeerServerNew = h.recoveryWindow.IsHealthy()
	}
	if h.shouldTrustPeerServer != shouldTrustPeerServerNew {
		log.Info(fmt.Sprintf("should trust peer server changed, old: %v, new: %v", h.shouldTrustPeerServer, shouldTrustPeerServerNew))
		h.shouldTrustPeerServer = shouldTrustPeerServerNew
		return
	}
}

func (h *HealthChecker) LatestHealthCheckResult() (*Resp, error) {
	return h.latestCheckResult, h.latestCheckErr
}

func (h *HealthChecker) RunChecker() {
	h.check()
	gopool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(h.checkIntervalBySecond) * time.Second):
				h.check()
			}
		}
	})
}
