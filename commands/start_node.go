package commands

import (
	"github.com/spf13/cobra"

	"github.com/YouDad/blockchain/api"
	"github.com/YouDad/blockchain/core"
	"github.com/YouDad/blockchain/log"
	"github.com/YouDad/blockchain/network"
	"github.com/YouDad/blockchain/store"
)

var (
	startNodeAddress string
)

func init() {
	StartNodeCmd.Flags().StringVar(&startNodeAddress, "address", "", "node's coin address")
}

var StartNodeCmd = &cobra.Command{
	Use:   "start_node",
	Short: "Start a node with ID specified in port.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infoln("Starting node", Port)
		network.Register(Port)
		go func() {
			//TODO: mining
		}()
		go func() {
			<-network.ServerReady
			err := network.GetKnownNodes()
			if err != nil {
				log.Warnln(err)
			}

			var bc *core.Blockchain
			if !store.IsDatabaseExists() {
				genesis, err := api.GetGenesis()
				if err != nil {
					log.Errln(err)
				}
				bc = core.CreateBlockchainFromGenesis(genesis)
			} else {
				bc = core.GetBlockchain()
			}
			utxoSet := core.GetUTXOSet()

			genesis := bc.GetGenesis()
			nowHeight := bc.GetHeight()
			height, err := api.SendVersion(nowHeight, genesis.Hash())
			if err == api.RootHashDifferentError {
				log.Warnln(err)
			} else if err == api.VersionDifferentError {
				log.Warnln(err)
			} else if err != nil {
				log.Warnln(err)
			}

			if height > nowHeight {
				blocks := api.GetBlocks(nowHeight+1, height)
				for _, block := range blocks {
					bc.AddBlock(block)
				}
				utxoSet.Reindex()
			}
		}()
		network.StartServer()
	},
}
