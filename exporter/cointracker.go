package exporter

import (
	"fmt"
	"io"
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

func (k *cointrackerExporter) WriteRecord(writer io.Writer, record ExportRecord) {
	//Date,Received Quantity,Received Currency,Sent Quantity,Sent Currency,Fee Amount,Fee Currency,Tag
	fmt.Fprint(writer, record.blockTime.UTC().Format("01/02/2006 15:04:05,"))
	fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.recvQty))
	fmt.Fprintf(writer, "%s,ALGO,", algoFmt(record.sentQty))
	fmt.Fprintf(writer, ",,")
	// cointracker only supports tag field for specifying type - can't pass txid, descriptions, etc.
	var tag string
	if record.reward {
		tag = "staked"
	}
	fmt.Fprintln(writer, tag)

}
