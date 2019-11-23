package types

import "gitlab.com/thorchain/bepswap/thornode/common"

type LastBlock struct {
	Height string `json:"statechain"`
}

type PoolPubKey struct {
	Chain common.Chain `json:"chain"`
	//SeqNo   uint64         `json:"seq_no"`
	//PubKey  common.PubKey  `json:"pub_key"`
	Address common.Address `json:"address"`
}

type StatechainPoolAddress struct {
	// Previous           []PoolPubKey `json:"previous"`
	Current []PoolPubKey `json:"current"`
	// Next               []PoolPubKey `json:"next"`
	// RotateAt           int64        `json:"rotate_at"`
	// RotateWindowOpenAt int64        `json:"rotate_window_open_at"`
}

type StatechainPools []struct {
	BalanceRune  int64 `json:"balance_rune,string"`
	BalanceAsset int64 `json:"balance_asset,string"`
	Asset        struct {
		Chain  string `json:"chain"`
		Symbol string `json:"symbol"`
		Ticker string `json:"string"`
	} `json:"asset"`
	PoolUnits   int64  `json:"pool_units,string"`
	PoolAddress string `json:"pool_address"`
	Status      string `json:"status"`
}

type Staker struct {
	StakerID     string `json:"staker_id"`
	PoolAndUnits []struct {
		Asset struct {
			Chain  string `json:"chain"`
			Symbol string `json:"symbol"`
			Ticker string `json:"string"`
		} `json:"asset"`
		Units int64 `json:"units,string"`
	} `json:"pool_and_units"`
}
