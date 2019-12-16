package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
)

type status string
type ObservedTxIndex common.TxIDs

const (
	Incomplete status = "incomplete"
	Done       status = "done"
	Reverted   status = "reverted"
)

// Meant to track if THORNode have processed a specific tx
type ObservedTx struct {
	Tx             common.Tx        `json:"tx"`
	Status         status           `json:"status"`
	OutHashes      common.TxIDs     `json:"out_hashes"` // completed chain tx hash. This is a slice to track if we've "double spent" an input
	BlockHeight    sdk.Uint         `json:"block_height"`
	Signers        []sdk.AccAddress `json:"signers"` // trust accounts saw this tx
	ObservedPubKey common.PubKey    `json:"observed_pub_key"`
}

type ObservedTxs []ObservedTx

func NewObservedTx(tx common.Tx, height sdk.Uint, pk common.PubKey) ObservedTx {
	return ObservedTx{
		Tx:             tx,
		Status:         Incomplete,
		BlockHeight:    height,
		ObservedPubKey: pk,
	}
}

func (tx ObservedTx) Valid() error {
	if err := tx.Tx.IsValid(); err != nil {
		return err
	}
	// Ideally memo should not be empty, THORNode check it here, but if
	// THORNode check it empty here, then the tx will be rejected by thorchain
	// given that THORNode is not going to refund the transaction, thus THORNode
	// will allow ObservedTx has empty to get into thorchain. Thorchain will refund user
	if tx.BlockHeight.IsZero() {
		return errors.New("block height can't be zero")
	}
	if tx.ObservedPubKey.IsEmpty() {
		return errors.New("observed pool pubkey is empty")
	}
	return nil
}

func (tx ObservedTx) Empty() bool {
	return tx.Tx.IsEmpty()
}

func (tx ObservedTx) Equals(tx2 ObservedTx) bool {
	if !tx.Tx.Equals(tx2.Tx) {
		return false
	}
	if !tx.ObservedPubKey.Equals(tx2.ObservedPubKey) {
		return false
	}
	return true
}

func (tx ObservedTx) String() string {
	return tx.Tx.String()
}

// HasSigned - check if given address has signed
func (tx ObservedTx) HasSigned(signer sdk.AccAddress) bool {
	for _, sign := range tx.Signers {
		if sign.Equals(signer) {
			return true
		}
	}
	return false
}

func (tx *ObservedTx) Sign(signer sdk.AccAddress) {
	if !tx.HasSigned(signer) {
		tx.Signers = append(tx.Signers, signer)
	}
}

func (tx *ObservedTx) SetDone(hash common.TxID, numOuts int) {
	for _, done := range tx.OutHashes {
		if done.Equals(hash) {
			return
		}
	}
	tx.OutHashes = append(tx.OutHashes, hash)
	if tx.IsDone(numOuts) {
		tx.Status = Done
	}
}

func (tx *ObservedTx) IsDone(numOuts int) bool {
	if len(tx.OutHashes) >= numOuts {
		return true
	}
	return false
}

type ObservedTxVoter struct {
	TxID    common.TxID `json:"tx_id"`
	Height  int64       `json:"height"`
	Txs     ObservedTxs `json:"in_tx"`   // copies of tx in by various observers.
	Actions []TxOutItem `json:"actions"` // outbound txs set to be sent
	OutTxs  common.Txs  `json:"out_txs"` // observed outbound transactions
}

type ObservedTxVoters []ObservedTxVoter

func NewObservedTxVoter(txID common.TxID, txs []ObservedTx) ObservedTxVoter {
	return ObservedTxVoter{
		TxID: txID,
		Txs:  txs,
	}
}

func (tx ObservedTxVoter) Valid() error {
	if tx.TxID.IsEmpty() {
		return errors.New("Cannot have an empty tx id")
	}

	for _, in := range tx.Txs {
		if err := in.Valid(); err != nil {
			return err
		}
	}

	return nil
}

func (tx ObservedTxVoter) Key() common.TxID {
	return tx.TxID
}

func (tx ObservedTxVoter) String() string {
	return tx.TxID.String()
}

func (tx *ObservedTxVoter) AddOutTx(in common.Tx) {
	for _, t := range tx.OutTxs {
		if in.ID.Equals(t.ID) {
			return
		}
	}
	tx.OutTxs = append(tx.OutTxs, in)
	for i := range tx.Txs {
		tx.Txs[i].SetDone(in.ID, len(tx.Actions))
	}
}

func (tx *ObservedTxVoter) IsDone() bool {
	return len(tx.Actions) <= len(tx.OutTxs)
}

func (tx *ObservedTxVoter) Add(observedTx ObservedTx, signer sdk.AccAddress) {
	// check if this signer has already signed, no take backs allowed
	for _, transaction := range tx.Txs {
		for _, siggy := range transaction.Signers {
			if siggy.Equals(signer) {
				return
			}
		}
	}

	for i := range tx.Txs {
		if tx.Txs[i].Equals(observedTx) {
			tx.Txs[i].Sign(signer)
			return
		}
	}

	observedTx.Signers = []sdk.AccAddress{signer}
	tx.Txs = append(tx.Txs, observedTx)
}

func (tx ObservedTxVoter) HasConensus(nodeAccounts NodeAccounts) bool {
	for _, txIn := range tx.Txs {
		var count int
		for _, signer := range txIn.Signers {
			if nodeAccounts.IsTrustAccount(signer) {
				count += 1
			}
		}
		if HasMajority(count, len(nodeAccounts)) {
			return true
		}
	}

	return false
}

func (tx ObservedTxVoter) GetTx(nodeAccounts NodeAccounts) ObservedTx {
	for _, txIn := range tx.Txs {
		var count int
		for _, signer := range txIn.Signers {
			if nodeAccounts.IsTrustAccount(signer) {
				count += 1
			}
		}
		if HasMajority(count, len(nodeAccounts)) {
			return txIn
		}
	}

	return ObservedTx{}
}
