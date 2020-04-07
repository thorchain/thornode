package common

import (
	"encoding/json"
	"time"

	. "gopkg.in/check.v1"
)

type DurationTestSuit struct{}

var _ = Suite(&DurationTestSuit{})

func (DurationTestSuit) TestDuration(c *C) {
	d := Duration{Duration: time.Second}
	result, err := json.Marshal(d)
	c.Assert(err, IsNil)
	c.Assert(result, NotNil)
	c.Assert(string(result), Equals, `"1s"`)
}

func (DurationTestSuit) TestUnmarshal(c *C) {
	inputs := []struct {
		duration       string
		errChecker     Checker
		expectedResult time.Duration
	}{
		{
			duration:       `"1s"`,
			errChecker:     IsNil,
			expectedResult: time.Second,
		},
		{
			duration:       `1000000000.0`,
			errChecker:     IsNil,
			expectedResult: time.Second,
		},
		{
			duration:       `true`,
			errChecker:     NotNil,
			expectedResult: time.Duration(0),
		},
		{
			duration:       `"whatev"`,
			errChecker:     NotNil,
			expectedResult: time.Duration(0),
		},
		{
			duration:       `"{"name":{what??}}"`,
			errChecker:     NotNil,
			expectedResult: time.Duration(0),
		},
	}
	for _, item := range inputs {
		var d Duration
		err := json.Unmarshal([]byte(item.duration), &d)
		c.Assert(err, item.errChecker)
		if item.expectedResult != time.Duration(0) {
			c.Assert(d.Duration, Equals, item.expectedResult)
		}
	}
}
