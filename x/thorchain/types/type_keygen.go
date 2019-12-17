package types

import (
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
)

type Keygen []common.PubKey

func (k Keygen) Valid() error {
	if len(k) < 2 {
		return errors.New("keygen cannot contain less than two pub keys")
	}
	for _, pk := range k {
		if _, err := common.NewPubKey(pk.String()); err != nil {
			return err
		}
	}
	return nil
}

func (k Keygen) Contains(pk common.PubKey) bool {
	for _, p := range k {
		if p.Equals(pk) {
			return true
		}
	}
	return false
}

// String implement stringer interface
func (k Keygen) String() string {
	strs := make([]string, len(k))
	for i := range k {
		strs[i] = k[i].String()
	}
	return strings.Join(strs, ", ")
}

type Keygens struct {
	Height  uint64   `json:"height"`
	Keygens []Keygen `json:"keygens"`
}

func NewKeygens(height uint64) Keygens {
	return Keygens{
		Height:  height,
		Keygens: make([]Keygen, 0),
	}
}

func (k Keygens) IsEmpty() bool {
	return len(k.Keygens) == 0
}

func (k Keygens) Valid() error {
	for _, key := range k.Keygens {
		if err := key.Valid(); err != nil {
			return err
		}
	}
	return nil
}
