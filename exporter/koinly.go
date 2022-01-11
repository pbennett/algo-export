package exporter

import (
	"fmt"
	"io"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

func init() {
	registerFormat("koinly", NewKoinlyExporter)
}

type koinlyExporter struct {
}

func NewKoinlyExporter() Interface {
	return &koinlyExporter{}
}

func (k koinlyExporter) Name() string {
	return "koinly"
}

func (k *koinlyExporter) WriteHeader(writer io.Writer) {
	fmt.Fprintln(writer, "Date,Sent Amount,Sent Currency,Received Amount,Received Currency,Fee Amount,Fee Currency,Net Worth Amount,Net Worth Currency,Label,Description,TxHash")
}

func (k *koinlyExporter) WriteRecord(writer io.Writer, record ExportRecord, assetMap map[uint64]models.Asset) {
	//Date,Sent Amount,Sent Currency,Received Amount,Received Currency,Fee Amount,Fee Currency,Net Worth Amount,Net Worth Currency,Label,Description,TxHash
	fmt.Fprintf(writer, "%s UTC,", record.blockTime.UTC().Format("2006-01-02 15:04:05"))
	switch {
	case record.sentQty != 0 && record.assetID != 0:
		fmt.Fprintf(writer, "%s,ASA-%d,", assetIDFmt(record.sentQty, record.assetID, assetMap),record.assetID)
	case record.sentQty != 0:
		fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.sentQty))
	default:
		fmt.Fprintf(writer, ",,")
	}

	switch {
	case record.recvQty != 0 && record.assetID != 0:
		fmt.Fprintf(writer, "%s,ASA-%d,", assetIDFmt(record.recvQty, record.assetID, assetMap),record.assetID)
	case record.recvQty != 0:
		fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.recvQty))
	default:
		fmt.Fprintf(writer, ",,")
	}

	if record.fee != 0 {
		fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.fee))
	} else {
		fmt.Fprintf(writer, ",,")
	}
	var label string
	if record.reward {
		label = "staking"
		record.txid = "reward-" + record.txid
	}
	fmt.Fprintf(writer, ",,")
	fmt.Fprintf(writer, "%s,,%s\n", label, record.txid)
}
