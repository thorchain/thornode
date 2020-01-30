package common

import (
	"fmt"
	"strings"
)

var EmptyBlame = Blame{}

type Blame struct {
	FailReason string   `json:"fail_reason"`
	BlameNodes []string `json:"blame_peers"`
}

func (b Blame) IsEmpty() bool {
	return len(b.BlameNodes) == 0 || len(b.FailReason) == 0
}

func (b Blame) String() string {
	sb := strings.Builder{}
	sb.WriteString("reason:" + b.FailReason + "\n")
	sb.WriteString("nodes:" + fmt.Sprintf("%+v", b.BlameNodes) + "\n")
	return sb.String()
}
