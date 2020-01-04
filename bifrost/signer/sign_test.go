package signer

import (
	"encoding/json"
	"testing"

	"github.com/binance-chain/go-sdk/common/types"
	"github.com/tendermint/tendermint/crypto"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type SignSuite struct{}

var _ = Suite(&SignSuite{})

type TestBinance struct {
	baseAccount types.BaseAccount
}

func (b *TestBinance) BroadcastTx(hexTx []byte) error {
	return nil
}
func (b *TestBinance) GetAddress(poolPubKey common.PubKey) string {
	return "0dd3d0a4a6eacc98cc4894791702e46c270bde76"
}
func (b *TestBinance) GetAccount(addr types.AccAddress) (types.BaseAccount, error) {
	return b.baseAccount, nil
}
func (b *TestBinance) GetPubKey() crypto.PubKey {
	return nil
}
func (b *TestBinance) SignTx(tai stypes.TxOutItem, height int64) ([]byte, map[string]string, error) {
	return nil, nil, nil
}

func (s *SignSuite) TestHandleYggReturn_Success_FeeSingleton(c *C) {
	sign := &Signer{
		Binance: &TestBinance{
			baseAccount: types.BaseAccount{
				Coins: types.Coins{
					types.Coin{Denom: "BNB", Amount: 1000000},
				},
			},
		},
	}
	input := `{ "memo": "yggdrasil-", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [] }`
	var item stypes.TxOutItem
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)

	newItem, err := sign.handleYggReturn(item)
	c.Assert(err, IsNil)
	c.Check(newItem.Coins[0].Amount.Uint64(), Equals, uint64(962500))
}

func (s *SignSuite) TestHandleYggReturn_Success_FeeMulti(c *C) {
	sign := &Signer{
		Binance: &TestBinance{
			baseAccount: types.BaseAccount{
				Coins: types.Coins{
					types.Coin{Denom: "BNB", Amount: 1000000},
					types.Coin{Denom: "RUNE", Amount: 1000000},
				},
			},
		},
	}
	input := `{ "memo": "yggdrasil-", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [] }`
	var item stypes.TxOutItem
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)

	newItem, err := sign.handleYggReturn(item)
	c.Assert(err, IsNil)
	c.Check(newItem.Coins[0].Amount.Uint64(), Equals, uint64(940000))
}

func (s *SignSuite) TestHandleYggReturn_Success_NotEnough(c *C) {
	sign := &Signer{
		Binance: &TestBinance{
			baseAccount: types.BaseAccount{
				Coins: types.Coins{
					types.Coin{Denom: "BNB", Amount: 10000},
				},
			},
		},
	}
	input := `{ "memo": "yggdrasil-", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [] }`
	var item stypes.TxOutItem
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)

	newItem, err := sign.handleYggReturn(item)
	c.Assert(err, IsNil)
	c.Check(newItem.Coins[0].Amount.Uint64(), Equals, uint64(0))
}
