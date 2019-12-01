package common

import (
	"encoding/json"
	"errors"
	"time"
)

// Duration embedded time.Duration so THORNode could use string to represent duration in json file
// for example ,1s ,1h , 5m etc
type Duration struct {
	time.Duration
}

// MarshalJSON marshal the duration to json string
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON convert the json value back to time.Duration
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}
