package admin

type HealthCheckWindow struct {
	windowSize       int
	checkPassResults []bool
	failureThreshold int
}

func NewHealthCheckWindow(windowSize int, failureThreshold int) *HealthCheckWindow {
	return &HealthCheckWindow{
		windowSize:       windowSize,
		checkPassResults: make([]bool, 0, windowSize),
		failureThreshold: failureThreshold,
	}
}

func (h *HealthCheckWindow) AddResult(pass bool) {
	h.checkPassResults = append(h.checkPassResults, pass)
	if len(h.checkPassResults) > h.windowSize {
		h.checkPassResults = h.checkPassResults[1:]
	}
}

func (h *HealthCheckWindow) IsHealthy() bool {
	failureCount := 0
	for _, pass := range h.checkPassResults {
		if !pass {
			failureCount++
		}
	}
	return failureCount < h.failureThreshold
}
