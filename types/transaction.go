package types

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/YouDad/blockchain/log"
	"github.com/YouDad/blockchain/utils"
)

type Transaction struct {
	Vin  []TxnInput
	Vout []TxnOutput
}

func (txn Transaction) Hash() HashValue {
	return utils.SHA256(txn)
}

func (txn Transaction) String() string {
	lines := []string{}

	lines = append(lines, fmt.Sprintf("\n\t\tTxnHash %s:", txn.Hash()))

	for i, input := range txn.Vin {

		lines = append(lines, fmt.Sprintf("\t\t+ Input %d:", i))
		lines = append(lines, fmt.Sprintf("\t\t  - VoutHash:   %s", input.VoutHash))
		lines = append(lines, fmt.Sprintf("\t\t  - VoutIndex:  %d", input.VoutIndex))
		lines = append(lines, fmt.Sprintf("\t\t  - Signature:  %x", input.Signature))
		lines = append(lines, fmt.Sprintf("\t\t  - PubKeyHash: %x", input.PubKeyHash))
	}

	for i, output := range txn.Vout {
		lines = append(lines, fmt.Sprintf("\t\t+ Output %d:", i))
		lines = append(lines, fmt.Sprintf("\t\t  - Value:      %d", output.Value))
		lines = append(lines, fmt.Sprintf("\t\t  - PubKeyHash: %x", output.PubKeyHash))
	}

	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

func (txn Transaction) IsCoinbase() bool {
	return len(txn.Vin) == 1 && txn.Vin[0].VoutIndex == -1
}

func (txn Transaction) TrimmedCopy() Transaction {
	var inputs []TxnInput
	var outputs []TxnOutput

	for _, vin := range txn.Vin {
		inputs = append(inputs, TxnInput{
			VoutHash:   vin.VoutHash,
			VoutIndex:  vin.VoutIndex,
			VoutValue:  vin.VoutValue,
			Signature:  nil,
			PubKeyHash: nil,
		})
	}

	for _, vout := range txn.Vout {
		outputs = append(outputs, TxnOutput{
			Value:      vout.Value,
			PubKeyHash: vout.PubKeyHash,
		})
	}

	txCopy := Transaction{
		Vin:  inputs,
		Vout: outputs,
	}

	return txCopy
}

func (txn *Transaction) Sign(sk PrivateKey, hashedTxn map[string]Transaction) {
	if txn.IsCoinbase() {
		return
	}

	txnCopy := txn.TrimmedCopy()

	for inIndex, vin := range txnCopy.Vin {
		prevTxn := hashedTxn[vin.VoutHash.String()]
		txnCopy.Vin[inIndex].PubKeyHash = prevTxn.Vout[vin.VoutIndex].PubKeyHash
		dataToSign := []byte(fmt.Sprintf("%s\n", txnCopy))

		r, s, err := ecdsa.Sign(rand.Reader, &sk, dataToSign)
		log.Err(err)
		signature := append(r.Bytes(), s.Bytes()...)

		txn.Vin[inIndex].Signature = signature
		txnCopy.Vin[inIndex].PubKeyHash = nil
	}
}

func (txn Transaction) Verify(hashedTxn map[string]Transaction) bool {
	if txn.IsCoinbase() {
		return true
	}

	txnCopy := txn.TrimmedCopy()
	curve := elliptic.P256()

	for inIndex, vin := range txn.Vin {
		prevTxn := hashedTxn[vin.VoutHash.String()]
		txnCopy.Vin[inIndex].PubKeyHash = prevTxn.Vout[vin.VoutIndex].PubKeyHash
		dataToVerify := []byte(fmt.Sprintf("%s\n", txnCopy))

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKeyHash)
		x.SetBytes(vin.PubKeyHash[:(keyLen / 2)])
		y.SetBytes(vin.PubKeyHash[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, dataToVerify, &r, &s) {
			return false
		}
		txnCopy.Vin[inIndex].PubKeyHash = nil
	}

	return true
}
