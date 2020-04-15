package tss

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
	// this is the key we are using to test TSS keysign result in BTC chain
	fmt.Println(base64.StdEncoding.EncodeToString(msg))
	if poolPubKey == "thorpub1addwnpepqdvw4jxzzpr4ulvrm045k967x5mfr2hcjl9wud692jvztxmx7td2szeyl8l" {
		return getSignature("Xln0CTTl5PPqm+O9Icj39cHnxueclo6M/oDNrrlhCww=", "dRTOBzO2+1BFRoo9OzsmnPR3OveiWx28oifNgLcAMbE=")
	}
	return nil, nil
}
