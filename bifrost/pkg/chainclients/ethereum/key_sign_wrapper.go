package ethereum

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	etypes "github.com/ethereum/go-ethereum/core/types"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

var (
	vByte     = byte(35)
	eipSigner = etypes.NewEIP155Signer(big.NewInt(1))
)

func getETHPrivateKey(key crypto.PrivKey) (*ecdsa.PrivateKey, error) {
	privKey, ok := key.(secp256k1.PrivKeySecp256k1)
	if !ok {
		return nil, errors.New("invalid private key type")
	}
	return ecrypto.ToECDSA(privKey[:])
}

// KeySignWrapper is a wrap of private key and also tss instance
type KeySignWrapper struct {
	privKey       *ecdsa.PrivateKey
	pubKey        common.PubKey
	tssKeyManager tss.ThorchainKeyManager
	logger        zerolog.Logger
}

func (w *KeySignWrapper) GetPrivKey() *ecdsa.PrivateKey {
	return w.privKey
}

func (w *KeySignWrapper) GetPubKey() common.PubKey {
	return w.pubKey
}

func (w *KeySignWrapper) Sign(tx *etypes.Transaction, poolPubKey common.PubKey, signerPubKeys common.PubKeys) ([]byte, error) {
	var err error
	var sig []byte
	if w.pubKey.Equals(poolPubKey) {
		sig, err = w.sign(tx)
	} else {
		sig, err = w.multiSig(tx, poolPubKey.String(), signerPubKeys)
		if sig != nil {
			sig = append(sig, vByte)
		}
	}
	if err != nil {
		return nil, err
	}
	newTx, err := tx.WithSignature(eipSigner, sig)
	if err != nil {
		return nil, err
	}
	enc, err := newTx.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func (w *KeySignWrapper) sign(tx *etypes.Transaction) ([]byte, error) {
	hash := eipSigner.Hash(tx)
	sig, err := ecrypto.Sign(hash[:], w.privKey)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (w *KeySignWrapper) multiSig(tx *etypes.Transaction, poolPubKey string, signerPubKeys common.PubKeys) ([]byte, error) {
	pk, err := sdk.GetAccPubKeyBech32(poolPubKey)
	if err != nil {
		return nil, fmt.Errorf("fail to get pub key: %w", err)
	}
	hash := eipSigner.Hash(tx)
	sig, err := w.tssKeyManager.RemoteSign(hash[:], poolPubKey, signerPubKeys)
	if err != nil || sig == nil {
		return nil, fmt.Errorf("fail to TSS sign: %w", err)
	}

	if pk.VerifyBytes(hash[:], sig) {
		w.logger.Info().Msg("we can successfully verify the bytes")
	} else {
		w.logger.Error().Msg("Oops! we cannot verify the bytes")
	}
	return sig, nil
}
