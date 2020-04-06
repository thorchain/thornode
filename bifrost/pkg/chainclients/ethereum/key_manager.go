package ethereum

// This file is largely a copy from https://github.com/binance-chain/go-sdk/blob/515ede99ef1b6c7b5eaf27c67fa7984d98be58e3/keys/keys.go.
// Needed a manual way to set `privKey` which the original source doesn't give a means to do so

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/rs/zerolog"
	"github.com/pkg/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/btcd/btcec"

	"gitlab.com/thorchain/thornode/common"
	tssp "gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/tss/go-tss/keysign"
	"gitlab.com/thorchain/tss/go-tss/tss"
)

var (
	eipSigner = etypes.NewEIP155Signer(big.NewInt(1))
)
	
type KeyManager struct {
	privKey  crypto.PrivKey
	addr     ecommon.Address
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

func (m *KeyManager) GetPrivKey() crypto.PrivKey {
	return m.privKey
}

func (m *KeyManager) GetAddr() ecommon.Address {
	return m.addr
}

func (m *KeyManager) makeSignature(tx *etypes.Transaction) ([]byte, error) {
	hash := eipSigner.Hash(tx)
	sigBytes, err := m.privKey.Sign(hash[:])
	if err != nil {
		return nil, err
	}
	return sigBytes, nil
}

func (m *KeyManager) Sign(tx *etypes.Transaction) ([]byte, error) {
	sig, err := m.makeSignature(tx)
	if err != nil {
		return nil, err
	}
	newTx, err := tx.WithSignature(eipSigner, sig)
	if err != nil {
		return nil, err
	}
	enc, err := rlp.EncodeToBytes(newTx)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func (m *KeyManager) SignWithPool(tx *etypes.Transaction, poolPubKey common.PubKey, signerPubKeys common.PubKeys) ([]byte, error) {
	sig, err := m.makeMultiSig(tx, poolPubKey.String(), signerPubKeys)
	if err != nil {
		return nil, err
	}
	newTx, err := tx.WithSignature(eipSigner, sig)
	if err != nil {
		return nil, err
	}
	enc, err := rlp.EncodeToBytes(newTx)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func (m *KeyManager) makeMultiSig(tx *etypes.Transaction, poolPubKey string, signerPubKeys common.PubKeys) ([]byte, error) {
	pk, err := sdk.GetAccPubKeyBech32(poolPubKey)
	if err != nil {
		return nil, fmt.Errorf("fail to get pub key: %w", err)
	}
	hash := eipSigner.Hash(tx)
	r, s, err := m.remoteSign(hash[:], poolPubKey, signerPubKeys)
	if err != nil {
		return nil, fmt.Errorf("fail to TSS sign: %w", err)
	}

	signPack, err := getSignature(r, s)
	if err != nil {
		return nil, fmt.Errorf("fail to decode tss signature: %w", err)
	}

	if signPack == nil {
		return nil, nil
	}

	if pk.VerifyBytes(hash[:], signPack) {
		m.logger.Info().Msg("we can successfully verify the bytes")
	} else {
		m.logger.Error().Msg("Oops! we cannot verify the bytes")
	}

	return signPack, nil
}

func (m *KeyManager) remoteSign(msg []byte, poolPubKey string, signerPubKeys common.PubKeys) (string, string, error) {
	if len(msg) == 0 {
		return "", "", nil
	}
	encodedMsg := base64.StdEncoding.EncodeToString(msg)
	rResult, sResult, err := m.toLocalTSSSigner(poolPubKey, encodedMsg, signerPubKeys)
	if err != nil {
		return "", "", fmt.Errorf("fail to tss sign: %w", err)
	}

	if len(rResult) == 0 && len(sResult) == 0 {
		// this means the node tried to do keygen , however this node has not been chosen to take part in the keysign committee
		return "", "", nil
	}
	m.logger.Debug().Str("R", rResult).Str("S", sResult).Msg("tss result")
	return rResult, sResult, nil
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

	sigBytes := make([]byte, 66)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(rBytes):32], rBytes)
	copy(sigBytes[64-len(sBytes):64], sBytes)
	// set v as 37
	copy(sigBytes[64:66], []byte{0x02, 0x05})
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
