package commands

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/YouDad/blockchain/api"
	"github.com/YouDad/blockchain/core"
	"github.com/YouDad/blockchain/global"
	"github.com/YouDad/blockchain/log"
	"github.com/YouDad/blockchain/network"
	"github.com/YouDad/blockchain/types"
	"github.com/YouDad/blockchain/wallet"
)

var speed uint

func init() {
	MiningCmd.Flags().StringVar(&global.Address, "address", "", "node's coin address")
	MiningCmd.MarkFlagRequired("address")
	MiningCmd.Flags().UintVar(&speed, "speed", 1, "mining speed: 0~100, 100 is 100% pc speed")
	MiningCmd.Flags().IntVar(&global.GroupNum, "group", 1, "process group of number")
}

var MiningCmd = &cobra.Command{
	Use:   "mining",
	Short: "Start a mining node with ID specified in port.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infoln("Starting node", global.Port)
		network.Register()
		core.Register(speed)

		// 挖矿
		go func() {
			<-network.ServerReady
			if !wallet.ValidateAddress(global.Address) {
				log.Errln("Address is not valid")
			}

			group := global.GetGroup()
			for {
				var txns [][]*types.Transaction
				for i := 0; i < global.GroupNum; i++ {
					txns = append(txns, []*types.Transaction{core.NewCoinbaseTxn(global.Address)})
					txns[i] = append(txns[i], global.GetMempool(group+i).GetTxns()...)
				}

				for {
					log.Debugf("core.MineBlocks group: %d, number: %d {{{{{{{{", group, global.GroupNum)
					newBlocks, err := core.MineBlocks(txns, group, global.GroupNum)
					if err != core.ErrBlockchainChange {
						log.Warn(err)
					}

					if err != nil {
						time.Sleep(10 * time.Second)
						break
					}

					for _, newBlock := range newBlocks {
						api.CallSelfBlock(newBlock)
					}

					for _, newBlock := range newBlocks {
						// 生成交易的Merkle树
						tree := core.NewTxnMerkleTree(newBlock.Txns)

						for txnIndex, txn := range newBlock.Txns {
							groups := make(map[int]bool)
							for _, out := range txn.Vout {
								groups[global.GetGroupByPubKeyHash(out.PubKeyHash)] = true
							}
							delete(groups, global.GetGroup())

							for group := range groups {
								api.GossipRelayTxn(global.GetGroup(), group,
									newBlock.Height, tree.FindPath(txnIndex), txn)
							}
						}
					}

					log.Debugf("core.MineBlocks group: %d, number: %d }}}}}}}}", group, global.GroupNum)
					time.Sleep(time.Second / 2)
				}
			}
		}()

		network.StartServer(api.Sync)
	},
}
