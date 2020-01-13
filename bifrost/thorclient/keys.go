package thorclient

import (
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client/keys"
	ckeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/tendermint/tendermint/crypto"
)

const (
	// folder name for thorchain thorcli
	thorchainCliFolderName = `.thorcli`
)

// Keys manages all the keys used by thorchain
type Keys struct {
	chainHomeFolder string
	signerName      string
	password        string // TODO this is a bad way , need to fix it
	signerInfo      ckeys.Info
	kb              ckeys.Keybase
}

// NewKeys create a new instance of keys
func NewKeys(chainHomeFolder, signerName, password string) (*Keys, error) {
	if len(signerName) == 0 {
		return nil, fmt.Errorf("signer name is empty")
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("password is empty")
	}
	kb, err := getKeybase(chainHomeFolder)
	if nil != err {
		return nil, fmt.Errorf("fail to get keybase,err:%w", err)
	}
	signerInfo, err := kb.Get(signerName)
	if nil != err {
		return nil, fmt.Errorf("fail to get signer info,err:%w", err)
	}
	return &Keys{
		chainHomeFolder: chainHomeFolder,
		signerName:      signerName,
		signerInfo:      signerInfo,
		password:        password,
		kb:              kb,
	}, nil
}

// getKeybase will create an instance of Keybase
func getKeybase(thorchainHome string) (ckeys.Keybase, error) {
	cliDir := thorchainHome
	if len(thorchainHome) == 0 {
		usr, err := user.Current()
		if nil != err {
			return nil, fmt.Errorf("fail to get current user,err:%w", err)
		}
		cliDir = filepath.Join(usr.HomeDir, thorchainCliFolderName)
	}
	return keys.NewKeyBaseFromDir(cliDir)
}

// GetSignerInfo return signer info
func (k *Keys) GetSignerInfo() ckeys.Info {
	return k.signerInfo
}

// GetPrivateKey return the private key
func (k *Keys) GetPrivateKey() (crypto.PrivKey, error) {
	return k.kb.ExportPrivateKeyObject(k.signerName, k.password)
}

// GetKeybase return the keybase
func (k *Keys) GetKeybase() ckeys.Keybase {
	return k.kb
}
