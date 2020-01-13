package types

import "gitlab.com/thorchain/thornode/common"

type Keygens struct {
	Height  uint64           `json:"height"`
	Keygens []common.PubKeys `json:"keygens"`
}

func NewKeygens(height uint64) Keygens {
	return Keygens{
		Height:  height,
		Keygens: make([]common.PubKeys, 0),
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
