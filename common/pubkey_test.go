package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	atypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	. "gopkg.in/check.v1"
)

type KeyData struct {
	priv     string
	pub      string
	addrBNB  string
	addrBTC  string
	addrETH  string
	addrTHOR string
}

type PubKeyTestSuite struct {
	keyData []KeyData
}

var _ = Suite(&PubKeyTestSuite{})

func (s *PubKeyTestSuite) SetUpSuite(c *C) {
	s.keyData = []KeyData{
		{
			priv:    "ef235aacf90d9f4aadd8c92e4b2562e1d9eb97f0df9ba3b508258739cb013db2",
			pub:     "02b4632d08485ff1df2db55b9dafd23347d1c47a457072a1e87be26896549a8737",
			addrETH: "0x3fd2D4cE97B082d4BcE3f9fee2A3D60668D2f473",
			addrBNB: "bnb1j08ys4ct2hzzc2hcz6h2hgrvlmsjynawtf2n0y",
			addrBTC: "bc1qj08ys4ct2hzzc2hcz6h2hgrvlmsjynawlht528",
		},
		{
			priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
			pub:     "037db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf7",
			addrETH: "0x970E8128AB834E8EAC17Ab8E3812F010678CF791",
			addrBTC: "bc1qzupk5lmc84r2dh738a9g3zscavannjy38ghlxu",
			addrBNB: "bnb1zupk5lmc84r2dh738a9g3zscavannjy3nkkcrl",
		},
		{
			priv:    "e810f1d7d6691b4a7a73476f3543bd87d601f9a53e7faf670eac2c5b517d83bf",
			pub:     "03f98464e8d3fc8e275e34c6f8dc9b99aa244e37b0d695d0dfb8884712ed6d4d35",
			addrETH: "0xF6dA288748eC4c77642F6c5543717539B3Ae001b",
			addrBTC: "bc1qqqnde7kqe5sf96j6zf8jpzwr44dh4gkdek4rvd",
			addrBNB: "bnb1qqnde7kqe5sf96j6zf8jpzwr44dh4gkddg5yfw",
		},
		{
			priv:    "a96e62ed3955e65be32703f12d87b6b5cf26039ecfa948dc5107a495418e5330",
			pub:     "02950e1cdfcb133d6024109fd489f734eeb4502418e538c28481f22bce276f248c",
			addrETH: "0xFabB9cC6Ec839b1214bB11c53377A56A6Ed81762",
			addrBTC: "bc1q0s4mg25tu6termrk8egltfyme4q7sg3h0e56p3",
			addrBNB: "bnb10s4mg25tu6termrk8egltfyme4q7sg3hm84ayj",
		},
	}
}

// TestPubKey implementation
func (s *PubKeyTestSuite) TestPubKey(c *C) {
	_, pubKey, _ := atypes.KeyTestPubAddr()
	spk, err := sdk.Bech32ifyAccPub(pubKey)
	c.Assert(err, IsNil)
	pk, err := NewPubKey(spk)
	c.Assert(err, IsNil)
	hexStr := pk.String()
	c.Assert(len(hexStr) > 0, Equals, true)
	pk1, err := NewPubKey(hexStr)
	c.Assert(err, IsNil)
	c.Assert(pk.Equals(pk1), Equals, true)

	result, err := json.Marshal(pk)
	c.Assert(err, IsNil)
	c.Log(result, Equals, fmt.Sprintf(`"%s"`, hexStr))
	var pk2 PubKey
	err = json.Unmarshal(result, &pk2)
	c.Assert(err, IsNil)
	c.Assert(pk2.Equals(pk), Equals, true)
}

func (s *PubKeyTestSuite) TestPubKeySet(c *C) {
	_, pubKey, _ := atypes.KeyTestPubAddr()
	spk, err := sdk.Bech32ifyAccPub(pubKey)
	c.Assert(err, IsNil)
	pk, err := NewPubKey(spk)
	c.Assert(err, IsNil)

	c.Check(PubKeySet{}.Contains(pk), Equals, false)

	pks := PubKeySet{
		Secp256k1: pk,
	}
	c.Check(pks.Contains(pk), Equals, true)
	pks = PubKeySet{
		Ed25519: pk,
	}
	c.Check(pks.Contains(pk), Equals, true)
}

func (s *PubKeyTestSuite) TestPubKeyGetAddress(c *C) {
	os.Setenv("NET", "mainnet")
	for _, d := range s.keyData {
		privB, _ := hex.DecodeString(d.priv)
		pubB, _ := hex.DecodeString(d.pub)

		var priv secp256k1.PrivKeySecp256k1
		copy(priv[:], privB)

		pubKey := priv.PubKey()
		pubT, _ := pubKey.(secp256k1.PubKeySecp256k1)
		pub := pubT[:]

		c.Assert(hex.EncodeToString(pub), Equals, hex.EncodeToString(pubB))

		pubBech32, err := sdk.Bech32ifyAccPub(pubKey)
		c.Assert(err, IsNil)

		pk, err := NewPubKey(pubBech32)
		c.Assert(err, IsNil)

		addrETH, err := pk.GetAddress(ETHChain)
		c.Assert(err, IsNil)
		c.Assert(addrETH.String(), Equals, d.addrETH)

		addrBTC, err := pk.GetAddress(BTCChain)
		c.Assert(err, IsNil)
		c.Assert(addrBTC.String(), Equals, d.addrBTC)

		addrBNB, err := pk.GetAddress(BNBChain)
		c.Assert(err, IsNil)
		c.Assert(addrBNB.String(), Equals, d.addrBNB)
	}
}
