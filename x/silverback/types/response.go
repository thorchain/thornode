package types

type Response struct {
	Stream string `json:"stream"`
	Data   struct {
		EventType 	string `json:"e"`
		EventHeight int    `json:"E"`
	}
}
