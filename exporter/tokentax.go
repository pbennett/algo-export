package exporter

import (
	"fmt"
	"io"
	"os"
)

/*
	NOTHING HERE IS TAX ADVICE NOR SHOULD BE INTERPRETED AS SUCH.

	This TokenTax exporter makes a key classification / assumption:

	All transactions are either a Withdrawal, Deposit or Staking reward.
	Since we can only see Algo-in / Algo-out, we classify "Algo-in" as a deposit, "Algo-out"
	as a withdrawal. Staking rewards are "Staking". If an "Algo-out" transaction is to
	a wallet that isn't yours, it'll likely show up in your Tax reporting software
	and will likely need to be reclassified from "Withdrawal" -> "Spend". (AGAIN THIS IS NOT
	TAX ADVICE NOT SHOULD BE INTERPRETED AS SUCH.)
*/

type TokenTaxExporter struct {
}

func (t TokenTaxExporter) Name() string {
	return "tokentax"
}

func (t *TokenTaxExporter) WriteHeader(writer io.Writer) {
	header := "Type,BuyAmount,BuyCurrency,SellAmount,SellCurrency,FeeAmount,FeeCurrency,Exchange,Group,Comment,Date"
	fmt.Fprintln(writer, header)
}

func (t *TokenTaxExporter) WriteRecord(writer io.Writer, record ExportRecord) {

	var transactionType, buyAmt, sellAmt string
	if record.reward {
		transactionType = "Staking"
		buyAmt = algoFmt(record.recvQty)
	} else {
		// If we have a Buy and Sell Amt, this is a "Trade".
		// This program isn't meant to track trades. Since we're just
		// tracking down transactions between wallets for tax purposes.
		// If you're buying and selling, you're likely using an exchange,
		// these exchanges can wire into TokenTax (in this case).
		if record.recvQty != 0 && record.sentQty != 0 {
			msg := "Detected Buy and Sell qtys, this is likely a trade. This program doesn't support trades right now"
			fmt.Println(msg)
			os.Exit(1)
		} else if record.recvQty != 0 {
			transactionType = "Deposit"
			buyAmt = algoFmt(record.recvQty)
		} else if record.sentQty != 0 {
			transactionType = "Withdrawal"
			sellAmt = algoFmt(record.sentQty)
		}
	}
	fmt.Fprintf(writer, "%s,", transactionType)
	if transactionType == "Deposit" || transactionType == "Staking" {
		fmt.Fprintf(writer, "%s,ALGO,,,,,,,,", buyAmt)
	} else if transactionType == "Withdrawal" {
		fmt.Fprintf(writer, ",,%s,ALGO,,,,,,", sellAmt)
	}
	// Finally, the date
	fmt.Fprintf(writer, "%s UTC\n", record.blockTime.UTC().Format("01/02/2006 15:04"))
}

func NewTokenTaxExporter() Interface {
	return &TokenTaxExporter{}
}

func init() {
	registerFormat("tokentax", NewTokenTaxExporter)
}
