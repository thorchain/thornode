package tss

// KeyGenResponse
type KeyGenResp struct {
	PubKey     string `json:"pub_key"`
	BNBAddress string `json:"bnb_address"`
	Status     int    `json:"status"`
}
