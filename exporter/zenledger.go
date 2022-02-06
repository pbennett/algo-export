package exporter

import (
	"fmt"
	"io"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

func init() {
	registerFormat("zenledger", NewzenledgerExporter)
}

type zenledgerExporter struct {
}

func NewzenledgerExporter() Interface {
	return &zenledgerExporter{}
}

func (k zenledgerExporter) Name() string {
	return "zenledger"
}

func (k *zenledgerExporter) WriteHeader(writer io.Writer) {
	// https://support.zenledger.io/en/articles/2615489-how-do-i-create-a-custom-csv
	// https://app.zenledger.io/static/assets/generalized-csv-sample.csv
	fmt.Fprintln(writer, "Timestamp,Type,IN Amount,IN Currency,Out Amount,Out Currency,Fee Amount,Fee Currency,Exchange(optional),US Based,Txid")
}

func (k *zenledgerExporter) WriteRecord(writer io.Writer, record ExportRecord, assetMap map[uint64]models.Asset) {
	//Timestamp,Type,IN Amount,IN Currency,Out Amount,Out Currency,Fee Amount,Fee Currency,Exchange(optional),US Based,Txid
	fmt.Fprint(writer, record.blockTime.UTC().Format("01/02/2006T15:04:05Z,"))
	
	// Types: 'buy', 'sell,' 'trade,' 'receive', 'send', 'Initial Coin Offering',
	//   'margin trade', 'staking', 'fork', 'airdrop', 'payment', 'mined',
	//   'gift sent', 'fee', 'staking reward', 'dividend received',
	//   'interest received', 'misc_reward', 'margin gain', 'margin loss',
	//   'lost', 'stolen,' 'nft_mint'
	switch {
	case record.reward:
		fmt.Fprintf(writer, "misc_reward,")
	case record.feeTx:
		fmt.Fprintf(writer, "fee,")
	case record.recvQty != 0 && record.sentQty != 0:
		fmt.Fprintf(writer, "trade,")
	case record.recvQty != 0:
		fmt.Fprintf(writer, "receive,")
	default:
		fmt.Fprintf(writer, "send,")
	}
	
	// IN Amount,IN Currency,
	switch {
	case record.recvQty != 0:
		fmt.Fprintf(writer, "%s,%s,", assetIDFmt(record.recvQty, record.assetID, assetMap), asaFmt(record.assetID, assetMap))
	default:
		fmt.Fprintf(writer, ",,")
	}
	
	// Out Amount,Out Currency,
	switch {
	case record.sentQty != 0:
		fmt.Fprintf(writer, "%s,%s,", assetIDFmt(record.sentQty, record.assetID, assetMap), asaFmt(record.assetID, assetMap))
	default:
		fmt.Fprintf(writer, ",,")
	}
	
	// Fee Amount,Fee Currency,
	switch {
	case record.fee != 0:
		fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.fee))
	default:
		fmt.Fprintf(writer, ",,")
	}
	
	// Exchange(optional),
	fmt.Fprintf(writer, "ALGO Wallet,")

	// US Based,
	fmt.Fprintf(writer, "Yes,")
	
	// Txid
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

	fmt.Fprintf(writer, "\n")
}
