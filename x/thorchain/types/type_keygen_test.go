package types

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeygenSuite struct{}

var _ = Suite(&KeygenSuite{})

func (s *KeygenSuite) TestKengenType(c *C) {
	input := map[KeygenType]string{
		UnknownKeygen:   "unknown",
		AsgardKeygen:    "asgard",
		YggdrasilKeygen: "yggdrasil",
	}
	for k, v := range input {
		c.Assert(k.String(), Equals, v)
	}
}

func (s *KeygenSuite) TestKeygen(c *C) {
	var members common.PubKeys
	for i := 0; i < 4; i++ {
		members = append(members, GetRandomPubKey())
	}
	keygen, err := NewKeygen(1, members, AsgardKeygen)
	c.Assert(err, IsNil)
	c.Assert(keygen.IsEmpty(), Equals, false)
	c.Assert(keygen.Valid(), IsNil)
	c.Log(keygen.String())
}

func (s *KeygenSuite) TestGetKeygenID(c *C) {
	var members common.PubKeys
	for i := 0; i < 4; i++ {
		members = append(members, GetRandomPubKey())
	}
	txID, err := getKeygenID(1, members, AsgardKeygen)
	c.Assert(err, IsNil)
	c.Assert(txID.IsEmpty(), Equals, false)
	txID1, err := getKeygenID(2, members, AsgardKeygen)
	c.Assert(err, IsNil)
	c.Assert(txID1.IsEmpty(), Equals, false)
	// with different block height two keygen item should be different
	c.Assert(txID1.Equals(txID), Equals, false)
	// with different
	txID2, err := getKeygenID(1, members, YggdrasilKeygen)
	c.Assert(err, IsNil)
	c.Assert(txID.Equals(txID2), Equals, false)

	txID3, err := getKeygenID(1, members, AsgardKeygen)
	c.Assert(err, IsNil)
	c.Assert(txID3.Equals(txID), Equals, true)
}

func (s *KeygenSuite) TestNewKeygenBlock(c *C) {
	kb := NewKeygenBlock(1)
	c.Assert(kb.IsEmpty(), Equals, false)
}
