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
	if poolPubKey == "thorpub1addwnpepqts24euwrgly2vtez3zdvusmk6u3cwf8leuzj8m4ynvmv5cst7us2vltqrh" {
		return getSignature("8RrEI1OG07hiGgRA82/Vfjw5U6OWE6YMrwE9lL5kflM=", "FbNxmjunwFNvwdpzKawW0XbbWSKED8rkt0se4S5KJJk=")
	}
	return nil, nil
}
