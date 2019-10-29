package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/binance-chain/go-sdk/common/types"
	atypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto"
	. "gopkg.in/check.v1"
)

type PubKeyTestSuite struct{}

var _ = Suite(&PubKeyTestSuite{})

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
func (PubKeyTestSuite) TestStuff(c *C) {
	if err := os.Setenv("NET", "testnet"); nil != err {
		panic(err)
	}
	address := "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6"
	buf1, err := types.GetFromBech32(address, "bnb")
	c.Assert(err, IsNil)
	pk1 := NewPubKey(buf1)
	fmt.Println(pk1.String())
	addr, err := pk1.GetAddress(BNBChain)
	fmt.Println(addr.String())
}
