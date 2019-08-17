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
	log "github.com/rs/zerolog/log"

	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

func Sign(txIn types.TxIn) types.StdTx {
	var (
		msg   types.Msg
		stdTx types.StdTx
	)

	for _, txItem := range txIn.TxArray {
		txHash := types.TxHash{
			Request: txItem.Tx,
			Status:  "incomplete",
			Txhash:  txItem.Tx,
			Memo:    txItem.Memo,
			Coins:   txItem.Coins,
			Sender:  txItem.Sender,
		}

		msg.Value.TxHashes = append(msg.Value.TxHashes, txHash)
	}

	msg.Type = "swapservice/MsgSetTxHash"
	msg.Value.Signer = ctypes.RuneAddress
	stdTx.Type = "cosmos-sdk/StdTx"
	stdTx.Value.Msg = append(stdTx.Value.Msg, msg)

	// @todo What should the gas be set to?
	stdTx.Value.Fee.Gas = "200000"

	payload, _ := json.Marshal(stdTx)
	file, _ := ioutil.TempFile("/tmp", "tx")

	err := ioutil.WriteFile(file.Name(), payload, 0644)
	if err != nil {
		log.Fatal().Msgf("Error while writing to a temporary file: %v", err)
	}

	sign := fmt.Sprintf("/bin/echo %v | sscli tx sign %v --from %v", ctypes.SignerPasswd, file.Name(), ctypes.RuneAddress)
	out, err := exec.Command("/bin/bash", "-c", sign).Output()
	if err != nil {
		log.Fatal().Msgf("Error while signing the request: %v %v", err, sign)
	}
	defer os.Remove(file.Name())

	var signed types.StdTx
	_ = json.Unmarshal(out, &signed)

	return signed
}

func Send(signed types.StdTx) {
	var setTx types.SetTx
	setTx.Mode = "sync"
	setTx.Tx.Msg = signed.Value.Msg
	setTx.Tx.Fee = signed.Value.Fee
	setTx.Tx.Signatures = signed.Value.Signatures
	setTx.Tx.Memo = signed.Value.Memo

	sendSetTx, _ := json.Marshal(setTx)

	uri := url.URL{
		Scheme: "http",
		Host:   ctypes.ChainHost,
		Path:   "/txs",
	}

	resp, err := http.Post(uri.String(), "application/json", bytes.NewBuffer(sendSetTx))
	if err != nil {
		log.Error().Msgf("Error %v", err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var commit types.Commit
	_ = json.Unmarshal(body, &commit)
	log.Info().Msgf("Received a TxHash of %v from the Statechain", commit.TxHash)
}
