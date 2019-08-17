package types

type Signature struct {
	PubKey struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
	Signature string `json:"signature"`
}
