package observer

import (
	"os"
	"fmt"
	"bytes"
	"net/url"
	"os/exec"
	"io/ioutil"
	"encoding/json"

	log "github.com/rs/zerolog/log"
	http "github.com/hashicorp/go-retryablehttp"

	"gitlab.com/thorchain/bepswap/observe/common/types"
)

type StateChain struct {
	ChainHost string
	RuneAddress string
	TxChan chan []byte
}

func NewStateChain(chainHost, runeAddress string, txChan chan []byte) *StateChain {
	return &StateChain{
		ChainHost: chainHost,
		RuneAddress: runeAddress,
		TxChan: txChan,
	}
}

func (s *StateChain) Send(inTx types.InTx) {
	var (
		msg types.Msg
		stdTx types.StdTx
	)

	for _, txItem := range inTx.TxArray {
		txHash := types.TxHash{
			Request: txItem.Tx,
			Status: "incomplete",
			Txhash: txItem.Tx,
			Memo: txItem.Memo,
			Coins: txItem.Coins,
			Sender: txItem.Sender,
		}

		msg.Type = "swapservice/MsgSetTxHash"
		msg.Value.TxHashes = append(msg.Value.TxHashes, txHash)
	}

	msg.Value.Signer = s.RuneAddress
	stdTx.Type = "cosmos-sdk/StdTx"
	stdTx.Value.Msg = append(stdTx.Value.Msg, msg)

	// @todo What should the gas be set to?
	stdTx.Value.Fee.Gas = "200000"

	payload, _ := json.Marshal(stdTx)
	file, _ := ioutil.TempFile("/tmp", "tx")
	
	err := ioutil.WriteFile(file.Name(), payload, 0644)
	if err != nil {
		log.Fatal().Msgf("%s Error: %v", LogPrefix(), err)
	}

	sign := fmt.Sprintf("/bin/echo %v | sscli tx sign %v --from %v", os.Getenv("SIGNER_PASSWD"), file.Name(), s.RuneAddress)
	out, err := exec.Command("/bin/bash", "-c", sign).Output()
	if err != nil {
		log.Fatal().Msgf("%s gError: %v %v", LogPrefix(), err, sign)
	}
	defer os.Remove(file.Name())

	var signed types.StdTx
	json.Unmarshal(out, &signed)

	var setTx types.SetTx
	setTx.Mode = "sync"
	setTx.Tx.Msg = signed.Value.Msg
	setTx.Tx.Fee = signed.Value.Fee
	setTx.Tx.Signatures = signed.Value.Signatures
	setTx.Tx.Memo = signed.Value.Memo

	sendSetTx, _ := json.Marshal(setTx)

	uri := url.URL{
		Scheme: "http",
		Host: s.ChainHost,
		Path: "/txs",
	}

	// Retry until we get a successful reply and log the reply.
	log.Info().Msgf("%s Sending to the StateChain %v", string(sendSetTx))
	resp, _ := http.Post(uri.String(), "application/json", bytes.NewBuffer(sendSetTx))
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var commit types.Commit
	json.Unmarshal(body, &commit)

	log.Info().Msgf("%s Received the following response from StateChain: %v", LogPrefix(), commit)

	s.TxChan <- []byte(commit.TxHash)
}
