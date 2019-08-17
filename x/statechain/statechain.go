package statechain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"

	http "github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	log "github.com/rs/zerolog/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	config "gitlab.com/thorchain/bepswap/observe/config"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"

	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

var execCommand = exec.Command
var msgType = "swapservice/MsgSetTxIn"

// Signs a file using sscli
func signFile(file, name, password string) ([]byte, error) {
	sign := fmt.Sprintf(
		"/bin/echo %s | sscli tx sign %s --from %s",
		config.SignerPasswd,
		file,
		config.SignerName,
	)

	return execCommand("/bin/sh", "-c", sign).Output()
}

func Sign(txIns []stypes.TxIn, signer sdk.AccAddress) (types.StdTx, error) {
	var (
		msg   types.Msg
		stdTx types.StdTx
		err   error
	)

	msg.Type = msgType
	msg.Value = stypes.NewMsgSetTxIn(txIns, signer)
	stdTx.Type = "cosmos-sdk/StdTx"
	stdTx.Value.Msg = append(stdTx.Value.Msg, msg)

	// TODO: What should the gas be set to?
	stdTx.Value.Fee.Gas = "200000"

	payload, err := json.Marshal(stdTx)
	if err != nil {
		return stdTx, err
	}

	// TODO: sign using the cosmos-sdk instead of writing to disk and utilizing
	// the cli. Should see a significant performance boost.
	file, err := ioutil.TempFile("/tmp", "tx")
	if err != nil {
		return stdTx, err
	}

	err = ioutil.WriteFile(file.Name(), payload, 0644)
	if err != nil {
		return stdTx, errors.Wrap(err, "Error while writing to a temporary file")
	}
	defer os.Remove(file.Name())

	out, err := signFile(file.Name(), config.SignerName, config.SignerPasswd)
	if err != nil {
		return stdTx, errors.Wrap(err, "Error while signing the request")
	}

	var signed types.StdTx
	err = json.Unmarshal(out, &signed)
	if err != nil {
		return stdTx, err
	}

	return signed, nil
}

func Send(signed types.StdTx, mode types.TxMode) error {
	if !mode.IsValid() {
		return fmt.Errorf("Transaction Mode (%s) is invalid", mode.String())
	}

	var setTx types.SetTx
	setTx.Mode = mode.String()
	setTx.Tx.Msg = signed.Value.Msg
	setTx.Tx.Fee = signed.Value.Fee
	setTx.Tx.Signatures = signed.Value.Signatures
	setTx.Tx.Memo = signed.Value.Memo

	sendSetTx, err := json.Marshal(setTx)
	if err != nil {
		return err
	}

	uri := url.URL{
		Scheme: "http",
		Host:   config.ChainHost,
		Path:   "/txs",
	}

	resp, err := http.Post(uri.String(), "application/json", bytes.NewBuffer(sendSetTx))
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	var commit types.Commit
	err = json.Unmarshal(body, &commit)
	if err != nil {
		return err
	}

	log.Info().Msgf("Received a TxHash of %v from the Statechain", commit.TxHash)
	return nil
}
