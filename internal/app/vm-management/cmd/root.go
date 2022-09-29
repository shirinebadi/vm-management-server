package cmd

import (
	"github.com/shirinebadi/vm-management-server/internal/app/vm-management/cmd/server"
	"github.com/shirinebadi/vm-management-server/internal/app/vm-management/config"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var root = &cobra.Command{
		Use: "vm-management-server",
	}
	cfg := config.Init()

	server.Register(root, cfg)

	return root
}
