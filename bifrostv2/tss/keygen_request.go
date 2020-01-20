package tss

// KeyGenRequest is the request send to tss_keygen
type KeyGenRequest struct {
	Keys []string `json:"keys"`
}
