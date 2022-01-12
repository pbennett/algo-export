package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
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

func exportTransactions(client *indexer.Client, export exporter.Interface, account string, outCsv io.Writer, assetMap map[uint64]models.Asset, txns []models.Transaction) error {
	for _, tx := range txns {
		// Recursive export of inner transactions.
		if len(tx.InnerTxns) > 0 {
			fmt.Printf("    processing %d inner transaction(s) for transaction id: %s\n", len(tx.InnerTxns), tx.Id)
			if err := exportTransactions(client, export, account, outCsv, assetMap, tx.InnerTxns); err != nil {
				return err
			}
		}

		// Populate assetMap if entry does not exist.
		if tx.AssetTransferTransaction.AssetId != 0 {
			if _, ok := assetMap[tx.AssetTransferTransaction.AssetId]; !ok {
				// Rate limited to <1 request per second.
				time.Sleep(2 * time.Second)

				fmt.Println("    looking up Asset ID:", tx.AssetTransferTransaction.AssetId)
				lookupASA := client.LookupAssetByID(tx.AssetTransferTransaction.AssetId)
				_, asset, err := lookupASA.Do(context.TODO())
				if err != nil {
					return fmt.Errorf("error looking up asset id: %w", err)
				}
				assetMap[tx.AssetTransferTransaction.AssetId] = asset
			}
		}
		for _, record := range exporter.FilterTransaction(tx, account, assetMap) {
			export.WriteRecord(outCsv, record, assetMap)
		}
	}
	return nil
}


func exportAccounts(client *indexer.Client, export exporter.Interface, accounts accountList, outDir string) error {
	state := LoadConfig()
	assetMap := make(map[uint64]models.Asset)

	fmt.Println("Exporting accounts:")
	for _, accountAddress := range accounts {
		// accountAddress contains the non-checksummed internal version - String() provides the
		// version users know - the base32 pubkey w/ checksum
		account := accountAddress.String()

		startRound := state.ForAccount(export.Name(), account).LastRound + 1
		fmt.Println(account, "starting at:", startRound)

		nextToken := ""
		numPages := 1
		for {
			lookupTx := client.LookupAccountTransactions(account)
			lookupTx.MinRound(startRound)
			lookupTx.NextToken(nextToken)
			transactions, err := lookupTx.Do(context.TODO())
			if err != nil {
				return fmt.Errorf("error looking up transactions: %w", err)
			}
			endRound := transactions.CurrentRound
			state.ForAccount(export.Name(), account).LastRound = endRound

			numTx := len(transactions.Transactions)
			fmt.Printf("  %v transactions\n", numTx)
			if numTx == 0 {
				break
			}
			outCsv, err := os.Create(filepath.Join(outDir, fmt.Sprintf("%s-%s-%d-%d-%d.csv", export.Name(), account, startRound, endRound, numPages)))
			export.WriteHeader(outCsv)
			if err := exportTransactions(client, export, account, outCsv, assetMap, transactions.Transactions); err != nil {
				return err
			}
			fmt.Printf("  %v NextToken at Page %d\n", transactions.NextToken, numPages)
			nextToken = transactions.NextToken
			numPages++

			// Rate limited to <1 request per second.
			time.Sleep(2 * time.Second)
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
