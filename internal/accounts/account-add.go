package accounts

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/pkg/flowcli/util"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsAdd struct {
	KeySigAlgo  string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm"`
	KeyHashAlgo string `default:"SHA3_256" flag:"hash-algo" info:"Hashing algorithm"`
	KeyIndex    int    `flag:"index" info:"Account key index"`
	KeyHex      string `flag:"private-key" info:"Private key in hex format"`
	Overwrite   bool   `flag:"overwrite,o" info:"Overwrite an existing account"`
}

var addFlags = flagsAdd{}

var AddCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "add <name> <address>",
		Short:   "Add account by name to config",
		Example: `flow accounts add alice 18d6e0586b0a20c5 --private-key=11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7`,
		Args:    cobra.ExactArgs(2),
	},
	Flags: &addFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		account, err := services.Accounts.Add(
			args[0], // name
			args[1], // address
			addFlags.KeySigAlgo,
			addFlags.KeyHashAlgo,
			addFlags.KeyIndex,
			addFlags.KeyHex,
			addFlags.Overwrite,
			util.ConfigPath,
		)
		if err != nil {
			return nil, err
		}

		return &AccountAddResult{account}, nil
	},
}

// AccountAddResult is the result from the "flow accounts add" command.
type AccountAddResult struct {
	*project.Account
}

func (r *AccountAddResult) JSON() interface{} {
	return r
}

func (r *AccountAddResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Address\t 0x%s\n", r.Account.Address())
	fmt.Fprintf(writer, "Hash Algo\t %s\n", r.Account.DefaultKey().HashAlgo())
	fmt.Fprintf(writer, "Signature Algo\t %s\n", r.Account.DefaultKey().SigAlgo())
	fmt.Fprintf(writer, "Key Index\t %d\n", r.Account.DefaultKey().Index())

	writer.Flush()

	return b.String()
}

func (r *AccountAddResult) Oneliner() string {
	return fmt.Sprintf(
		"Address: 0x%s, Hash Algo: %s, Sig Algo: %s",
		r.Address(),
		r.Account.DefaultKey().HashAlgo(),
		r.Account.DefaultKey().SigAlgo(),
	)
}
