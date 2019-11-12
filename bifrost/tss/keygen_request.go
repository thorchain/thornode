package tss

// KeyGenRequest is the request send to tss_keygen
type KeyGenRequest struct {
	PubKeys []string `json:"pub_keys"`
	PrivKey string   `json:"priv_key"`
}
