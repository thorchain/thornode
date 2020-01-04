package types

import (
	"encoding/json"
	"errors"
	"strings"
)

// EventStatus use in event
type EventStatus uint8

// EventStatuses a list of event status
type EventStatuses []EventStatus

const (
	Success EventStatus = iota
	Failed
	Pending
	Refund
)

var eventStatusStr = map[string]EventStatus{
	"Success": Success,
	"Failed":  Failed,
	"Pending": Pending,
	"Refund":  Refund,
}

// String implement fmt.Stringer, convert from EventStatus to string
func (es EventStatus) String() string {
	for k, v := range eventStatusStr {
		if v == es {
			return k
		}
	}
	return ""
}

// Valid is to check whether the EventStatus is valid
func (es EventStatus) Valid() error {
	if es.String() == "" {
		return errors.New("invalid EventStatus")
	}
	return nil
}

// MarshalJSON marshal EventStatus to json
func (es EventStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(es.String())
}

// UnmarshalJSON deserialize from json
func (es *EventStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); nil != err {
		return err
	}
	*es = GetEventStatus(s)
	return nil
}

// GetEventStatus from string
func GetEventStatus(es string) EventStatus {
	for key, item := range eventStatusStr {
		if strings.EqualFold(key, es) {
			return item
		}
	}

	return Pending
}

// GetEventStatuses convert a list of status in string type to EventStatus
func GetEventStatuses(es []string) EventStatuses {
	var result EventStatuses
	for _, item := range es {
		if len(item) == 0 {
			continue
		}
		result = append(result, GetEventStatus(item))
	}
	return result
}

// Contains check whether the given event status is in the list
func (es EventStatuses) Contains(eventStatus EventStatus) bool {
	for _, item := range es {
		if item == eventStatus {
			return true
		}
	}
	return false
}
