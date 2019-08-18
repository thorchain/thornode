package statechain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os/user"
	"path/filepath"
	"strconv"

	http "github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"

	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"gitlab.com/thorchain/bepswap/observe/config"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"

	"gitlab.com/thorchain/bepswap/common"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

func Sign(txIns []stypes.TxIn, signer sdk.AccAddress) (authtypes.StdTx, error) {
	name := config.SignerName
	msg := stypes.NewMsgSetTxIn(txIns, signer)
	stdTx := authtypes.NewStdTx(
		[]sdk.Msg{msg},                   // messages
		authtypes.NewStdFee(200000, nil), // fee
		nil,                              // signatures
		"",                               // memo
	)

	// TODO: make keys directory configurable
	usr, err := user.Current()
	if err != nil {
		return stdTx, err
	}
	sscliDir := filepath.Join(usr.HomeDir, ".sscli")

	// Get keys database
	kb, err := keys.NewKeyBaseFromDir(sscliDir)
	if err != nil {
		return stdTx, err
	}

	// Get signer user information
	info, err := kb.Get(name)
	if err != nil {
		return stdTx, err
	}

	// Get account number and sequence via rest API
	uri := url.URL{
		Scheme: "http",
		Host:   config.ChainHost,
		Path:   fmt.Sprintf("/auth/accounts/%s", info.GetAddress()),
	}

	resp, err := http.Get(uri.String())
	if err != nil {
		return stdTx, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return stdTx, err
	}

	var baseAccount types.BaseAccount
	err = json.Unmarshal(body, &baseAccount)
	if err != nil {
		return stdTx, err
	}
	base := baseAccount.Value
	acctNumber, _ := strconv.ParseInt(base.AccountNumber, 10, 64)
	seq, _ := strconv.ParseInt(base.Sequence, 10, 64)

	stdMsg := authtypes.StdSignMsg{
		ChainID:       "sschain", // TODO : make this configurable
		AccountNumber: uint64(acctNumber),
		Sequence:      uint64(seq),
		Fee:           stdTx.Fee,
		Msgs:          stdTx.GetMsgs(),
		Memo:          stdTx.GetMemo(),
	}

	sig, err := authtypes.MakeSignature(kb, name, config.SignerPasswd, stdMsg)
	if err != nil {
		return stdTx, err
	}

	signedStdTx := authtypes.NewStdTx(
		stdTx.GetMsgs(),
		stdTx.Fee,
		[]authtypes.StdSignature{sig},
		stdTx.GetMemo(),
	)

	return signedStdTx, nil
}

func Send(signed authtypes.StdTx, mode types.TxMode) (common.TxID, error) {
	var noTxID = common.TxID("")
	if !mode.IsValid() {
		return noTxID, fmt.Errorf("Transaction Mode (%s) is invalid", mode.String())
	}

	var setTx types.SetTx
	setTx.Mode = mode.String()
	setTx.Tx.Msg = signed.Msgs
	setTx.Tx.Fee = signed.Fee
	setTx.Tx.Signatures = signed.Signatures
	setTx.Tx.Memo = signed.Memo

	sendSetTx, err := json.Marshal(setTx)
	if err != nil {
		return noTxID, err
	}

	uri := url.URL{
		Scheme: "http",
		Host:   config.ChainHost,
		Path:   "/txs",
	}

	resp, err := http.Post(uri.String(), "application/json", bytes.NewBuffer(sendSetTx))
	if err != nil {
		return noTxID, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return noTxID, err
	}

	var commit types.Commit
	err = json.Unmarshal(body, &commit)
	if err != nil {
		return noTxID, err
	}

	log.Info().Msgf("Received a TxHash of %v from the Statechain", commit.TxHash)
	return common.NewTxID(commit.TxHash)
}
