// Please put all the test related function to here
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
	"gitlab.com/thorchain/bepswap/common"
)

// GetRandomNodeAccount create a random generated node account , used for test purpose
func GetRandomNodeAccount(status NodeStatus) NodeAccount {
	name := RandStringBytesMask(10)
	addr := sdk.AccAddress(crypto.AddressHash([]byte(name)))
	bnb, _ := common.NewBnbAddress("tbnb" + RandStringBytesMask(39))
	v, _ := tmtypes.RandValidator(true, 100)
	k, _ := sdk.Bech32ifyConsPub(v.PubKey)
	na := NewNodeAccount(addr, status, NewTrustAccount(bnb, addr, k))
	return na
}

func GetRandomBNBAddress() common.BnbAddress {
	bnb, _ := common.NewBnbAddress("tbnb" + RandStringBytesMask(39))
	return bnb
}

func GetRandomTxHash() common.TxID {
	txHash, _ := common.NewTxID(RandStringBytesMask(64))
	return txHash
}
