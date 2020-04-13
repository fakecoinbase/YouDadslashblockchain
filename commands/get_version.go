package commands

import (
	"github.com/YouDad/blockchain/api"
	"github.com/YouDad/blockchain/log"
	"github.com/YouDad/blockchain/p2p"
	"github.com/spf13/cobra"
)

var GetVersionCmd = &cobra.Command{
	Use:   "get_version",
	Short: "Print version information the blocks of the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		version, err := api.GetVersion()
		if err != nil {
			p2p.Register(Port)
			go p2p.StartServer("")
			<-p2p.ServerReady
			version, err = api.GetVersion()
			if err != nil {
				log.Errln(err)
			}
		}
		log.Infof("Version :%d\n", version.Version)
		log.Infof("Height  :%d\n", version.Height)
		log.Infof("RootHash:%x\n", version.RootHash)
	},
}