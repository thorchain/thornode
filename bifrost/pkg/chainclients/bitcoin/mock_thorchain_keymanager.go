package bitcoin

import (
	"encoding/base64"
	"fmt"
	"math/big"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/btcsuite/btcd/btcec"
	"github.com/tendermint/tendermint/crypto"

	"gitlab.com/thorchain/thornode/common"
)

// MockThorchainKeymanager is to mock the TSS , so as we could test it
type MockThorchainKeyManager struct {
}

func (k *MockThorchainKeyManager) Sign(tx.StdSignMsg) ([]byte, error) {
	return nil, nil
}
func (k *MockThorchainKeyManager) GetPrivKey() crypto.PrivKey {
	return nil
}
func (k *MockThorchainKeyManager) GetAddr() ctypes.AccAddress {
	return nil
}

func (k *MockThorchainKeyManager) ExportAsMnemonic() (string, error) {
	return "", nil
}
func (k *MockThorchainKeyManager) ExportAsPrivateKey() (string, error) {
	return "", nil
}
func (k *MockThorchainKeyManager) ExportAsKeyStore(password string) (*keys.EncryptedKeyJSON, error) {
	return nil, nil
}
func (k *MockThorchainKeyManager) SignWithPool(msg tx.StdSignMsg, poolPubKey common.PubKey, signerPubKeys common.PubKeys) ([]byte, error) {
	return nil, nil
}

func (k *MockThorchainKeyManager) RemoteSign(msg []byte, poolPubKey string, signerPubKeys common.PubKeys) ([]byte, error) {
	fmt.Println(base64.StdEncoding.EncodeToString(msg))
	// this is the key we are using to test
	if poolPubKey == "thorpub1addwnpepqdvw4jxzzpr4ulvrm045k967x5mfr2hcjl9wud692jvztxmx7td2szeyl8l" {
		return getSignature("ORXh10F2qLJeb/maHLTobieHxNQDp6YIb757nFiZNhQ=", "Y9v/zE5OiZ8EDkpkHNsmkWs1dout5HKi/a/Lr9wJvQY=")
	}
	return nil, nil
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
