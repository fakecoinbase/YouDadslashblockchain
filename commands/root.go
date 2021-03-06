package commands

import (
	"github.com/YouDad/blockchain/global"
	"github.com/YouDad/blockchain/log"
	"github.com/spf13/cobra"
)

var (
	logLevel uint
)

func init() {
	RootCmd.PersistentFlags().StringVar(&global.Port, "port", "",
		"The port is service's portid at localhost network")
	RootCmd.MarkPersistentFlagRequired("port")
	RootCmd.PersistentFlags().UintVarP(&logLevel, "verbose", "v", 2,
		"Verbose information 0~3")
	RootCmd.PersistentFlags().IntVarP(&global.MaxGroupNum, "max_group_number", "g", 4,
		"Group hash's max group number, must bigger than 0")
}

var RootCmd = &cobra.Command{
	Use:   "blockchain",
	Short: "Blockchain coin Application",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Register(logLevel, global.Port)
	},
}
