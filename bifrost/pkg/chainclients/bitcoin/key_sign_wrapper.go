package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"gitlab.com/thorchain/txscript"

	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

// KeySignWrapper is a wrap of private key and also tss instance
// it also implement the txscript.Signable interface, and will decide which method to use based on the pubkey
type KeySignWrapper struct {
	privateKey    *btcec.PrivateKey
	pubKey        common.PubKey
	tssKeyManager tss.ThorchainKeyManager
	bridge        *thorclient.ThorchainBridge
	logger        zerolog.Logger
}

// NewKeysignWrapper create a new instance of Keysign Wrapper
func NewKeySignWrapper(privateKey *btcec.PrivateKey, bridge *thorclient.ThorchainBridge, tssKeyManager tss.ThorchainKeyManager) (*KeySignWrapper, error) {
	pubKey, err := GetBech32AccountPubKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("fail to get the pubkey: %w", err)
	}
	return &KeySignWrapper{
		privateKey:    privateKey,
		pubKey:        pubKey,
		tssKeyManager: tssKeyManager,
		bridge:        bridge,
		logger:        log.With().Str("module", "keysign_wrapper").Logger(),
	}, nil
}

// GetBech32AccountPubKey convert the given private key to
func GetBech32AccountPubKey(key *btcec.PrivateKey) (common.PubKey, error) {
	buf := key.PubKey().SerializeCompressed()
	var pk secp256k1.PubKeySecp256k1
	copy(pk[:], buf)
	return common.NewPubKeyFromCrypto(pk)
}

// GetSignable based on the given poolPubKey
func (w *KeySignWrapper) GetSignable(poolPubKey common.PubKey) txscript.Signable {
	if w.pubKey.Equals(poolPubKey) {
		return txscript.NewPrivateKeySignable(w.privateKey)
	}
	s, err := NewTssSignable(poolPubKey, w.bridge, w.tssKeyManager)
	if err != nil {
		w.logger.Err(err).Msg("fail to create tss signable")
		return nil
	}
	return s
}
