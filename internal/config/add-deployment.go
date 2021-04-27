/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsAddDeployment struct {
	Network   string   `flag:"network" info:"Network name used for deployment"`
	Account   string   `flag:"account" info:"Account name used for deployment"`
	Contracts []string `flag:"contract" info:"Name of the contract to be deployed"`
}

var addDeploymentFlags = flagsAddDeployment{}

var AddDeploymentCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "deployment",
		Short:   "Add deployment to configuration",
		Example: "flow config add deployment",
		Args:    cobra.NoArgs,
	},
	Flags: &addDeploymentFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		p, err := project.Load(globalFlags.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("configuration does not exists")
		}
		conf := p.Config()

		deployData, flagsProvided, err := flagsToDeploymentData(addDeploymentFlags)
		if err != nil {
			return nil, err
		}

		if !flagsProvided {
			deployData = output.NewDeploymentPrompt(conf.Networks, conf.Accounts, conf.Contracts)
		}

		deployment := config.StringToDeployment(
			deployData["network"].(string),
			deployData["account"].(string),
			deployData["contracts"].([]string),
		)

		conf.Deployments.AddOrUpdate(deployment)

		err = p.SaveDefault()
		if err != nil {
			return nil, err
		}

		return &ConfigResult{
			result: "deploy added",
		}, nil
	},
}

func init() {
	AddDeploymentCommand.AddToParent(AddCmd)
}

func flagsToDeploymentData(flags flagsAddDeployment) (map[string]interface{}, bool, error) {
	if flags.Network == "" && flags.Account == "" && len(flags.Contracts) == 0 {
		return nil, false, nil
	}

	if flags.Network == "" {
		return nil, true, fmt.Errorf("network name must be provided")
	} else if flags.Account == "" {
		return nil, true, fmt.Errorf("account name must be provided")
	} else if len(flags.Contracts) == 0 {
		return nil, true, fmt.Errorf("at least one contract name must be provided")
	}

	return map[string]interface{}{
		"network":   flags.Network,
		"account":   flags.Account,
		"contracts": flags.Contracts,
	}, true, nil
}
