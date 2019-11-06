package common

import (
	"encoding/json"
	"fmt"

	"gitlab.com/thorchain/bepswap/thornode/cmd"
	. "gopkg.in/check.v1"

	"github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	atypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto"
)

type PubKeyTestSuite struct{}

var _ = Suite(&PubKeyTestSuite{})

func (PubKeyTestSuite) SetUpSuite(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
}

// TestPubKey implementation
func (PubKeyTestSuite) TestPubKey(c *C) {
	_, pubKey, _ := atypes.KeyTestPubAddr()
	inputBytes := crypto.AddressHash(pubKey.Bytes())
	pk := NewPubKey(inputBytes)
	hexStr := pk.String()
	c.Assert(len(hexStr) > 0, Equals, true)
	pk1, err := NewPubKeyFromHexString(hexStr)
	c.Assert(err, IsNil)
	c.Assert(pk.Equals(pk1), Equals, true)

	addr, err := pk.GetAddress(BNBChain)
	c.Assert(err, IsNil)
	c.Assert(addr.Equals(NoAddress), Equals, false)

	result, err := json.Marshal(pk)
	c.Assert(err, IsNil)
	c.Log(result, Equals, fmt.Sprintf(`"%s"`, hexStr))
	var pk2 PubKey
	err = json.Unmarshal(result, &pk2)
	c.Assert(err, IsNil)
	c.Assert(pk2.Equals(pk), Equals, true)
}

func (PubKeyTestSuite) TestPubKeyGen(c *C) {
	address := "thor1t60jghk90f820xlf5d6zq9w3t2e4gkxyfuaw4j"
	buf1, err := types.GetFromBech32(address, "thor")
	c.Assert(err, IsNil)
	pk1 := NewPubKey(buf1)
	addr, err := pk1.GetAddress(ThorChain)
	c.Assert(address, Equals, addr.String())
	c.Assert(address, Equals, pk1.GetThorAddress().String())
	acc, err := sdk.AccAddressFromBech32(pk1.GetThorAddress().String())
	c.Assert(err, IsNil)
	c.Check(acc.String(), Equals, address)
}
