package main

import (
	"fmt"
	"github.com/xuperchain/xuperchain/core/crypto/account"
	"log"
	"os"
	"strings"
	"github.com/spf13/cobra"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
)

// AccountRestoreCommand restore account by mnemonic
type AccountRestoreCommand struct {
	cli *Cli
	cmd *cobra.Command

	outputdir  string
	mnemonic   string
	lang       string
	cryptoType string
}

// NewAccountRestoreCommand
func NewAccountRestoreCommand(cli *Cli) *cobra.Command {
	c := new(AccountRestoreCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "restore",
		Short: "restore account (address,public key,private key) by mnemonic",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.restoreAccount()
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *AccountRestoreCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.outputdir, "output", "o", "", "output directory,  If not specified output directory, it will print the results to the console ")
	c.cmd.Flags().StringVarP(&c.mnemonic, "mnemonic", "m", "", "mnemonic,such as --mnemonic/-m \"累 铝 影 予 细 碳 永 诺 您 态 肯 宫 烘 充 揭 情 勇 梁\"")
	c.cmd.Flags().StringVar(&c.lang, "lang", "zh", "mnemonic language, zh|en")
}

// restore account by mnemonic
func (c *AccountRestoreCommand) restoreAccount() error {
	mnemonic := c.mnemonic
	outputdir := c.outputdir
	langstr := c.lang
	lang := 1
	switch langstr {
	case "zh":
		lang = 1
	case "en":
		lang = 2
	default:
		return fmt.Errorf("bad lang:%s use zh|en instead", langstr)
	}
	c.cryptoType = c.cli.RootOptions.CryptoType
	cryptoClient, cryptoErr := crypto_client.CreateCryptoClient(c.cryptoType)
	if cryptoErr != nil {
		return fmt.Errorf("fail to create crypto client, err:%s", cryptoErr)
	}

	ecdsaAccount, err := cryptoClient.RetrieveAccountByMnemonic(mnemonic, lang)
	if err != nil {
		return fmt.Errorf("restore account by mnemonic failed:%s", err)
	}

	if outputdir != "" {
		if _, err := os.Stat(outputdir); err == nil {
			return fmt.Errorf("output directory exists, abort")
		}
		if err := os.MkdirAll(outputdir, os.ModePerm); nil != err {
			return fmt.Errorf("failed to create output dir before restore account:%s", err)
		}
		if strings.LastIndex(outputdir, "/") != len([]rune(outputdir))-1 {
			outputdir = outputdir + "/"
		}
		err = account.WriteToFile(outputdir+"mnemonic", []byte(ecdsaAccount.Mnemonic))
		if err != nil {
			log.Printf("Export mnemonic file failed, the err is %v", err)
			return err
		}
		err = account.WriteToFile(outputdir+"private.key", []byte(ecdsaAccount.JSONPrivateKey))
		if err != nil {
			log.Printf("Export private key file failed, the err is %v", err)
			return err
		}
		err = account.WriteToFile(outputdir+"public.key", []byte(ecdsaAccount.JSONPublicKey))
		if err != nil {
			log.Printf("Export public key file failed, the err is %v", err)
			return err
		}
		err = account.WriteToFile(outputdir+"address", []byte(ecdsaAccount.Address))
		if err != nil {
			log.Printf("Export address file failed, the err is %v", err)
			return err
		}
		fmt.Printf("export account in : %s\n", outputdir)
	} else {
		fmt.Printf("address : %s\n", ecdsaAccount.Address)
		fmt.Printf("privateKey : %s\n", ecdsaAccount.JSONPrivateKey)
		fmt.Printf("publicKey : %s\n", ecdsaAccount.JSONPublicKey)
		fmt.Printf("mnemonic : %s\n", ecdsaAccount.Mnemonic)
	}
	return nil
}
