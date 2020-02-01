package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gitlab.com/thorchain/thornode/common"
)

// KeygenBlock represent the TSS Keygen in a block
// if you wonder why there is a Keygens which is a slice of Keygen , that is because thorchain can potentially have to trigger multiple TSS Keygen in one block
// for example multiple Asgard, also when later on Yggdrasil start to use TSS as well
type KeygenBlock struct {
	Height  int64    `json:"height"`
	Keygens []Keygen `json:"keygens"`
}

// NewKeygenBlock create a new KeygenBlock
func NewKeygenBlock(height int64) KeygenBlock {
	return KeygenBlock{
		Height: height,
	}
}

// IsEmpty determinate whether KeygenBlock is empty
func (k KeygenBlock) IsEmpty() bool {
	return len(k.Keygens) == 0 && k.Height == 0
}

// Contains will go through the keygen items and find out whether the given keygen already exist in the block or not
func (k KeygenBlock) Contains(keygen Keygen) bool {
	for _, item := range k.Keygens {
		if item.ID.Equals(keygen.ID) {
			return true
		}
	}
	return false
}

type KeygenType byte

const (
	UnknownKeygen KeygenType = iota
	// AsgardKeygen obviously
	AsgardKeygen
	YggdrasilKeygen
)

// String implement fmt.Stringer
func (kt KeygenType) String() string {
	switch kt {
	case UnknownKeygen:
		return "unknown"
	case AsgardKeygen:
		return "asgard"
	case YggdrasilKeygen:
		return "yggdrasil"
	}
	return ""
}

func GetKeygenTypeFromString(t string) KeygenType {
	switch {
	case strings.EqualFold(t, "asgard"):
		return AsgardKeygen
	case strings.EqualFold(t, "yggdrasil"):
		return YggdrasilKeygen
	}
	return UnknownKeygen
}

// MarshalJSON marshal PoolStatus to JSON in string form
func (kt KeygenType) MarshalJSON() ([]byte, error) {
	return json.Marshal(kt.String())
}

// UnmarshalJSON convert string form back to PoolStatus
func (kt *KeygenType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*kt = GetKeygenTypeFromString(s)
	return nil
}

// Keygen one keygen
type Keygen struct {
	ID      common.TxID    `json:"id"`
	Type    KeygenType     `json:"type"`
	Members common.PubKeys `json:"members"`
}

// NewKeygen create a new instance of Keygen
func NewKeygen(height int64, members common.PubKeys, keygenType KeygenType) (Keygen, error) {
	// sort the members
	sort.SliceStable(members, func(i, j int) bool {
		return members[i].String() < members[j].String()
	})
	id, err := getKeygenID(height, members, keygenType)
	if err != nil {
		return Keygen{}, fmt.Errorf("fail to create new keygen: %w", err)
	}
	return Keygen{
		ID:      id,
		Members: members,
		Type:    keygenType,
	}, nil
}

// getKeygenID will create ID based on the pub keys
func getKeygenID(height int64, members common.PubKeys, keygenType KeygenType) (common.TxID, error) {
	sb := strings.Builder{}
	sb.WriteString(strconv.FormatInt(height, 10))
	sb.WriteString(keygenType.String())
	for _, m := range members {
		sb.WriteString(m.String())
	}
	h := sha256.New()
	_, err := h.Write([]byte(sb.String()))
	if err != nil {
		return "", fmt.Errorf("fail to write to hash: %w", err)
	}

	return common.TxID(hex.EncodeToString(h.Sum(nil))), nil
}

// IsEmpty check whether there are any keys in the keygen
func (k Keygen) IsEmpty() bool {
	return len(k.Members) == 0 && len(k.ID) == 0
}

// Valid is to check whether the keygen members are valid
func (k Keygen) Valid() error {
	if k.Type == UnknownKeygen {
		return errors.New("unknown keygen")
	}
	return k.Members.Valid()
}

// String implement of fmt.Stringer
func (k Keygen) String() string {
	return fmt.Sprintf(`id:%s
	type:%s
	member:%+v
`, k.ID, k.Type, k.Members)
}
