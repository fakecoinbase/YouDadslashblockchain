package commands

import (
	"time"

	"github.com/YouDad/blockchain/api"
	"github.com/YouDad/blockchain/global"
	"github.com/YouDad/blockchain/global/mempool"
	"github.com/YouDad/blockchain/log"
	"github.com/YouDad/blockchain/network"
	"github.com/YouDad/blockchain/wallet"
	"github.com/spf13/cobra"
)

var tps int64
var wait int64

func init() {
	SendTestCmd.Flags().IntVar(&global.GroupNum, "group", 1, "process group of number")
	SendTestCmd.Flags().StringVar(&global.Address, "from", "", "Source wallet address")
	SendTestCmd.Flags().Int64Var(&tps, "tps", 10, "send speed, transaction per second")
	SendTestCmd.Flags().Int64Var(&wait, "wait", 60, "the time before sendloop")
	SendTestCmd.MarkFlagRequired("from")
}

var SendTestCmd = &cobra.Command{
	Use:   "send_test",
	Short: "Send A lot of Txns for test, from FROM",
	Run: func(cmd *cobra.Command, args []string) {
		if !wallet.ValidateAddress(global.Address) {
			log.Errln("Sender address is not valid")
		}

		network.Register()
		go network.StartServer(api.Sync)
		time.Sleep(time.Duration(wait) * time.Second)

		<-network.ServerReady
		group := global.GetGroup()
		for {
			for mempool.GetMempoolSize(group) >= 5*int(tps) {
				// log.Traceln("sendtest sleep")
				time.Sleep(time.Second * 3)
			}

			for mempool.GetMempoolSize(group) < 5*int(tps) {
				sendTestTo := string(wallet.NewWallet().GetAddress())
				log.Infoln("SendTest", mempool.GetMempoolSize(group),
					global.Address, sendTestTo)
				err := api.SendCMD(global.Address, sendTestTo, 1)

				if err != nil {
					log.Warnln("SendTest Warn?", err)
				} else {
					log.Infoln("SendTest Success!")
				}
				time.Sleep(time.Second / 2)
			}
		}
	},
}
