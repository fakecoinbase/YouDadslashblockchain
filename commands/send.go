package commands

import (
	"github.com/YouDad/blockchain/api"
	"github.com/YouDad/blockchain/core"
	"github.com/YouDad/blockchain/log"
	"github.com/YouDad/blockchain/wallet"
	"github.com/spf13/cobra"
)

var (
	sendFrom   string
	sendTo     string
	sendAmount int
	sendMine   bool
)

func init() {
	SendCmd.Flags().StringVar(&sendFrom, "from", "", "Source wallet address")
	SendCmd.Flags().StringVar(&sendTo, "to", "", "Destination wallet address")
	SendCmd.Flags().IntVar(&sendAmount, "amount", -1, "Amount to send")
	SendCmd.Flags().BoolVar(&sendMine, "mine", false, "")
	SendCmd.MarkFlagRequired("from")
	SendCmd.MarkFlagRequired("to")
	SendCmd.MarkFlagRequired("amount")
}

var SendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send AMOUNT of coins from FROM address to TO",
	Run: func(cmd *cobra.Command, args []string) {
		if !wallet.ValidateAddress(sendFrom) {
			log.Errln("Sender address is not valid")
		}

		if !wallet.ValidateAddress(sendTo) {
			log.Errln("Recipient address is not valid")
		}

		bc := core.GetBlockchain()
		utxoSet := core.NewUTXOSet()

		tx := utxoSet.NewUTXOTransaction(sendFrom, sendTo, sendAmount)

		if sendMine {
			cbTx := core.NewCoinbaseTxn(sendFrom)
			txs := []*core.Transaction{cbTx, tx}

			newBlocks := bc.MineBlock(txs)
			utxoSet.Update(newBlocks)
		} else {
			api.SendTransaction(tx)
		}
		log.Infoln("Success!")
	},
}