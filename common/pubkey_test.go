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

type KeyDataAddr struct {
	mainnet string
	testnet string
	mocknet string
}

type KeyData struct {
	priv     string
	pub      string
	addrBNB  KeyDataAddr
	addrBTC  KeyDataAddr
	addrETH  KeyDataAddr
	addrTHOR KeyDataAddr
}

type PubKeyTestSuite struct {
	keyData []KeyData
}

var _ = Suite(&PubKeyTestSuite{})

func (s *PubKeyTestSuite) SetUpSuite(c *C) {
	s.keyData = []KeyData{
		{
			priv: "ef235aacf90d9f4aadd8c92e4b2562e1d9eb97f0df9ba3b508258739cb013db2",
			pub:  "02b4632d08485ff1df2db55b9dafd23347d1c47a457072a1e87be26896549a8737",
			addrETH: KeyDataAddr{
				mainnet: "0x3fd2d4ce97b082d4bce3f9fee2a3d60668d2f473",
				testnet: "0x3fd2d4ce97b082d4bce3f9fee2a3d60668d2f473",
				mocknet: "0x3fd2d4ce97b082d4bce3f9fee2a3d60668d2f473",
			},
			addrBNB: KeyDataAddr{
				mainnet: "bnb1j08ys4ct2hzzc2hcz6h2hgrvlmsjynawtf2n0y",
				testnet: "tbnb1j08ys4ct2hzzc2hcz6h2hgrvlmsjynaw9urh04",
				mocknet: "tbnb1j08ys4ct2hzzc2hcz6h2hgrvlmsjynaw9urh04",
			},
			addrBTC: KeyDataAddr{
				mainnet: "bc1qj08ys4ct2hzzc2hcz6h2hgrvlmsjynawlht528",
				testnet: "tb1qj08ys4ct2hzzc2hcz6h2hgrvlmsjynaw43s835",
				mocknet: "bcrt1qj08ys4ct2hzzc2hcz6h2hgrvlmsjynawhcf2xa",
			},
		},
		{
			priv: "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
			pub:  "037db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf7",
			addrETH: KeyDataAddr{
				mainnet: "0x970e8128ab834e8eac17ab8e3812f010678cf791",
				testnet: "0x970e8128ab834e8eac17ab8e3812f010678cf791",
				mocknet: "0x970e8128ab834e8eac17ab8e3812f010678cf791",
			},
			addrBNB: KeyDataAddr{
				mainnet: "bnb1zupk5lmc84r2dh738a9g3zscavannjy3nkkcrl",
				testnet: "tbnb1zupk5lmc84r2dh738a9g3zscavannjy3arlurw",
				mocknet: "tbnb1zupk5lmc84r2dh738a9g3zscavannjy3arlurw",
			},
			addrBTC: KeyDataAddr{
				mainnet: "bc1qzupk5lmc84r2dh738a9g3zscavannjy38ghlxu",
				testnet: "tb1qzupk5lmc84r2dh738a9g3zscavannjy3dwvva0",
				mocknet: "bcrt1qzupk5lmc84r2dh738a9g3zscavannjy3084p2x",
			},
		},
		{
			priv: "e810f1d7d6691b4a7a73476f3543bd87d601f9a53e7faf670eac2c5b517d83bf",
			pub:  "03f98464e8d3fc8e275e34c6f8dc9b99aa244e37b0d695d0dfb8884712ed6d4d35",
			addrETH: KeyDataAddr{
				mainnet: "0xf6da288748ec4c77642f6c5543717539b3ae001b",
				testnet: "0xf6da288748ec4c77642f6c5543717539b3ae001b",
				mocknet: "0xf6da288748ec4c77642f6c5543717539b3ae001b",
			},
			addrBNB: KeyDataAddr{
				mainnet: "bnb1qqnde7kqe5sf96j6zf8jpzwr44dh4gkddg5yfw",
				testnet: "tbnb1qqnde7kqe5sf96j6zf8jpzwr44dh4gkdraaqfl",
				mocknet: "tbnb1qqnde7kqe5sf96j6zf8jpzwr44dh4gkdraaqfl",
			},
			addrBTC: KeyDataAddr{
				mainnet: "bc1qqqnde7kqe5sf96j6zf8jpzwr44dh4gkdek4rvd",
				testnet: "tb1qqqnde7kqe5sf96j6zf8jpzwr44dh4gkdnswsh7",
				mocknet: "bcrt1qqqnde7kqe5sf96j6zf8jpzwr44dh4gkd3ehaqh",
			},
		},
		{
			priv: "a96e62ed3955e65be32703f12d87b6b5cf26039ecfa948dc5107a495418e5330",
			pub:  "02950e1cdfcb133d6024109fd489f734eeb4502418e538c28481f22bce276f248c",
			addrETH: KeyDataAddr{
				mainnet: "0xfabb9cc6ec839b1214bb11c53377a56a6ed81762",
				testnet: "0xfabb9cc6ec839b1214bb11c53377a56a6ed81762",
				mocknet: "0xfabb9cc6ec839b1214bb11c53377a56a6ed81762",
			},
			addrBNB: KeyDataAddr{
				mainnet: "bnb10s4mg25tu6termrk8egltfyme4q7sg3hm84ayj",
				testnet: "tbnb10s4mg25tu6termrk8egltfyme4q7sg3h4jueyr",
				mocknet: "tbnb10s4mg25tu6termrk8egltfyme4q7sg3h4jueyr",
			},
			addrBTC: KeyDataAddr{
				mainnet: "bc1q0s4mg25tu6termrk8egltfyme4q7sg3h0e56p3",
				testnet: "tb1q0s4mg25tu6termrk8egltfyme4q7sg3h9l0f6z",
				mocknet: "bcrt1q0s4mg25tu6termrk8egltfyme4q7sg3h8kkydt",
			},
		},
		{
			priv: "9294f4d108465fd293f7fe299e6923ef71a77f2cb1eb6d4394839c64ec25d5c0",
			pub:  "0238383ee4d60176d27cf46f0863bfc6aea624fe9bfc7f4273cc5136d9eb483e4a",
			addrETH: KeyDataAddr{
				mainnet: "0x1f30a82340f08177aba70e6f48054917c74d7d38",
				testnet: "0x1f30a82340f08177aba70e6f48054917c74d7d38",
				mocknet: "0x1f30a82340f08177aba70e6f48054917c74d7d38",
			},
			addrBNB: KeyDataAddr{
				mainnet: "bnb1jw8h4l3dtz5xxc7uyh5ys70qkezspgfuh4nh0z",
				testnet: "tbnb1jw8h4l3dtz5xxc7uyh5ys70qkezspgfueq6n0n",
				mocknet: "tbnb1jw8h4l3dtz5xxc7uyh5ys70qkezspgfueq6n0n",
			},
			addrBTC: KeyDataAddr{
				mainnet: "bc1qjw8h4l3dtz5xxc7uyh5ys70qkezspgfurtjs2p",
				testnet: "tb1qjw8h4l3dtz5xxc7uyh5ys70qkezspgfufdfr3j",
				mocknet: "bcrt1qjw8h4l3dtz5xxc7uyh5ys70qkezspgfutyswxm",
			},
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
	original := os.Getenv("NET")
	defer func() {
		os.Setenv("NET", original)
	}()

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

		os.Setenv("NET", "mainnet")
		addrETH, err := pk.GetAddress(ETHChain)
		c.Assert(err, IsNil)
		c.Assert(addrETH.String(), Equals, d.addrETH.mainnet)

		addrBNB, err := pk.GetAddress(BNBChain)
		c.Assert(err, IsNil)
		c.Assert(addrBNB.String(), Equals, d.addrBNB.mainnet)

		addrBTC, err := pk.GetAddress(BTCChain)
		c.Assert(err, IsNil)
		c.Assert(addrBTC.String(), Equals, d.addrBTC.mainnet)

		os.Setenv("NET", "testnet")
		addrETH, err = pk.GetAddress(ETHChain)
		c.Assert(err, IsNil)
		c.Assert(addrETH.String(), Equals, d.addrETH.testnet)

		addrBNB, err = pk.GetAddress(BNBChain)
		c.Assert(err, IsNil)
		c.Assert(addrBNB.String(), Equals, d.addrBNB.testnet)

		addrBTC, err = pk.GetAddress(BTCChain)
		c.Assert(err, IsNil)
		c.Assert(addrBTC.String(), Equals, d.addrBTC.testnet)

		os.Setenv("NET", "mocknet")
		addrETH, err = pk.GetAddress(ETHChain)
		c.Assert(err, IsNil)
		c.Assert(addrETH.String(), Equals, d.addrETH.mocknet)

		addrBNB, err = pk.GetAddress(BNBChain)
		c.Assert(err, IsNil)
		c.Assert(addrBNB.String(), Equals, d.addrBNB.mocknet)

		addrBTC, err = pk.GetAddress(BTCChain)
		c.Assert(err, IsNil)
		c.Assert(addrBTC.String(), Equals, d.addrBTC.mocknet)

	}
}
