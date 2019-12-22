// Please put all the test related function to here
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	atypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	"gitlab.com/thorchain/thornode/cmd"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// GetRandomNodeAccount create a random generated node account , used for test purpose
func GetRandomNodeAccount(status NodeStatus) NodeAccount {
	v, _ := tmtypes.RandValidator(true, 100)
	k, _ := sdk.Bech32ifyConsPub(v.PubKey)
	bondAddr := GetRandomBNBAddress()
	pubKeys := common.PubKeys{
		Secp256k1: GetRandomPubKey(),
		Ed25519:   GetRandomPubKey(),
	}
	addr, _ := pubKeys.Secp256k1.GetThorAddress()
	na := NewNodeAccount(addr, status, pubKeys, k, sdk.NewUint(100*common.One), bondAddr, 1)
	na.Version = constants.SWVersion
	if na.Status == Active {
		na.ActiveBlockHeight = 10
	}

	return na
}

func GetRandomObservedTx() ObservedTx {
	return NewObservedTx(GetRandomTx(), sdk.NewUint(33), GetRandomPubKey())
}

// GetRandomTx
func GetRandomTx() common.Tx {
	return common.NewTx(
		GetRandomTxHash(),
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		common.BNBGasFeeSingleton,
		"",
	)
}

// GetRandomBech32Addr is an account address used for test
func GetRandomBech32Addr() sdk.AccAddress {
	name := common.RandStringBytesMask(10)
	return sdk.AccAddress(crypto.AddressHash([]byte(name)))
}

func GetRandomBech32ConsensusPubKey() string {
	_, pubKey, _ := atypes.KeyTestPubAddr()
	result, err := sdk.Bech32ifyConsPub(pubKey)
	if nil != err {
		panic(err)
	}
	return result
}

// GetRandomBNBAddress will just create a random bnb address used for test purpose
func GetRandomBNBAddress() common.Address {
	name := common.RandStringBytesMask(10)
	str, _ := common.ConvertAndEncode("tbnb", crypto.AddressHash([]byte(name)))
	bnb, _ := common.NewAddress(str)
	return bnb
}

// GetRandomTxHash create a random txHash used for test purpose
func GetRandomTxHash() common.TxID {
	txHash, _ := common.NewTxID(common.RandStringBytesMask(64))
	return txHash
}

// GetRandomPubKeys return a random common.PubKeys for test purpose
func GetRandomPubKeys() common.PubKeys {
	return common.NewPubKeys(GetRandomPubKey(), GetRandomPubKey())
}

func GetRandomVault() Vault {
	return NewVault(32, ActiveVault, AsgardVault, GetRandomPubKey())
}

func GetRandomPubKey() common.PubKey {
	_, pubKey, _ := atypes.KeyTestPubAddr()
	bech32PubKey, _ := sdk.Bech32ifyAccPub(pubKey)
	pk, _ := common.NewPubKey(bech32PubKey)
	return pk
}

// SetupConfigForTest used for test purpose
func SetupConfigForTest() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
}
