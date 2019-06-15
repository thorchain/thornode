package types

import "strings"

// Query Result Payload for a resolve query
type QueryResResolve struct {
	Value string `json:"value"`
}

// implement fmt.Stringer
func (r QueryResResolve) String() string {
	return r.Value
}

// Query Result Payload for a pooldatas query
type QueryResPoolDatas []string

// implement fmt.Stringer
func (n QueryResPoolDatas) String() string {
	return strings.Join(n[:], "\n")
}
