// Please put all the test related function to here
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
	"gitlab.com/thorchain/bepswap/common"

	"gitlab.com/thorchain/bepswap/statechain/cmd"
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

func SetupConfigForTest() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
}
