package exporter

import (
	"fmt"
	"io"
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

func (k *koinlyExporter) WriteRecord(writer io.Writer, record ExportRecord) {
	//Date,Sent Amount,Sent Currency,Received Amount,Received Currency,Fee Amount,Fee Currency,Net Worth Amount,Net Worth Currency,Label,Description,TxHash
	fmt.Fprintf(writer, "%s UTC,", record.blockTime.UTC().Format("2006-01-02 15:04:05"))
	if record.sentQty != 0 {
		fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.sentQty))
	} else {
		fmt.Fprintf(writer, ",,")
	}
	if record.recvQty != 0 {
		fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.recvQty))
	} else {
		fmt.Fprintf(writer, ",,")
	}
	fmt.Fprintf(writer, ",,")
	var label string
	if record.reward {
		label = "staking"
		record.txid = "reward-" + record.txid
	}
	fmt.Fprintf(writer, ",,")
	fmt.Fprintf(writer, "%s,,%s\n", label, record.txid)
}
