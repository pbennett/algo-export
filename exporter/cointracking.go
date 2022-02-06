package exporter

import (
	"fmt"
	"io"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

func init() {
	registerFormat("cointracking", NewcointrackingExporter)
}

type cointrackingExporter struct {
}

func NewcointrackingExporter() Interface {
	return &cointrackingExporter{}
}

func (k cointrackingExporter) Name() string {
	return "cointracking"
}

func (k *cointrackingExporter) WriteHeader(writer io.Writer) {
	// https://cointracking.info/import/import_csv/
	// If you want to create your own CSV file, please ensure the format is like this:
	// "Type", "Buy Amount", "Buy Currency", "Sell Amount", "Sell Currency", "Fee", "Fee Currency", "Exchange", "Trade-Group", "Comment", "Date"
	// Optionally you can add those 3 columns at the end (after the "Date" column):
	// "Tx-ID", "Buy Value in your Account Currency", "Sell Value in your Account Currency"
	fmt.Fprintln(writer, "Type,Buy Amount,Buy Currency,Sell Amount,Sell Currency,Fee,Fee Currency,Exchange,Trade-Group,Comment,Date,Tx-ID")
}

func (k *cointrackingExporter) WriteRecord(writer io.Writer, record ExportRecord, assetMap map[uint64]models.Asset) {
	// Type,Buy Amount,Buy Currency,Sell Amount,Sell Currency,Fee,Fee Currency,Exchange,Trade-Group,Comment,Date,Tx-ID

	// Type,
	// https://cointracking.freshdesk.com/en/support/solutions/articles/29000034379-expanded-transaction-types-may-2020-
	switch {
	case record.reward:
		fmt.Fprintf(writer, "Reward / Bonus,")
	case record.feeTx:
		fmt.Fprintf(writer, "Withdrawal,")
	case record.recvQty != 0 && record.sentQty != 0:
		fmt.Fprintf(writer, "Trade,")
	case record.recvQty != 0:
		fmt.Fprintf(writer, "Deposit,")
	default:
		fmt.Fprintf(writer, "Withdrawal,")
	}
	
	// Buy Amount,Buy Currency,
	switch {
	case record.recvQty != 0:
		fmt.Fprintf(writer, "%s,%s,", assetIDFmt(record.recvQty, record.assetID, assetMap), asaFmt(record.assetID, assetMap))
	default:
		fmt.Fprintf(writer, ",,")
	}
	
	// Sell Amount,Sell Currency,
	switch {
	case record.sentQty != 0:
		fmt.Fprintf(writer, "%s,%s,", assetIDFmt(record.sentQty, record.assetID, assetMap), asaFmt(record.assetID, assetMap))
	default:
		fmt.Fprintf(writer, ",,")
	}
	
	// Fee,Fee Currency,
	switch {
	case record.fee != 0:
		fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.fee))
	default:
		fmt.Fprintf(writer, ",,")
	}

	// Exchange,
	fmt.Fprintf(writer, "ALGO Wallet,")

	// Trade-Group,
	fmt.Fprintf(writer, "%s,", record.account)

	// Comment,
	fmt.Fprintf(writer, "%q,", asaComment(record.assetID, assetMap))

	// Date,
	fmt.Fprint(writer, record.blockTime.UTC().Format("2006-01-02T15:04:05Z,"))

	// Tx-ID,
	switch {
	case record.topTxID != "":
		fmt.Fprintf(writer, "%s_%s", record.topTxID, record.account[:10])
	default:
		fmt.Fprintf(writer, "%s_%s", record.txid, record.account[:10])
	}
	switch {
	case record.reward:
		fmt.Fprintf(writer, "_reward")
	case record.feeTx:
		fmt.Fprintf(writer, "_fee")
	}
	
	fmt.Fprint(writer, "\n")
}
