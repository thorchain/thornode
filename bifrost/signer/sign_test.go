package signer

import (
	"encoding/json"
	"testing"

	"github.com/tendermint/tendermint/crypto"

	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	ttypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type SignSuite struct{}

var _ = Suite(&SignSuite{})

type MockCheckTransactionChain struct {
	chainclients.DummyChain
	validateMetaData bool
}

func (m *MockCheckTransactionChain) ValidateMetadata(_ interface{}) bool {
	return m.validateMetaData
}

func (s *SignSuite) TestCheckTxn(c *C) {
	storage, err := NewSignerStore("", "")
	c.Assert(err, IsNil)

	mockChain := &MockCheckTransactionChain{
		validateMetaData: true,
	}
	chain, err := common.NewChain("MOCK")
	c.Assert(err, IsNil)

	chains := make(map[common.Chain]chainclients.ChainClient)
	chains[chain] = mockChain

	signer := &Signer{
		chains:  chains,
		storage: storage,
	}

	status, err := signer.CheckTransaction("", "bad chain", nil)
	c.Assert(err, NotNil)
	c.Check(status, Equals, TxUnknown)

	status, err = signer.CheckTransaction("", chain, nil)
	c.Assert(err, IsNil)
	c.Check(status, Equals, TxUnavailable)

	tx := NewTxOutStoreItem(12, ttypes.TxOutItem{Memo: "foo"})
	c.Assert(storage.Set(tx), IsNil)

	status, err = signer.CheckTransaction(tx.Key(), chain, nil)
	c.Assert(err, IsNil)
	c.Check(status, Equals, TxAvailable)

	spent := NewTxOutStoreItem(100, ttypes.TxOutItem{Memo: "spent"})
	spent.Status = TxSpent
	c.Assert(storage.Set(spent), IsNil)

	status, err = signer.CheckTransaction(spent.Key(), chain, nil)
	c.Assert(err, IsNil)
	c.Check(status, Equals, TxSpent)
}

type MockChainClient struct {
	account common.Account
}

func (b *MockChainClient) SignTx(tai stypes.TxOutItem, height int64, keys common.PubKeys) ([]byte, error) {
	return nil, nil
}

func (b *MockChainClient) GetHeight() (int64, error) {
	return 0, nil
}

func (b *MockChainClient) GetGasFee(count uint64) common.Gas {
	return common.GetBNBGasFee(count)
}

func (b *MockChainClient) CheckIsTestNet() (string, bool) {
	return "", true
}

func (b *MockChainClient) GetChain() common.Chain {
	return common.BNBChain
}

func (b *MockChainClient) ValidateMetadata(inter interface{}) bool {
	return true
}

func (b *MockChainClient) BroadcastTx(tx []byte) error {
	return nil
}

func (b *MockChainClient) GetAddress(poolPubKey common.PubKey) string {
	return "0dd3d0a4a6eacc98cc4894791702e46c270bde76"
}

func (b *MockChainClient) GetAccount(addr string) (common.Account, error) {
	return b.account, nil
}

func (b *MockChainClient) GetPubKey() crypto.PubKey {
	return nil
}

func (b *MockChainClient) Start(globalTxsQueue chan stypes.TxIn, pubkeyMgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) error {
	return nil
}

func (b *MockChainClient) Stop() error {
	return nil
}

func (s *SignSuite) TestHandleYggReturn_Success_FeeSingleton(c *C) {
	sign := &Signer{
		chains: map[common.Chain]chainclients.ChainClient{
			common.BNBChain: &MockChainClient{
				account: common.Account{
					Coins: common.AccountCoins{
						common.AccountCoin{Denom: common.BNBChain.String(), Amount: 1000000},
					},
				},
			},
		},
	}
	input := `{ "chain": "BNB", "memo": "yggdrasil-", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [] }`
	var item stypes.TxOutItem
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)

	newItem, err := sign.handleYggReturn(item)
	c.Assert(err, IsNil)
	c.Check(newItem.Coins[0].Amount.Uint64(), Equals, uint64(962500))
}

func (s *SignSuite) TestHandleYggReturn_Success_FeeMulti(c *C) {
	sign := &Signer{
		chains: map[common.Chain]chainclients.ChainClient{
			common.BNBChain: &MockChainClient{
				account: common.Account{
					Coins: common.AccountCoins{
						common.AccountCoin{Denom: common.BNBChain.String(), Amount: 1000000},
						common.AccountCoin{Denom: "RUNE", Amount: 1000000},
					},
				},
			},
		},
	}
	input := `{ "chain": "BNB", "memo": "yggdrasil-", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [] }`
	var item stypes.TxOutItem
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)

	newItem, err := sign.handleYggReturn(item)
	c.Assert(err, IsNil)
	c.Check(newItem.Coins[0].Amount.Uint64(), Equals, uint64(940000))
}

func (s *SignSuite) TestHandleYggReturn_Success_NotEnough(c *C) {
	sign := &Signer{
		chains: map[common.Chain]chainclients.ChainClient{
			common.BNBChain: &MockChainClient{
				account: common.Account{
					Coins: common.AccountCoins{
						common.AccountCoin{Denom: common.BNBChain.String(), Amount: 0},
					},
				},
			},
		},
	}
	input := `{ "chain": "BNB", "memo": "yggdrasil-", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [] }`
	var item stypes.TxOutItem
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)

	newItem, err := sign.handleYggReturn(item)
	c.Assert(err, IsNil)
	c.Check(newItem.Coins[0].Amount.Uint64(), Equals, uint64(0))
}
