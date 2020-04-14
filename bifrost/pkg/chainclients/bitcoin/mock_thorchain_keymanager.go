package bitcoin

import (
	"encoding/base64"
	"fmt"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/tx"
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
	sigBytes := make([]byte, 64)
	copy(sigBytes[32-len(rBytes):32], rBytes)
	copy(sigBytes[64-len(sBytes):64], sBytes)
	return sigBytes, nil
}
