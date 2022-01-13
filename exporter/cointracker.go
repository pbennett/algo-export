package exporter

import (
	"fmt"
	"io"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

func init() {
	registerFormat("cointracker", NewCointrackerExporter)
}

type cointrackerExporter struct {
}

func NewCointrackerExporter() Interface {
	return &cointrackerExporter{}
}

func (k cointrackerExporter) Name() string {
	return "cointracker"
}

func (k *cointrackerExporter) WriteHeader(writer io.Writer) {
	fmt.Fprintln(writer, "Date,Received Quantity,Received Currency,Sent Quantity,Sent Currency,Fee Amount,Fee Currency,Tag")
}

func (k *cointrackerExporter) WriteRecord(writer io.Writer, record ExportRecord, assetMap map[uint64]models.Asset) {
	//Date,Received Quantity,Received Currency,Sent Quantity,Sent Currency,Fee Amount,Fee Currency,Tag
	fmt.Fprint(writer, record.blockTime.UTC().Format("01/02/2006 15:04:05,"))
	
	if record.assetID != 0 {
		// https://www.cointracker.io/currencies/custom
		fmt.Fprintf(writer, "%s,%s,", assetIDFmt(record.recvQty, record.assetID, assetMap), asaFmt(record.assetID, assetMap))
		fmt.Fprintf(writer, "%s,%s,", assetIDFmt(record.sentQty, record.assetID, assetMap), asaFmt(record.assetID, assetMap))
	} else {
		switch {
		case record.recvQty == 0:
			fmt.Fprintf(writer, ",,")
		default:
			fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.recvQty))
		}
		switch {
		case record.sentQty == 0:
			fmt.Fprintf(writer, ",,")
		default:
			fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.sentQty))
		}
	}

	if record.fee != 0 {
		fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.fee))
	} else {
		fmt.Fprintf(writer, ",,")
	}

	// cointracker only supports tag field for specifying type - can't pass txid, descriptions, etc.
	var tag string
	if record.reward {
		tag = "staked"
	}
	fmt.Fprintln(writer, tag)

}
