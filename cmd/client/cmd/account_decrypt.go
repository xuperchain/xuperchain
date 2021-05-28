/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 *
 * Usage: Decrypt account from a encrypted private key file.
 *        ./xchain-cli account decrypt --output data/tmpkey --key private.key
 */

package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/xuperchain/crypto/core/account"
	"github.com/xuperchain/crypto/core/hdwallet/key"
)

// AccountDecryptCommand decrypt account struct
type AccountDecryptCommand struct {
	cli *Cli
	cmd *cobra.Command

	output string
	file   string
}

// NewAccountDecryptCommand new decrypt account command
func NewAccountDecryptCommand(cli *Cli) *cobra.Command {
	t := &AccountDecryptCommand{}
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt plain address,public and private key from encrypted private key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.decrypt(ctx)
		},
	}
	t.addFlags()

	return t.cmd
}

func (c *AccountDecryptCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.output, "output", "o", "data/keys", "output directory")
	c.cmd.Flags().StringVar(&c.file, "key", "private.key", "The encrypted private key file path")
}

func (c *AccountDecryptCommand) decrypt(ctx context.Context) error {
	// check parameters
	if _, err := os.Stat(c.output); err == nil {
		return fmt.Errorf("output directory exists, abort")
	}

	if c.cli.RootOptions.Crypto != "default" {
		return fmt.Errorf("only support default crypto plugin by now")
	}

	// get encrypted key
	content, err := ioutil.ReadFile(c.file)
	if err != nil {
		fmt.Println("failed to read encrypted private key")
		return err
	}
	encKey, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		fmt.Println("failed to base64 decode private key")
		return err
	}

	// read password
	validate := func(input string) error {
		// suggest at least 6 chars
		if len(input) < 4 {
			return errors.New("Password must at least 4 characters")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Password",
		Validate: validate,
		Mask:     '*',
	}
	passwd, err := prompt.Run()
	if err != nil {
		fmt.Println("failed to get password")
		return err
	}

	// decrypt account
	sk, err := key.GetBinaryEcdsaPrivateKeyFromString(string(encKey), passwd)
	if err != nil {
		fmt.Println("failed to restore private key")
		return err
	}

	// get private key
	eccPrivkey, err := account.GetEcdsaPrivateKeyFromJson(sk)
	if err != nil {
		fmt.Println("failed to restore private key, please check your key or password")
		return err
	}

	// get public key
	pk, err := account.GetEcdsaPublicKeyJsonFormat(eccPrivkey)
	if err != nil {
		fmt.Println("failed to get public key")
		return err
	}

	// get address
	addr, err := account.GetAddressFromPublicKey(&eccPrivkey.PublicKey)
	if err != nil {
		fmt.Println("failed to get address")
		return err
	}

	// print address
	fmt.Println("decrypted address:", addr)
	err = c.saveAccount([]byte(addr), []byte(pk), sk)
	if err != nil {
		fmt.Println("failed to save account")
		return err
	}
	fmt.Println("account decrypted successfully, account info saved at ", c.output)
	return nil
}

// save account info into output folder
func (c *AccountDecryptCommand) saveAccount(addr, pk, sk []byte) error {
	if err := os.MkdirAll(c.output, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output dir before restore account:%s", err)
	}

	if err := ioutil.WriteFile(c.output+"/address", addr, 0644); err != nil {
		return fmt.Errorf("failed to save address:%s", err)
	}

	if err := ioutil.WriteFile(c.output+"/private.key", sk, 0644); err != nil {
		return fmt.Errorf("failed to save private key:%s", err)
	}

	if err := ioutil.WriteFile(c.output+"/public.key", pk, 0644); err != nil {
		return fmt.Errorf("failed to save public key:%s", err)
	}
	return nil
}
