package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/algorand/go-algorand-sdk/types"
	"github.com/pbennett/algo-export/exporter"
)

type accountList []types.Address

func (al *accountList) String() string {
	return fmt.Sprint(*al)
}

func (al *accountList) Set(value string) error {
	*al = accountList{}
	for _, val := range strings.Split(value, ",") {
		address, err := types.DecodeAddress(val)
		if err != nil {
			return fmt.Errorf("address:%v not valid: %w", address, err)
		}
		*al = append(*al, address)
	}
	return nil
}

func main() {
	var (
		accounts         accountList
		formatFlag       = flag.String("f", exporter.Formats()[0], fmt.Sprintf("Format to export: [%s]", strings.Join(exporter.Formats(), ", ")))
		hostAddrFlag     = flag.String("s", "localhost:8980", "Index server to connect to")
		apiKey           = flag.String("api", "", "Optional API Key for local indexer, or for PureStake")
		pureStakeApiFlag = flag.Bool("p", false, "Use PureStake API - ignoring -s argument")
		outDirFlag       = flag.String("o", "", "output directory path for exported files")
	)
	flag.Var(&accounts, "a", "Account or list of comma delimited accounts to export")
	flag.Parse()

	if len(accounts) == 0 {
		fmt.Println("One or more account addresses to export must be specified.")
		flag.Usage()
		os.Exit(1)
	}
	var export = exporter.GetFormatter(*formatFlag)
	if export == nil {
		fmt.Println("Unable to find formatter for:", *formatFlag)
		fmt.Println("Valid formats are:\n", strings.Join(exporter.Formats(), "\n "))
		os.Exit(1)
	}

	client, err := getClient(*hostAddrFlag, *apiKey, *pureStakeApiFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if !fileExist(*outDirFlag) {
		if err = os.MkdirAll(*outDirFlag, 0755); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err := exportAccounts(client, export, accounts, *outDirFlag); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getClient(serverFlag string, apiKey string, usePureStake bool) (*indexer.Client, error) {
	var (
		client     *indexer.Client
		serverAddr *url.URL
		err        error
	)
	if !usePureStake {
		serverAddr, err = url.Parse(fmt.Sprintf("http://%s", serverFlag))
		if err != nil {
			return nil, fmt.Errorf("error in server address: %w", err)
		}
		client, err = indexer.MakeClient(serverAddr.String(), apiKey)
		if err != nil {
			return nil, fmt.Errorf("error creating indexer client: %w", err)
		}
	} else {
		commonClient, err := common.MakeClientWithHeaders("https://mainnet-algorand.api.purestake.io/idx2", "X-API-Key", apiKey, []*common.Header{})
		if err != nil {
			return nil, fmt.Errorf("error creating indexer client to purestake: %w", err)
		}
		client = (*indexer.Client)(commonClient)
	}
	return client, err
}

func exportAccounts(client *indexer.Client, export exporter.Interface, accounts accountList, outDir string) error {
	state := LoadConfig()
	fmt.Println("Exporting accounts:")
	for _, accountAddress := range accounts {
		// accountAddress contains the non-checksummed internal version - String() provides the
		// version users know - the base32 pubkey w/ checksum
		account := accountAddress.String()

		startRound := state.ForAccount(export.Name(), account).LastRound + 1
		fmt.Println(account, "starting at:", startRound)

		lookupTx := client.LookupAccountTransactions(account)
		lookupTx.MinRound(startRound)
		transactions, err := lookupTx.Do(context.TODO())
		if err != nil {
			return fmt.Errorf("error looking up transactions: %w", err)
		}
		endRound := transactions.CurrentRound
		state.ForAccount(export.Name(), account).LastRound = endRound

		fmt.Printf("  %v transactions\n", len(transactions.Transactions))
		if len(transactions.Transactions) == 0 {
			continue
		}
		outCsv, err := os.Create(filepath.Join(outDir, fmt.Sprintf("%s-%s-%d-%d.csv", export.Name(), account, startRound, endRound)))
		export.WriteHeader(outCsv)
		for _, tx := range transactions.Transactions {
			for _, record := range exporter.FilterTransaction(tx, account) {
				export.WriteRecord(outCsv, record)
			}
		}
	}
	state.SaveConfig()
	return nil
}

func fileExist(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Fatalln(err)
	}
	return true
}
