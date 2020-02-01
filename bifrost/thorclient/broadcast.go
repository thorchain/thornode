package thorclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

// Broadcast Broadcasts tx to thorchain
func (b *ThorchainBridge) Broadcast(stdTx authtypes.StdTx, mode types.TxMode) (common.TxID, error) {
	b.broadcastLock.Lock()
	defer b.broadcastLock.Unlock()

	var noTxID = common.TxID("")
	if !mode.IsValid() {
		return noTxID, errors.New(fmt.Sprintf("transaction Mode (%s) is invalid", mode))
	}
	start := time.Now()
	defer func() {
		b.m.GetHistograms(metrics.SendToThorchainDuration).Observe(time.Since(start).Seconds())
	}()

	blockHeight, err := b.GetBlockHeight()
	if err != nil {
		return noTxID, err
	}
	if blockHeight > b.blockHeight {
		var seqNum uint64
		b.accountNumber, seqNum, err = b.getAccountNumberAndSequenceNumber()
		if err != nil {
			return noTxID, errors.Wrap(err, "fail to get account number and sequence number from thorchain ")
		}
		b.blockHeight = blockHeight
		if seqNum > b.seqNumber {
			b.seqNumber = seqNum
		}
	}

	b.logger.Info().Uint64("account_number", b.accountNumber).Uint64("sequence_number", b.accountNumber).Msg("account info")
	stdMsg := authtypes.StdSignMsg{
		ChainID:       b.cfg.ChainID,
		AccountNumber: b.accountNumber,
		Sequence:      b.seqNumber,
		Fee:           stdTx.Fee,
		Msgs:          stdTx.GetMsgs(),
		Memo:          stdTx.GetMemo(),
	}
	sig, err := authtypes.MakeSignature(b.keys.GetKeybase(), b.cfg.SignerName, b.cfg.SignerPasswd, stdMsg)
	if err != nil {
		b.errCounter.WithLabelValues("fail_sign", "").Inc()
		return noTxID, errors.Wrap(err, "fail to sign the message")
	}

	signed := authtypes.NewStdTx(
		stdTx.GetMsgs(),
		stdTx.Fee,
		[]authtypes.StdSignature{sig},
		stdTx.GetMemo(),
	)

	b.m.GetCounter(metrics.TxToThorchainSigned).Inc()

	var setTx types.SetTx
	setTx.Mode = mode.String()
	setTx.Tx.Msg = signed.Msgs
	setTx.Tx.Fee = signed.Fee
	setTx.Tx.Signatures = signed.Signatures
	setTx.Tx.Memo = signed.Memo
	result, err := b.cdc.MarshalJSON(setTx)
	if err != nil {
		b.errCounter.WithLabelValues("fail_marshal_settx", "").Inc()
		return noTxID, errors.Wrap(err, "fail to marshal settx to json")
	}

	b.logger.Info().Str("payload", string(result)).Msg("post to thorchain")

	body, err := b.post(BroadcastTxsEndpoint, "application/json", bytes.NewBuffer(result))
	if err != nil {
		return noTxID, errors.Wrap(err, "fail to post tx to thorchain")
	}

	// NOTE: we can actually see two different json responses for the same end.
	// This complicates things pretty well.
	// Sample 1: { "height": "0", "txhash": "D97E8A81417E293F5B28DDB53A4AD87B434CA30F51D683DA758ECC2168A7A005", "raw_log": "[{\"msg_index\":0,\"success\":true,\"log\":\"\",\"events\":[{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"set_observed_txout\"}]}]}]", "logs": [ { "msg_index": 0, "success": true, "log": "", "events": [ { "type": "message", "attributes": [ { "key": "action", "value": "set_observed_txout" } ] } ] } ] }
	// Sample 2: { "height": "0", "txhash": "6A9AA734374D567D1FFA794134A66D3BF614C4EE5DDF334F21A52A47C188A6A2", "code": 4, "raw_log": "{\"codespace\":\"sdk\",\"code\":4,\"message\":\"signature verification failed; verify correct account sequence and chain-id\"}" }
	var commit types.Commit
	err = json.Unmarshal(body, &commit)
	if err != nil || len(commit.Logs) == 0 {
		b.errCounter.WithLabelValues("fail_unmarshal_commit", "").Inc()
		b.logger.Error().Err(err).Msg("fail unmarshal commit")

		var badCommit types.BadCommit // since commit doesn't work, lets try bad commit
		err = json.Unmarshal(body, &badCommit)
		if err != nil {
			b.logger.Error().Err(err).Msg("fail unmarshal bad commit")
			return noTxID, errors.Wrap(err, "fail to unmarshal bad commit")
		}

		// check for any failure logs
		if badCommit.Code > 0 {
			err := errors.New(badCommit.Log)
			b.logger.Error().Err(err).Msg("fail to broadcast")
			return badCommit.TxHash, errors.Wrap(err, "fail to broadcast")
		}
	}

	for _, log := range commit.Logs {
		if !log.Success {
			err := errors.New(log.Log)
			b.logger.Error().Err(err).Msg("fail to broadcast")
			return noTxID, errors.Wrap(err, "fail to broadcast")
		}
	}

	b.m.GetCounter(metrics.TxToThorchain).Inc()
	b.logger.Info().Msgf("Received a TxHash of %v from the thorchain", commit.TxHash)

	// increment seqNum
	atomic.AddUint64(&b.seqNumber, 1)

	return commit.TxHash, nil
}
