package v1

import (
	"fmt"
	"time"

	guuid "github.com/gofrs/uuid"
)

func (x *Event) Flag() string {
	return fmt.Sprintf("id: %s,action: %s, subject: %s", x.Id, x.Action, x.Subject)
}

func (x *EventList) Flag() string {
	return fmt.Sprintf("event count %d", len(x.Events))
}

func NewEventID() (string, error) {
	uuid, err := guuid.NewV4()
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}

func Timestamp() int64 {
	return time.Now().UnixNano()
}
