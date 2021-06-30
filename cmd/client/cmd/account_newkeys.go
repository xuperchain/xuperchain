/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	crypto_client "github.com/xuperchain/xupercore/lib/crypto/client"
)

// AccountNewkeysCommand create account addr
type AccountNewkeysCommand struct {
	cli *Cli
	cmd *cobra.Command

	outputdir    string
	strength     uint8
	lang         string
	forceOveride bool
	cryptoType   string
}

// NewAccountNewkeysCommand new addr account cmd
func NewAccountNewkeysCommand(cli *Cli) *cobra.Command {
	c := new(AccountNewkeysCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "newkeys",
		Short: "Create an address with public key and private key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.createAccount()
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *AccountNewkeysCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.outputdir, "output", "o", "./data/keys", "output directory")
	c.cmd.Flags().Uint8Var(&c.strength, "strength", 0, "using mnemonic with specific strength(easy:1 mid:2 hard:3)")
	c.cmd.Flags().StringVar(&c.lang, "lang", "zh", "mnemonic language, zh|en")
	c.cmd.Flags().BoolVarP(&c.forceOveride, "force", "f", false, "Force override existing account files")
}

func (c *AccountNewkeysCommand) createAccount() error {
	if _, err := os.Stat(c.outputdir); err == nil && !c.forceOveride {
		return fmt.Errorf("output directory exists, abort")
	}
	if err := os.MkdirAll(c.outputdir, os.ModePerm); nil != err {
		return fmt.Errorf("failed to create output dir before create account:%s", err)
	}
	c.cryptoType = c.cli.RootOptions.Crypto
	var err error
	if c.strength > 0 {
		// intversion, _ := strconv.ParseInt(xchainversion.Version, 0, 8)
		// version := uint8(intversion)
		err = c.createMnmAccount(c.strength, c.lang)
		if err != nil {
			os.RemoveAll(c.outputdir)
			return err
		}
		return nil
	}
	err = c.createSimpleAccount()
	if err != nil {
		os.RemoveAll(c.outputdir)
		return err
	}
	return nil
}

// create a simple account, without mnemonic
func (c *AccountNewkeysCommand) createSimpleAccount() error {
	// create crypto client
	fmt.Printf("create account using crypto type %s\n", c.cryptoType)
	cryptoClient, cryptoErr := crypto_client.CreateCryptoClient(c.cryptoType)
	if cryptoErr != nil {
		return fmt.Errorf("fail to create crypto client, err:%s", cryptoErr)
	}

	err := cryptoClient.ExportNewAccount(c.outputdir)
	if err != nil {
		return err
	}
	fmt.Printf("create account in %s\n", c.outputdir)
	return nil
}

// create a account with mnemonic
func (c *AccountNewkeysCommand) createMnmAccount(strength uint8, langstr string) error {
	lang := 1
	switch langstr {
	case "zh":
		lang = 1
	case "en":
		lang = 2
	default:
		return fmt.Errorf("bad lang:%s use zh|en instead", langstr)
	}
	// create crypto client
	cryptoClient, cryptoErr := crypto_client.CreateCryptoClient(c.cryptoType)
	if cryptoErr != nil {
		return fmt.Errorf("fail to create crypto client, err:%s", cryptoErr)
	}
	err := cryptoClient.ExportNewAccountWithMnemonic(c.outputdir, lang, strength)
	if err != nil {
		return fmt.Errorf("create new account with mnemonic failed:%s", err)
	}
	fmt.Printf("create account in %s\n", c.outputdir)
	return nil
}
