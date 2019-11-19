package types

// ValidatorMeta save the meta data used for validator rotation
type ValidatorMeta struct {
	Nominated                     NodeAccounts
	RotateAtBlockHeight           int64 // indicate when we will update the validator set
	RotateWindowOpenAtBlockHeight int64
	Queued                        NodeAccounts
	LeaveQueue                    NodeAccounts // nodes that have requested to leave
	LeaveOpenWindow               int64
	LeaveProcessAt                int64
	Ragnarok                      bool // execute Ragnarok protocol at LeaveProcessAt block height
}
