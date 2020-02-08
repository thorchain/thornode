package common

import (
	ckeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type TxSuite struct{}

var _ = Suite(&TxSuite{})

func (s TxSuite) TestTxID(c *C) {
	ID := "A7DA8FF1B7C290616D68A276F30AC618315E6CCE982EB8F7A79339E163798F49"
	tx, err := NewTxID(ID)
	c.Assert(err, IsNil)
	c.Check(tx.String(), Equals, ID)
	c.Check(tx.IsEmpty(), Equals, false)
	c.Check(tx.Equals(TxID(ID)), Equals, true)

	// check eth hash
	_, err = NewTxID("0xb41cf456e942f3430681298c503def54b79a96e3373ef9d44ea314d7eae41952")
	c.Assert(err, IsNil)

	_, err = NewTxID("bogus")
	c.Check(err, NotNil)
}

func (s TxSuite) TestTx(c *C) {
	id, err := NewTxID("0xb41cf456e942f3430681298c503def54b79a96e3373ef9d44ea314d7eae41952")
	c.Assert(err, IsNil)
	tx := NewTx(
		id,
		Address("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6"),
		Address("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6"),
		Coins{NewCoin(BNBAsset, sdk.NewUint(5*One))},
		BNBGasFeeSingleton,
		"hello memo",
	)
	c.Check(tx.ID.Equals(id), Equals, true)
	c.Check(tx.IsEmpty(), Equals, false)
	c.Check(tx.FromAddress.IsEmpty(), Equals, false)
	c.Check(tx.ToAddress.IsEmpty(), Equals, false)
	c.Assert(tx.Coins, HasLen, 1)
	c.Check(tx.Coins[0].Equals(NewCoin(BNBAsset, sdk.NewUint(5*One))), Equals, true)
	c.Check(tx.Memo, Equals, "hello memo")

	// Test that we can sign/verify a tx
	keys := ckeys.NewInMemory()
	keys.CreateMnemonic("user", ckeys.English, "password", ckeys.Secp256k1)
	priv, err := keys.ExportPrivateKeyObject("user", "password")
	c.Assert(err, IsNil)
	sig, err := tx.Sign(priv)
	c.Assert(err, IsNil)
	c.Check(sig, HasLen, 64, Commentf("%d", len(sig)))
	ok, err := tx.Verify(priv.PubKey(), sig)
	c.Assert(err, IsNil)
	c.Check(ok, Equals, true)

	// verify that a bad signature does fail
	keys2 := ckeys.NewInMemory()
	keys2.CreateMnemonic("user2", ckeys.English, "password", ckeys.Secp256k1)
	priv2, err := keys2.ExportPrivateKeyObject("user2", "password")
	c.Assert(err, IsNil)
	ok, err = tx.Verify(priv2.PubKey(), sig)
	c.Assert(err, IsNil)
	c.Check(ok, Equals, false)
}
