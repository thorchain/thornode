package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// StakeTxDetail represent all the stake activity
// Staker can stake on the same pool for multiple times
type StakeTxDetail struct {
	RequestTxHash common.TxID `json:"request_tx_hash"` // the tx hash from binance chain , represent staker send token to the pool
	RuneAmount    sdk.Uint    `json:"rune_amount"`     // amount of rune that send in at the time
	TokenAmount   sdk.Uint    `json:"token_amount"`    // amount of token that send in at the time
}

// StakerPoolItem represent the staker's activity in a pool
// Staker can stake on multiple pool
type StakerPoolItem struct {
	Asset        common.Asset    `json:"asset"`
	Units        sdk.Uint        `json:"units"`
	StakeDetails []StakeTxDetail `json:"stake_details"`
}

func (spi StakerPoolItem) Valid() error {
	if spi.Asset.IsEmpty() {
		return errors.New("Asset cannot be empty")
	}

	for _, detail := range spi.StakeDetails {
		if detail.RequestTxHash.IsEmpty() {
			return errors.New("Request Tx Hash cannot be empty")
		}
	}

	return nil
}

// StakerPool represent staker and their activities in the pools
// A staker can stake on multiple pool.
// A Staker can stake on the same pool for multiple times
// json representation will be looking like
// {
//    "staker_id": "bnbxasdfaswqerqwe",
//    "pool_and_units": [
//        {
//            "asset": "BNB",
//            "units": "200",
//            "stake_details": [
//                {
//                    "request_tx_hash": "txhash from binance chain",
//                    "rune_amount": "100",
//                    "token_amount": "100"
//                },
//                {
//                    "request_tx_hash": "another hash",
//                    "rune_amount": "100",
//                    "token_amount": "100"
//                }
//            ]
//        },
//        {
//            "asset": "BTC",
//            "units": "200",
//            "stake_details": [
//                {
//                    "request_tx_hash": "txhash from binance chain",
//                    "rune_amount": "100",
//                    "token_amount": "100"
//                },
//                {
//                    "request_tx_hash": "another hash",
//                    "rune_amount": "100",
//                    "token_amount": "100"
//                }
//            ]
//        }
//    ]
// }
type StakerPool struct {
	StakerID  common.Address    `json:"staker_id"`      // this will be staker's address on binance chain
	PoolUnits []*StakerPoolItem `json:"pool_and_units"` // the key of this map will be the pool id , value will bt [UNIT,RUNE,TOKEN]
}

// NewStakerPool create a new instance of StakerPool
func NewStakerPool(id common.Address) StakerPool {
	return StakerPool{
		StakerID:  id,
		PoolUnits: []*StakerPoolItem{},
	}
}

func (sp StakerPool) Valid() error {
	if sp.StakerID.IsEmpty() {
		return errors.New("Staker ID cannot be empty")
	}

	for _, item := range sp.PoolUnits {
		if err := item.Valid(); err != nil {
			return err
		}
	}

	return nil
}

// String return a user readable string representation of Staker Pool
func (sp StakerPool) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintln("staker-id: " + sp.StakerID))
	if nil != sp.PoolUnits {
		for _, item := range sp.PoolUnits {
			sb.WriteString(fmt.Sprintf("pool-id: %s, staker unitsL %s", item.Units.String(), item.Units.String()))
		}
	}
	return sb.String()
}

// GetStakerPoolItem return the StakerPoolItem with given pool id
func (sp *StakerPool) GetStakerPoolItem(asset common.Asset) *StakerPoolItem {
	for _, item := range sp.PoolUnits {
		if asset.Equals(item.Asset) {
			return item
		}
	}
	return &StakerPoolItem{
		Asset:        asset,
		Units:        sdk.ZeroUint(),
		StakeDetails: []StakeTxDetail{},
	}
}

// RemoveStakerPoolItem delete the stakerpoolitem with given pool id from the struct
func (sp *StakerPool) RemoveStakerPoolItem(asset common.Asset) {
	deleteIdx := -1
	for idx, item := range sp.PoolUnits {
		if item.Asset.Equals(asset) {
			deleteIdx = idx
		}
	}
	if deleteIdx != -1 {
		if deleteIdx == 0 {
			sp.PoolUnits = []*StakerPoolItem{}
		} else {
			sp.PoolUnits = append(sp.PoolUnits[:deleteIdx-1], sp.PoolUnits[deleteIdx:]...)
		}
	}
}

// UpsertStakerPoolItem if the given stakerPoolItem exist then it will update , otherwise it will add
func (sp *StakerPool) UpsertStakerPoolItem(stakerPoolItem *StakerPoolItem) {
	pos := -1
	for idx, item := range sp.PoolUnits {
		if item.Asset.Equals(stakerPoolItem.Asset) {
			pos = idx
		}
	}
	if pos != -1 {
		sp.PoolUnits[pos] = stakerPoolItem
		return
	}
	sp.PoolUnits = append(sp.PoolUnits, stakerPoolItem)
}

// AddStakerTxDetail to the StakerPool structure
func (spi *StakerPoolItem) AddStakerTxDetail(requestTxHash common.TxID, runeAmount, tokenAmount sdk.Uint) {
	std := StakeTxDetail{
		RequestTxHash: requestTxHash,
		RuneAmount:    runeAmount,
		TokenAmount:   tokenAmount,
	}
	spi.StakeDetails = append(spi.StakeDetails, std)
}
