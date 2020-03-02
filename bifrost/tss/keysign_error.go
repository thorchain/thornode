package tss

import (
	"fmt"

	tssCommon "gitlab.com/thorchain/tss/go-tss/common"
)

// KeysignError is a custom error create to include which party to blame
type KeysignError struct {
	Blame tssCommon.Blame
}

// NewKeysignError create a new instance of KeysignError
func NewKeysignError(blame tssCommon.Blame) KeysignError {
	return KeysignError{
		Blame: blame,
	}
}

// Error implement error interface
func (k KeysignError) Error() string {
	return fmt.Sprintf("fail to complete TSS keysign,reason:%s, culprit:%+v", k.Blame.FailReason, k.Blame.BlameNodes)
}
