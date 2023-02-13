// Copyright 2023 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0
package autorestore

import (
	"fmt"

	"github.com/redpanda-data/redpanda/src/go/rpk/pkg/api/admin"
	"github.com/redpanda-data/redpanda/src/go/rpk/pkg/config"
	"github.com/redpanda-data/redpanda/src/go/rpk/pkg/out"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// TODO: add flag to get detailed info (e.g. topic download counts).
func newStatusCommand(fs afero.Fs) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Fetch the status of the autorestore process",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			p := config.ParamsFromCommand(cmd)
			cfg, err := p.Load(fs)
			out.MaybeDie(err, "unable to load config: %v", err)

			client, err := admin.NewClient(fs, cfg)
			out.MaybeDie(err, "unable to initialize admin client: %v", err)

			status, err := client.PollAutomatedRecoveryStatus(cmd.Context())
			out.MaybeDie(err, "unable to fetch auto-restore status: %v", err)

			fmt.Printf("Auto-restore status: %s\n", status.State)
		},
	}

	return cmd
}
