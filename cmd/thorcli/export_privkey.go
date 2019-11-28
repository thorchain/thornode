package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func exportPrivateKeyForTSS() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tss <name>",
		Short: "Export private keys that will be used for TSS",
		Long:  "Export a private key from the local keybase base64 encoded format",
		Args:  cobra.ExactArgs(1),
		RunE:  runExportPrivateKey,
	}
	return cmd
}
func runExportPrivateKey(cmd *cobra.Command, args []string) error {
	kb, err := keys.NewKeyBaseFromHomeFlag()
	if err != nil {
		return err
	}
	buf := bufio.NewReader(cmd.InOrStdin())
	decryptPassword, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
	if err != nil {
		return err
	}
	priKey, err := kb.ExportPrivateKeyObject(args[0], decryptPassword)
	if nil != err {
		return err
	}
	pk, ok := priKey.(secp256k1.PrivKeySecp256k1)
	if !ok {
		return errors.New("invalid private key")
	}
	hexPrivKey := hex.EncodeToString(pk[:])
	cmd.Println(base64.StdEncoding.EncodeToString([]byte(hexPrivKey)))
	return nil
}
