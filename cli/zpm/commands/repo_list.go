/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package commands

import (
	"github.com/fezz-io/zps/cli"
	"github.com/fezz-io/zps/zpm"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

type ZpmRepoListCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRepoListCommand() *ZpmRepoListCommand {
	cmd := &ZpmRepoListCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "list"
	cmd.Short = "List repositories configured for a ZPS image"
	cmd.Long = "List repositories configured for a ZPS image"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmRepoListCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRepoListCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	repos, err := mgr.RepoList()
	if err != nil {
		z.Fatal(err.Error())
	}
	if repos == nil {
		z.Warn("No configured repositories")
		return nil
	}

	z.Out(columnize.SimpleFormat(repos))

	return nil
}
