package signer

import (
	. "gopkg.in/check.v1"
)

type EncryptionSuite struct{}

var _ = Suite(&EncryptionSuite{})

func (s *EncryptionSuite) TestEncryption(c *C) {
	body := []byte("hello world!")
	passphrase := "my super secret password!"

	encryp, err := encrypt(body, passphrase)
	c.Assert(err, IsNil)

	decryp, err := decrypt(encryp, passphrase)
	c.Assert(err, IsNil)

	c.Check(body, DeepEquals, decryp)
}
