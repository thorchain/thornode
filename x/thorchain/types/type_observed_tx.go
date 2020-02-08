package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
)

type (
	status          string
	ObservedTxIndex common.TxIDs
)

const (
	Incomplete status = "incomplete"
	Done       status = "done"
	Reverted   status = "reverted"
)

type ObservedSigner struct {
	Address   sdk.AccAddress `json:"address"`
	PubKey    common.PubKey  `json:"pubkey"`
	Signature []byte         `json:"signature"`
}

func NewObservedSigner(addr sdk.AccAddress, pk common.PubKey, sig []byte) ObservedSigner {
	return ObservedSigner{
		Address:   addr,
		PubKey:    pk,
		Signature: sig,
	}
}

// Meant to track if THORNode have processed a specific tx
type ObservedTx struct {
	Tx             common.Tx        `json:"tx"`
	Status         status           `json:"status"`
	OutHashes      common.TxIDs     `json:"out_hashes"` // completed chain tx hash. This is a slice to track if we've "double spent" an input
	BlockHeight    int64            `json:"block_height"`
	Signers        []ObservedSigner `json:"signers"` // node keys of node account saw this tx
	ObservedPubKey common.PubKey    `json:"observed_pub_key"`
}

type ObservedTxs []ObservedTx

func NewObservedTx(tx common.Tx, height int64, pk common.PubKey) ObservedTx {
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
	if tx.BlockHeight == 0 {
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

func (tx ObservedTx) Verify(signer ObservedSigner) bool {
	addr, err := signer.PubKey.GetThorAddress()
	if err != nil {
		return false
	}
	if !addr.Equals(signer.Address) {
		return false
	}
	pk, err := signer.PubKey.CryptoPubKey()
	if err != nil {
		return false
	}
	ok, err := tx.Tx.Verify(pk, signer.Signature)
	if err != nil || !ok {
		return false
	}
	return ok
}

func (tx ObservedTx) AddressHasSigned(addr sdk.AccAddress) bool {
	for _, sign := range tx.Signers {
		if sign.Address.Equals(addr) {
			return true
		}
	}
	return false
}

// HasSigned - check if given address has signed
func (tx ObservedTx) HasSigned(signer ObservedSigner) bool {
	for _, sign := range tx.Signers {
		if sign.Address.Equals(signer.Address) {
			pk, err := sign.PubKey.CryptoPubKey()
			if err != nil {
				continue
			}
			ok, err := tx.Tx.Verify(pk, sign.Signature)
			if err != nil || !ok {
				continue
			}
			return true
		}
	}
	return false
}

func (tx *ObservedTx) Sign(signer ObservedSigner) {
	if !tx.HasSigned(signer) && tx.Verify(signer) {
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
	TxID         common.TxID `json:"tx_id"`
	Height       int64       `json:"height"`
	ProcessedIn  bool        `json:"processed_in"`  // used to track if has been processed txin
	ProcessedOut bool        `json:"processed_out"` // used to track if has been processed txout
	Txs          ObservedTxs `json:"in_tx"`         // copies of tx in by various observers.
	Actions      []TxOutItem `json:"actions"`       // outbound txs set to be sent
	OutTxs       common.Txs  `json:"out_txs"`       // observed outbound transactions
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

// String implement fmt.Stringer
func (tx ObservedTxVoter) String() string {
	return tx.TxID.String()
}

// matchActionItem is to check the given outboundTx again the list of actions , return true of the outboundTx matched any of the actions
func (tx ObservedTxVoter) matchActionItem(outboundTx common.Tx) bool {
	for _, toi := range tx.Actions {
		// note: Coins.Contains will match amount as well
		if strings.EqualFold(toi.Memo, outboundTx.Memo) &&
			toi.ToAddress.Equals(outboundTx.ToAddress) &&
			toi.Chain.Equals(outboundTx.Chain) &&
			outboundTx.Coins.Contains(toi.Coin) {
			return true
		}
	}
	return false
}

// AddOutTx trying to add the outbound tx into OutTxs ,
// return value false indicate the given outbound tx doesn't match any of the actions items , node account should be slashed for a malicious tx
// true indicated the outbound tx matched an action item , and it has been added into internal OutTxs
func (tx *ObservedTxVoter) AddOutTx(in common.Tx) bool {
	if !tx.matchActionItem(in) {
		// no action item match the outbound tx
		return false
	}

	for _, t := range tx.OutTxs {
		if in.ID.Equals(t.ID) {
			return true
		}
	}
	tx.OutTxs = append(tx.OutTxs, in)
	for i := range tx.Txs {
		tx.Txs[i].SetDone(in.ID, len(tx.Actions))
	}
	return true
}

func (tx *ObservedTxVoter) IsDone() bool {
	return len(tx.Actions) <= len(tx.OutTxs)
}

func (tx *ObservedTxVoter) Add(observedTx ObservedTx, signer ObservedSigner) {
	// check if this signer has already signed, no take backs allowed
	for _, transaction := range tx.Txs {
		if transaction.HasSigned(signer) {
			return
		}
	}

	// find a matching observed tx, and sign it
	for i := range tx.Txs {
		if tx.Txs[i].Equals(observedTx) {
			tx.Txs[i].Sign(signer)
			return
		}
	}

	// could not find a matching observed tx to sign, add a new one
	observedTx.Sign(signer)
	tx.Txs = append(tx.Txs, observedTx)
}

func (tx ObservedTxVoter) HasConsensus(nodeAccounts NodeAccounts) bool {
	for _, txIn := range tx.Txs {
		var count int
		for _, signer := range txIn.Signers {
			if nodeAccounts.IsNodeKeys(signer.Address) && txIn.Verify(signer) {
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
			if nodeAccounts.IsNodeKeys(signer.Address) {
				count += 1
			}
		}
		if HasMajority(count, len(nodeAccounts)) {
			return txIn
		}
	}

	return ObservedTx{}
}
