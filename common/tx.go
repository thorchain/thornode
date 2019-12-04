package common

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxID string
type TxIDs []TxID

var BlankTxID = TxID("0000000000000000000000000000000000000000000000000000000000000000")

func NewTxID(hash string) (TxID, error) {
	switch len(hash) {
	case 64:
		// do nothing
	case 66: // ETH check
		if !strings.HasPrefix(hash, "0x") {
			err := fmt.Errorf("TxID Error: Must be 66 characters (got %d)", len(hash))
			return TxID(""), err
		}
	default:
		err := fmt.Errorf("TxID Error: Must be 64 characters (got %d)", len(hash))
		return TxID(""), err
	}

	return TxID(strings.ToUpper(hash)), nil
}

func (tx TxID) Equals(tx2 TxID) bool {
	return strings.EqualFold(tx.String(), tx2.String())
}

func (tx TxID) IsEmpty() bool {
	return strings.TrimSpace(tx.String()) == ""
}

func (tx TxID) String() string {
	return string(tx)
}

type Tx struct {
	ID          TxID    `json:"id"`
	Chain       Chain   `json:"chain"`
	FromAddress Address `json:"from_address"`
	ToAddress   Address `json:"to_address"`
	Coins       Coins   `json:"coins"`
	Gas         Gas     `json:"gas"`
	Memo        string  `json:"memo"`
}

type Txs []Tx

func GetRagnarokTx(chain Chain) Tx {
	return Tx{
		Chain:       chain,
		ID:          BlankTxID,
		FromAddress: RagnarokAddr,
		ToAddress:   RagnarokAddr,
		Coins: Coins{
			// used for ragnarok, so doesn't really matter
			NewCoin(RuneAsset(), sdk.OneUint()),
		},
		Memo: "Ragnarok",
	}
}

func NewTx(txID TxID, from Address, to Address, coins Coins, gas Gas, memo string) Tx {
	var chain Chain
	for _, coin := range coins {
		chain = coin.Asset.Chain
		break
	}
	return Tx{
		ID:          txID,
		Chain:       chain,
		FromAddress: from,
		ToAddress:   to,
		Coins:       coins,
		Gas:         gas,
		Memo:        memo,
	}
}
func (tx Tx) String() string {
	return fmt.Sprintf("%s: %s ==> %s (Memo: %s) %s", tx.ID, tx.FromAddress, tx.ToAddress, tx.Memo, tx.Coins)
}

func (tx Tx) IsEmpty() bool {
	return tx.ID.IsEmpty()
}

func (tx Tx) IsValid() error {
	if tx.ID.IsEmpty() {
		return fmt.Errorf("Tx ID cannot be empty")
	}
	if tx.FromAddress.IsEmpty() {
		return fmt.Errorf("From address cannot be empty")
	}
	if tx.ToAddress.IsEmpty() {
		return fmt.Errorf("To address cannot be empty")
	}
	if tx.Chain.IsEmpty() {
		return fmt.Errorf("Chain cannot be empty")
	}
	if len(tx.Coins) == 0 {
		return fmt.Errorf("Must have at least 1 coin")
	}
	if err := tx.Coins.IsValid(); err != nil {
		return err
	}
	if len(tx.Gas) == 0 {
		return fmt.Errorf("Must have at least 1 gas coin")
	}
	if err := tx.Gas.IsValid(); err != nil {
		return err
	}
	return nil
}
