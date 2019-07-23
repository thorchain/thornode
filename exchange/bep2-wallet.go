package exchange

import (
	"bytes"
	"encoding/gob"
	"strings"

	"github.com/pkg/errors"
)

// Bep2Wallet has all the required field members and methods to deal with wallets
type Bep2Wallet struct {
	PublicAddress string
	PrivateKey    string
	Mnemonic      string
	AssetSymbol   string
}

// FromBytes create a new instance of Bep2Wallet from bytes
func FromBytes(value []byte) (*Bep2Wallet, error) {
	var wallet Bep2Wallet
	d := gob.NewDecoder(bytes.NewReader(value))
	if err := d.Decode(&wallet); nil != err {
		return nil, errors.Wrapf(err, "fail to decode wallet")
	}
	return &wallet, nil
}

// ToBytes gob serilize the object to bytes array
func (bw Bep2Wallet) ToBytes() ([]byte, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	if err := e.Encode(bw); err != nil {
		return nil, errors.Wrapf(err, "fail to encode wallet, asset symbol: %s", bw.AssetSymbol)
	}
	return b.Bytes(), nil
}

func (bw Bep2Wallet) String() string {
	b := strings.Builder{}
	b.WriteString(" PublicAddress:" + bw.PublicAddress)
	b.WriteString(" PrivateKey:" + bw.PrivateKey)
	b.WriteString(" AssetSymbol:" + bw.AssetSymbol)
	b.WriteString(" Mnemonic:" + bw.Mnemonic)
	return b.String()
}
