package types

type Account struct {
	Stream string `json:"stream"`
	Data   struct {
		EventType 	string `json:"e"`
		EventHeight int    `json:"E"`
		Balances 		[]struct {
			Asset 			string `json:"a"`
			Free 				string `json:"f"`
			Frozen 			string `json:"r"`
			Locked 			string `json:"l"`
		} `json:"B"`
	} `json:"data"`
}
