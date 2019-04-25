package events

import (
	"context"
	"testing"
)

type mockListener struct {}

func (m mockListener) OnEvent(event ContextEvent)  {}

func TestListener(t *testing.T) {
	eventType := "mocktest"
	event := NewContextEvent(eventType, context.Background())
	Dispatch(event)

	ml := &mockListener{}
	RemoveListener(eventType, ml)
	AddListener(eventType, ml)


	ml2 := &mockListener{}
	AddListener(eventType, ml2)
	Dispatch(event)

	RemoveListener(eventType, ml)
	RemoveListener(eventType, ml2)

	Clean()
}
