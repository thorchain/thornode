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

	types "gitlab.com/thorchain/bepswap/observe/x/observer/types"
)

type StateChain struct {
	ChainHost string
	RuneAddress string
}

func NewStateChain(chainHost, runeAddress string) *StateChain {
	return &StateChain{
		ChainHost: chainHost,
		RuneAddress: runeAddress,
	}
}

func (s *StateChain) Send(inTx types.InTx) {
	var (
		msg types.Msg
		stdTx types.StdTx
	)

	for _, txItem := range inTx.TxArray {
		var coins []types.Coins
		coins = append(coins, txItem.Coins)

		txHash := types.TxHash{Request: txItem.Tx,
			Status: "incomplete",
			Txhash: txItem.Tx,
			Memo: txItem.Memo,
			Coins: coins,
			Sender: txItem.Sender}
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
		log.Fatal().Msgf("Error: %v", err)
	}

	sign := fmt.Sprintf("/bin/echo %v | sscli tx sign %v --from %v", os.Getenv("SIGNER_PASSWD"), file.Name(), s.RuneAddress)
	out, err := exec.Command("/bin/bash", "-c", sign).Output()
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
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

	// Retry until we get a successful reply and log the commit hash.
	resp, _ := http.Post(uri.String(), "application/json", bytes.NewBuffer(sendSetTx))
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var commit types.Commit
	json.Unmarshal(body, &commit)

	log.Info().Msgf("Commited hash: %v", commit.TxHash)
}
