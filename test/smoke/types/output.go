package types

type Output struct {
	Tx      int     `json:"TX"`
	Master  Balance `json:"MASTER"`
	Admin   Balance `json:"ADMIN"`
	User    Balance `json:"USER-1"`
	Staker1 Balance `json:"STAKER-1"`
	Staker2 Balance `json:"STAKER-2"`
	Pool    Balance `json:"POOL"`
}

type Balance struct {
	Rune int64 `json:"BNB.RUNE-A1F"`
	Bnb  int64 `json:"BNB.BNB"`
	Lok  int64 `json:"BNB.LOK-3C0"`
}
