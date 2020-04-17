package tss

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/btcd/btcec"
	"github.com/tendermint/tendermint/crypto"

	"gitlab.com/thorchain/tss/go-tss/keysign"
	tss "gitlab.com/thorchain/tss/go-tss/tss"

	"gitlab.com/thorchain/thornode/common"
)

// KeySign is a proxy between signer and TSS
type KeySign struct {
	logger zerolog.Logger
	server *tss.TssServer
}

// NewKeySign create a new instance of KeySign
func NewKeySign(server *tss.TssServer) (*KeySign, error) {
	return &KeySign{
		server: server,
		logger: log.With().Str("module", "tss_signer").Logger(),
	}, nil
}

// GetPrivKey THORNode don't actually have any private key , but just return something
func (s *KeySign) GetPrivKey() crypto.PrivKey {
	return nil
}

func (s *KeySign) GetAddr() ctypes.AccAddress {
	return nil
}

// ExportAsMnemonic THORNode don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsMnemonic() (string, error) {
	return "", nil
}

// ExportAsPrivateKey THORNode don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsPrivateKey() (string, error) {
	return "", nil
}

// ExportAsKeyStore THORNode don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsKeyStore(password string) (*keys.EncryptedKeyJSON, error) {
	return nil, nil
}

func (s *KeySign) makeSignature(msg tx.StdSignMsg, poolPubKey string, signerPubKeys common.PubKeys) (sig tx.StdSignature, err error) {
	var stdSignature tx.StdSignature
	pk, err := sdk.GetAccPubKeyBech32(poolPubKey)
	if err != nil {
		return stdSignature, fmt.Errorf("fail to get pub key: %w", err)
	}
	signPack, err := s.RemoteSign(msg.Bytes(), poolPubKey, signerPubKeys)
	if err != nil {
		return stdSignature, fmt.Errorf("fail to TSS sign: %w", err)
	}

	if signPack == nil {
		return stdSignature, nil
	}
	if pk.VerifyBytes(msg.Bytes(), signPack) {
		s.logger.Info().Msg("we can successfully verify the bytes")
	} else {
		s.logger.Error().Msg("Oops! we cannot verify the bytes")
	}

	return tx.StdSignature{
		AccountNumber: msg.AccountNumber,
		Sequence:      msg.Sequence,
		PubKey:        pk,
		Signature:     signPack,
	}, nil
}

func (s *KeySign) Sign(msg tx.StdSignMsg) ([]byte, error) {
	return nil, nil
}

func (s *KeySign) SignWithPool(msg tx.StdSignMsg, poolPubKey common.PubKey, signerPubKeys common.PubKeys) ([]byte, error) {
	sig, err := s.makeSignature(msg, poolPubKey.String(), signerPubKeys)
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

func (s *KeySign) RemoteSign(msg []byte, poolPubKey string, signerPubKeys common.PubKeys) ([]byte, error) {
	if len(msg) == 0 {
		return nil, nil
	}
	hashedMsg := crypto.Sha256(msg)
	encodedMsg := base64.StdEncoding.EncodeToString(hashedMsg)
	rResult, sResult, err := s.toLocalTSSSigner(poolPubKey, encodedMsg, signerPubKeys)
	if err != nil {
		return nil, fmt.Errorf("fail to tss sign: %w", err)
	}

	if len(rResult) == 0 && len(sResult) == 0 {
		// this means the node tried to do keygen , however this node has not been chosen to take part in the keysign committee
		return nil, nil
	}
	s.logger.Debug().Str("R", rResult).Str("S", sResult).Msg("tss result")
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
func (s *KeySign) toLocalTSSSigner(poolPubKey, sendmsg string, signerPubKeys common.PubKeys) (string, string, error) {
	tssMsg := keysign.Request{
		PoolPubKey: poolPubKey,
		Message:    sendmsg,
	}
	for _, k := range signerPubKeys {
		tssMsg.SignerPubKeys = append(tssMsg.SignerPubKeys, k.String())
	}
	s.logger.Debug().Str("payload", fmt.Sprintf("PoolPubKey: %s, Message: %s, Signers: %+v", tssMsg.PoolPubKey, tssMsg.Message, tssMsg.SignerPubKeys)).Msg("msg to tss Local node")

	keySignResp, err := s.server.KeySign(tssMsg)
	if err != nil {
		return "", "", fmt.Errorf("fail to send request to local TSS node: %w", err)
	}

	// 1 means success,2 means fail , 0 means NA
	if keySignResp.Status == 1 && keySignResp.Blame.IsEmpty() {
		return keySignResp.R, keySignResp.S, nil
	}

	// Blame need to be passed back to thorchain , so as thorchain can use the information to slash relevant node account
	return "", "", NewKeysignError(keySignResp.Blame)
}
