package types

type Account struct {
	Stream string `json:"stream"`
	Data	struct {
		Event       string `json:"e"`
		EventHeight int    `json:"E"`
		H           string `json:"H"`
		From 				string `json:"f"`
		T	[]struct {
			O string `json:"o"`
			C []struct {
				Asset string `json:"a"`
				A     string `json:"A"`
			} `json:"c"`
		} `json:"t"`
	} `json:"data"`
}
