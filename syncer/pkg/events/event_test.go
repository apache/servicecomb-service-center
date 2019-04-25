package events

import (
	"context"
	"testing"
)

func TestNewContextEvent(t *testing.T) {
	event := NewContextEvent("test", context.Background())
	event.Type()
	event.Context()
}
