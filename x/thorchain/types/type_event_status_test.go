package types

import (
	"encoding/json"
	"strings"

	. "gopkg.in/check.v1"
)

type EventStatusSuite struct{}

var _ = Suite(&EventStatusSuite{})

func (EventStatusSuite) TestEventStatus(c *C) {
	input := []string{
		"success", "failed", "pending",
	}
	for _, item := range input {
		es := GetEventStatus(item)
		err := es.Valid()
		c.Assert(err, IsNil)
		c.Check(strings.EqualFold(es.String(), item), Equals, true)
	}
	invalidEventStatus := EventStatus(len(input) + 100)
	c.Assert(invalidEventStatus.Valid(), NotNil)

	es := GetEventStatus("success")
	buf, err := json.Marshal(es)
	c.Assert(err, IsNil)
	c.Check(strings.EqualFold(string(buf), `"success"`), Equals, true)
	var es1 EventStatus
	err = json.Unmarshal([]byte(`"success"`), &es1)
	c.Assert(err, IsNil)
	c.Check(es1 == Success, Equals, true)
	c.Check(GetEventStatus("whatever") == Pending, Equals, true)
	err1 := json.Unmarshal([]byte("test"), &es1)
	c.Assert(err1, NotNil)
}
