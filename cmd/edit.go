package cmd

import (
	"github.com/spf13/cobra"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"os"
)

var editCmd = &cobra.Command{
	Use:   "edit <NAME>",
	Short: "Edit a cluster state",
	Long: `Use this command to edit an state.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			logger.Critical("Too many arguments.")
			os.Exit(1)
		} else {
			ao.Name = args[0]
		}

		err := RunEdit(ao)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}

	},
}

func init() {
	RootCmd.AddCommand(editCmd)
}

func RunEdit(options *ApplyOptions) error {

}