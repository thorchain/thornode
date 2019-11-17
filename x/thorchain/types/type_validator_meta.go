package types

// ValidatorMeta save the meta data used for validator rotation
type ValidatorMeta struct {
	Nominated                     NodeAccounts
	RotateAtBlockHeight           int64 // indicate when we will update the validator set
	RotateWindowOpenAtBlockHeight int64
	Queued                        NodeAccounts
}
