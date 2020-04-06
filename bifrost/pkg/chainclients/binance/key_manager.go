package binance

// This file is largely a copy from https://github.com/binance-chain/go-sdk/blob/515ede99ef1b6c7b5eaf27c67fa7984d98be58e3/keys/keys.go.
// Needed a manual way to set `privKey` which the original source doesn't give a means to do so

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/rs/zerolog"
	"github.com/pkg/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/btcd/btcec"

	"gitlab.com/thorchain/thornode/common"
	tssp "gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/tss/go-tss/tss"
	"gitlab.com/thorchain/tss/go-tss/keysign"
)

type KeyManager struct {
	privKey  crypto.PrivKey
	addr     ctypes.AccAddress
	pubkey   common.PubKey
	mnemonic string
	logger   zerolog.Logger
	server   *tss.TssServer
}

func (m *KeyManager) Pubkey() common.PubKey {
	return m.pubkey
}

func (m *KeyManager) ExportAsMnemonic() (string, error) {
	if m.mnemonic == "" {
		return "", fmt.Errorf("This key manager is not recover from mnemonic or anto generated ")
	}
	return m.mnemonic, nil
}

func (m *KeyManager) ExportAsPrivateKey() (string, error) {
	secpPrivateKey, ok := m.privKey.(secp256k1.PrivKeySecp256k1)
	if !ok {
		return "", fmt.Errorf(" Only PrivKeySecp256k1 key is supported ")
	}
	return hex.EncodeToString(secpPrivateKey[:]), nil
}

func (m *KeyManager) ExportAsKeyStore(password string) (*keys.EncryptedKeyJSON, error) {
	// Do nothing
	return nil, nil
}

func (m *KeyManager) Sign(msg tx.StdSignMsg) ([]byte, error) {
	sig, err := m.makeSignature(msg)
	if err != nil {
		return nil, err
	}
	newTx := tx.NewStdTx(msg.Msgs, []tx.StdSignature{sig}, msg.Memo, msg.Source, msg.Data)
	bz, err := tx.Cdc.MarshalBinaryLengthPrefixed(&newTx)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func (m *KeyManager) GetPrivKey() crypto.PrivKey {
	return m.privKey
}

func (m *KeyManager) GetAddr() ctypes.AccAddress {
	return m.addr
}

func (m *KeyManager) makeSignature(msg tx.StdSignMsg) (sig tx.StdSignature, err error) {
	sigBytes, err := m.privKey.Sign(msg.Bytes())
	if err != nil {
		return
	}
	return tx.StdSignature{
		AccountNumber: msg.AccountNumber,
		Sequence:      msg.Sequence,
		PubKey:        m.privKey.PubKey(),
		Signature:     sigBytes,
	}, nil
}

func (m *KeyManager) makeMultiSig(msg tx.StdSignMsg, poolPubKey string, signerPubKeys common.PubKeys) (sig tx.StdSignature, err error) {
	var stdSignature tx.StdSignature
	pk, err := sdk.GetAccPubKeyBech32(poolPubKey)
	if err != nil {
		return stdSignature, fmt.Errorf("fail to get pub key: %w", err)
	}
	signPack, err := m.remoteSign(msg.Bytes(), poolPubKey, signerPubKeys)
	if err != nil {
		return stdSignature, fmt.Errorf("fail to TSS sign: %w", err)
	}

	if signPack == nil {
		return stdSignature, nil
	}

	if pk.VerifyBytes(msg.Bytes(), signPack) {
		m.logger.Info().Msg("we can successfully verify the bytes")
	} else {
		m.logger.Error().Msg("Oops! we cannot verify the bytes")
	}

	return tx.StdSignature{
		AccountNumber: msg.AccountNumber,
		Sequence:      msg.Sequence,
		PubKey:        pk,
		Signature:     signPack,
	}, nil
}

func (m *KeyManager) SignWithPool(msg tx.StdSignMsg, poolPubKey common.PubKey, signerPubKeys common.PubKeys) ([]byte, error) {
	sig, err := m.makeMultiSig(msg, poolPubKey.String(), signerPubKeys)
	if err != nil {
		return nil, err
	}
	if len(sig.Signature) == 0 {
		return nil, errors.New("fail to make signature")
	}
	newTx := tx.NewStdTx(msg.Msgs, []tx.StdSignature{sig}, msg.Memo, msg.Source, msg.Data)
	bz, err := tx.Cdc.MarshalBinaryLengthPrefixed(&newTx)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func (m *KeyManager) remoteSign(msg []byte, poolPubKey string, signerPubKeys common.PubKeys) ([]byte, error) {
	if len(msg) == 0 {
		return nil, nil
	}
	encodedMsg := base64.StdEncoding.EncodeToString(msg)
	rResult, sResult, err := m.toLocalTSSSigner(poolPubKey, encodedMsg, signerPubKeys)
	if err != nil {
		return nil, fmt.Errorf("fail to tss sign: %w", err)
	}

	if len(rResult) == 0 && len(sResult) == 0 {
		// this means the node tried to do keygen , however this node has not been chosen to take part in the keysign committee
		return nil, nil
	}
	m.logger.Debug().Str("R", rResult).Str("S", sResult).Msg("tss result")
	data, err := getSignature(rResult, sResult)
	if err != nil {
		return nil, fmt.Errorf("fail to decode tss signature: %w", err)
	}

	return data, nil
}

func getSignature(r, s string) ([]byte, error) {
	rBytes, err := base64.StdEncoding.DecodeString(r)
	if err != nil {
		return nil, err
	}
	sBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	R := new(big.Int).SetBytes(rBytes)
	S := new(big.Int).SetBytes(sBytes)
	N := btcec.S256().N
	halfOrder := new(big.Int).Rsh(N, 1)
	// see: https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
	if S.Cmp(halfOrder) == 1 {
		S.Sub(N, S)
	}

	// Serialize signature to R || S.
	// R, S are padded to 32 bytes respectively.
	rBytes = R.Bytes()
	sBytes = S.Bytes()

	sigBytes := make([]byte, 64)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(rBytes):32], rBytes)
	copy(sigBytes[64-len(sBytes):64], sBytes)
	return sigBytes, nil
}

// toLocalTSSSigner will send the request to local signer
func (m *KeyManager) toLocalTSSSigner(poolPubKey, sendmsg string, signerPubKeys common.PubKeys) (string, string, error) {
	tssMsg := keysign.Request{
		PoolPubKey: poolPubKey,
		Message:    sendmsg,
	}
	for _, k := range signerPubKeys {
		tssMsg.SignerPubKeys = append(tssMsg.SignerPubKeys, k.String())
	}
	m.logger.Debug().Str("payload", fmt.Sprintf("PoolPubKey: %s, Message: %s, Signers: %+v", tssMsg.PoolPubKey, tssMsg.Message, tssMsg.SignerPubKeys)).Msg("msg to tss Local node")

	keySignResp, err := m.server.KeySign(tssMsg)
	if err != nil {
		return "", "", errors.Wrapf(err, "fail to send request to local TSS node")
	}

	// 1 means success,2 means fail , 0 means NA
	if keySignResp.Status == 1 && keySignResp.Blame.IsEmpty() {
		return keySignResp.R, keySignResp.S, nil
	}

	// Blame need to be passed back to thorchain , so as thorchain can use the information to slash relevant node account
	return "", "", tssp.NewKeysignError(keySignResp.Blame)
}
