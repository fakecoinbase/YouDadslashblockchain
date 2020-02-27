package core

import (
	"encoding/hex"

	"github.com/YouDad/blockchain/conf"
	"github.com/YouDad/blockchain/log"
	"github.com/YouDad/blockchain/utils"
)

type UTXOSet struct {
	*Blockchain
}

func GetUTXOSet() *UTXOSet {
	return &UTXOSet{GetBlockchain()}
}

func (set *UTXOSet) Update(b *Block) {
	log.NotImplement()

}

func (set *UTXOSet) Reindex() {
	hashedUtxos := set.FindUTXO()
	set.SetTable(conf.UTXOSET).Clear()

	for txnHash, utxos := range hashedUtxos {
		hash, err := hex.DecodeString(txnHash)
		log.Err(err)
		set.Set(hash, utils.Encode(utxos))
	}
}

func (set *UTXOSet) NewUTXOTransaction(from, to string, amount int) *Transaction {
	log.NotImplement()
	return nil
}

func (set *UTXOSet) FindUTXOByHash(pubKeyHash []byte) []TxnOutput {
	utxos := []TxnOutput{}

	set.SetTable(conf.UTXOSET).Foreach(func(k, v []byte) bool {
		outs := BytesToTxnOutputs(v)

		for _, out := range outs {
			if out.IsLockedWithKey(pubKeyHash) {
				utxos = append(utxos, out)
			}
		}
		return true
	})

	return utxos
}
