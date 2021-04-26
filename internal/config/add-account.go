package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsAddAccount struct {
	Name     string `flag:"name" info:"Name for the account"`
	Address  string `flag:"address" info:"Account address"`
	KeyIndex string `default:"0" flag:"key-index" info:"Account key index"`
	SigAlgo  string `default:"ECDSA_P256" flag:"sig-algo" info:"Account key signature algorithm"`
	HashAlgo string `default:"SHA3_256" flag:"hash-algo" info:"Account hash used for the digest"`
	Key      string `flag:"key" info:"Account private key"`
}

var addAccountFlags = flagsAddAccount{}

var AddAccountCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "account",
		Short:   "Add account to configuration",
		Example: "flow config add account",
		Args:    cobra.NoArgs,
	},
	Flags: &addAccountFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		accountData, flagsProvided, err := flagsToAccountData(addAccountFlags)
		if err != nil {
			return nil, err
		}

		if !flagsProvided {
			accountData = output.NewAccountPrompt()
		}

		account, err := config.StringToAccount(
			accountData["name"],
			accountData["address"],
			accountData["keyIndex"],
			accountData["sigAlgo"],
			accountData["hashAlgo"],
			accountData["key"],
		)
		if err != nil {
			return nil, err
		}

		err = services.Config.AddAccount(*account)
		if err != nil {
			return nil, err
		}

		return &ConfigResult{
			result: "account added",
		}, nil

	},
}

func flagsToAccountData(flags flagsAddAccount) (map[string]string, bool, error) {
	if flags.Name == "" && flags.Address == "" && flags.Key == "" {
		return nil, false, nil
	}

	if flags.Name == "" {
		return nil, true, fmt.Errorf("name must be provided")
	} else if flags.Address == "" {
		return nil, true, fmt.Errorf("address must be provided")
	} else if flags.Key == "" {
		return nil, true, fmt.Errorf("key must be provided")
	}

	return map[string]string{
		"name":     flags.Name,
		"address":  flags.Address,
		"keyIndex": flags.KeyIndex,
		"sigAlgo":  flags.SigAlgo,
		"hashAlg":  flags.HashAlgo,
		"key":      flags.Key,
	}, true, nil
}
