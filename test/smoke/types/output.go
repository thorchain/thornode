package types

type TestResults struct {
	Tx   int     `json:"TX"`
	Rune Balance `json:"BNB.RUNE-A1F"`
	Lok  Balance `json:"BNB.LOK-3C0"`
	Bnb  Balance `json:"BNB.BNB"`
}

type Balance struct {
	Master  int64 `json:"MASTER"`
	Admin   int64 `json:"ADMIN"`
	User    int64 `json:"USER-1"`
	Staker1 int64 `json:"STAKER-1"`
	Staker2 int64 `json:"STAKER-2"`
	Pool    int64 `json:"POOL"`
}
