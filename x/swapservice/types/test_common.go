// Please put all the test related function to here
package types

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
	"gitlab.com/thorchain/bepswap/common"

	"gitlab.com/thorchain/bepswap/statechain/cmd"
)

var addresses = []string{
	"bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u",
	"bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38",
	"bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq",
	"bnb1yk882gllgv3rt2rqrsudf6kn2agr94etnxu9a7",
	"bnb1t3c49u74fum2gtgekwqqdngg5alt4txrq3txad",
	"bnb1hpa7tfffxadq9nslyu2hu9vc44l2x6ech3767y",
	"bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq",
	"bnb1llvmhawaxxjchwmfmj8fjzftvwz4jpdhapp5hr",
	"bnb1s3f8vxaqum3pft6cefyn99px8wq6uk3jdtyarn",
	"bnb1e6y59wuz9qqcnqjhjw0cl6hrp2p8dvsyxyx9jm",
	"bnb1zxseqkfm3en5cw6dh9xgmr85hw6jtwamnd2y2v",
}

// GetRandomNodeAccount create a random generated node account , used for test purpose
func GetRandomNodeAccount(status NodeStatus) NodeAccount {
	name := RandStringBytesMask(10)
	addr := sdk.AccAddress(crypto.AddressHash([]byte(name)))
	bnb := GetRandomBNBAddress()
	v, _ := tmtypes.RandValidator(true, 100)
	k, _ := sdk.Bech32ifyConsPub(v.PubKey)
	bondAddr := GetRandomBNBAddress()
	na := NewNodeAccount(addr, status, NewTrustAccount(bnb, addr, k), sdk.NewUint(100*common.One), bondAddr)
	return na
}

// GetRandomBech32Addr is an account address used for test
func GetRandomBech32Addr() sdk.AccAddress {
	name := RandStringBytesMask(10)
	return sdk.AccAddress(crypto.AddressHash([]byte(name)))
}

func GetRandomBNBAddress() common.Address {
	bnb, _ := common.NewAddress(addresses[rand.Intn(len(addresses))])
	return bnb
}

// GetRandomTxHash create a random txHash used for test purpose
func GetRandomTxHash() common.TxID {
	txHash, _ := common.NewTxID(RandStringBytesMask(64))
	return txHash
}

// SetupConfigForTest used for test purpose
func SetupConfigForTest() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
}
