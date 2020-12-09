>"...in this world nothing can be said to be certain, except death and taxes." 
>Benjamin Franklin

Cryptocurrency taxation is certainly controversial, but for citizens in most countries, it's an unavoidable part of citizenship. For those in a region which treats crypto as a taxable asset, it's not something to ignore. Because most chains are by design, completely public, there's no hiding. Privacy chains or anonymization services can even trigger suspicion - just in their use.

Unfortunately, it can be difficult to track your crypto transactions. At least in the United States, *every single transaction* is effectively a taxable event and must be tracked appropriately. This can get cumbersome quickly. 

This is why there are so many sites which will track your crypto activity for you and provide tax reports, tracking the proper cost-basis across potentially hundreds or thousands of transactions. They can automatically sync with exchanges as well as on-chain activity from wallet addresses you provide. Most of these sites support wallet scanning directly, but Algorand is still young, so direct on-chain scanning isn't available yet.


This solution provides a simple starting point for a tool to export on-chain Algorand transactions to CSV files compatible with two of these crypto tax sites:

* [CoinTracker](https://www.cointracker.io/)
* [Koinly](https://koinly.io/)
* [TokenTax](https://tokentax.co/)

CoinTracker has an excellent tax guide if you'd lke more details on the subject: https://www.cointracker.io/blog/crypto-tax-guide

Koinly is an excellent choice as well. There are pros/cons to all of these sites and with varying fees & features. They're worth a look.

TokenTax is another simple and easy to use option, but comes at a cost. The report generated from this tool tracks "Deposits", "Withdrawals" and "Staking" income. Combined with
data one's other wallets/exchanges, TokenTax can show captial gains as well as income received from Algorand staking.

[TOC]

# Requirements 

- A basic working knowledge of Go is recommended
- A working [Go](https://golang.org/) installation
- The code for this program is all in Go, and uses the new Algorand-SDK V2 client [SDK](https://github.com/algorand/go-algorand-sdk) 
- This is used to retrieve all transactions for each account.
- A local indexer to connect to.
- See the [Indexer](https://developer.algorand.org/docs/run-a-node/setup/indexer/) page
- **or** a PureStake API key
- For low-volume use, you can use the PureStake API service for free. 
- You can sign up for a [free account](https://www.purestake.com/technology/algorand-api/).
- Using a public API service like https://algoexplorer.io is an option, but it doesn't support the V2 indexer API yet.

# Overview

This solution will walk through a simple command-line program for exporting one or more Algorand accounts and their transaction histories to CSV files for consumption by tax reporting sites.

The program is simple but parsing out the transaction details aren't particularly obvious.

!!! note
To keep this solution simple, Algorand Standard Assets will be left out for a later exercise.

# Example use

Before walking through the code, a quick example of the program's use is in order.

I picked a random account on MainNet as well as one of the accounts it sent to, and exported both using the following command:

```
algo-export -o test -f koinly -a HFTA36U4OCTSMXRUH4ZX3OACJBTJCR56AIH3G345TRPUQJHJBEXKLMMO4E,AV5EPTMH2RZJ2V72PR2WC63EMAMQOPKI2EDN4TU2XFA2WTAJN4VKKLODVI
Exporting accounts:
HFTA36U4OCTSMXRUH4ZX3OACJBTJCR56AIH3G345TRPUQJHJBEXKLMMO4E starting at: 1
62 transactions
AV5EPTMH2RZJ2V72PR2WC63EMAMQOPKI2EDN4TU2XFA2WTAJN4VKKLODVI starting at: 1
60 transactions
``` 
Looking in the local `test/` directory it created, we see:

```
>ls -l test
total 40
-rw-r--r--+ 1 patrickb staff 6478 Jul 14 20:13 koinly-AV5EPTMH2RZJ2V72PR2WC63EMAMQOPKI2EDN4TU2XFA2WTAJN4VKKLODVI-1-7858071.csv
-rw-r--r--+ 1 patrickb staff 10636 Jul 14 20:13 koinly-HFTA36U4OCTSMXRUH4ZX3OACJBTJCR56AIH3G345TRPUQJHJBEXKLMMO4E-1-7858070.csv
```

If we examine the first few lines of the second file we see the csv:

```csv
Date,Sent Amount,Sent Currency,Received Amount,Received Currency,Fee Amount,Fee Currency,Net Worth Amount,Net Worth Currency,Label,Description,TxHash
2020-07-14 22:05:44 UTC,4957.108696,ALGO,,,,,,,,,5GDWCVNIDHIAWGMI323DAZ2HSWB7NK6UQRXRXPW6NSH5EEZNRWQA
2020-07-14 22:05:43 UTC,,,0.371700,ALGO,,,,,staking,,reward-5GDWCVNIDHIAWGMI323DAZ2HSWB7NK6UQRXRXPW6NSH5EEZNRWQA
2020-07-14 10:51:45 UTC,,,4956.733300,ALGO,,,,,,,U5JN2L65WXVAGXHWDHAZQPC4GDG2AGIB4HR5MAKBLAAIY52B55YQ
2020-07-04 06:13:17 UTC,8652.734300,ALGO,,,,,,,,,4CJQ6AXIOLWLD2J5BQS6Z7QHUX2KO5E7F45OKJ3SBYG6CWESS7NQ
2020-07-04 06:13:16 UTC,,,0.207648,ALGO,,,,,staking,,reward-4CJQ6AXIOLWLD2J5BQS6Z7QHUX2KO5E7F45OKJ3SBYG6CWESS7NQ
```

...and in table form: 

|Date|Sent Amount|Sent Currency|Received Amount|Received Currency|Fee Amount|Fee Currency|Net Worth Amount|Net Worth Currency|Label|Description|TxHash|
|---|---|---|---|---|---|---|---|---|---|---|---| 
|2020-07-14 22:05:44 UTC|4957.108696|ALGO| | | | | | | | |5GDWCVNIDHIAWGMI323DAZ2HSWB7NK6UQRXRXPW6NSH5EEZNRWQA|
|2020-07-14 22:05:43 UTC| | |0.371700|ALGO| | | | |staking| |reward-5GDWCVNIDHIAWGMI323DAZ2HSWB7NK6UQRXRXPW6NSH5EEZNRWQA|
|2020-07-14 10:51:45 UTC| | |4956.733300|ALGO| | | | | | |U5JN2L65WXVAGXHWDHAZQPC4GDG2AGIB4HR5MAKBLAAIY52B55YQ|
|2020-07-04 06:13:17 UTC|8652.734300|ALGO| | | | | | | | |4CJQ6AXIOLWLD2J5BQS6Z7QHUX2KO5E7F45OKJ3SBYG6CWESS7NQ|
|2020-07-04 06:13:16 UTC| | |0.207648|ALGO| | | | |staking| |reward-4CJQ6AXIOLWLD2J5BQS6Z7QHUX2KO5E7F45OKJ3SBYG6CWESS7NQ|

Notice we see sends, receives, and synthesized 'staking' *reward* transactions. 

As a quick example, for Koinly (since that was what was chosen to export above), I created an account, clicked "Wallets", then "Add Wallet / Exchange".

The screen will look like:
![Add Wallet](https://algorand-devloper-portal-app.s3.amazonaws.com/static/EditorImages/2020/07/15%2001%3A07/Screen_Shot_2020-07-14_at_8.27.43_PM.png)

Type something like "csv" in the search box and hit enter. If it's not a known exchange, wallet, it'll assume it's a custom import. 

Click the "Create custom wallet with name 'csv' link." 

On the next screen, change the Wallet name to be something like Algo-XXXX where it's the first 4 or so characters of your account address. 

Click "Upload csv file"s, and then drag and drop the proper csv file or click Browse and load the file.

Repeat the procedure for each of your accounts (when importing your own).

For this example, I added the two wallets:

![Wallets](https://algorand-devloper-portal-app.s3.amazonaws.com/static/EditorImages/2020/07/15%2001%3A08/Screen_Shot_2020-07-14_at_9.06.19_PM.png)

Each time you re-run the algo-export program it will continue where it left off, exporting any new transactions that have occurred since it last ran. If there were no new transactons it won't create a file. So whenever you have new transactions, re-run the program with the same arguments, and import the new files it created.

These particular accounts sent to accounts which weren't imported so Koinly will assume they were sends, not transfers since it doesn't have a matching receive.

A brief view of the transactions as Koinly shows them is seen below: 

!!! note
Notice the staking rewards that were tagged. Koinly shows them as a Staking reward.

![Koinly Example](https://algorand-devloper-portal-app.s3.amazonaws.com/static/EditorImages/2020/07/15%2000%3A56/Screen_Shot_2020-07-14_at_8.32.33_PM.png) 

Both Koinly and CoinTracker have a ton of features and I recommend you evaluate on your own.

Now that we've shown the basics, let's get to the code. 

# Code walkthrough

## Account list definition

If you look at the main.go code, you'll notice the definition of a simple type called `accountList`. This is merely to wrap a new type that will be set by a command-line flag using the built-in flag package in Go. We just want to allow the user to specify one or more accounts (comma delimited), parse them for correctness, and add them to a slice. Notice that the addresses are converted into `types.Address` values by `types.DecodeAddress`. This function will return an error if the passed in account isn't a (possibly) valid Algorand account address. 

This type will be used later to define the type expected for the `-a` account flag.

``` go
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
```

## Flag initialization

Now we define the flags we want to accept. 
They are:

* `-f` - the 'format' our files should be in when created.
* `-s` - the index server to connect to (defaults to local indexer).
* `-a` - one or more accounts to export (comma delimited if more than one)
* `-api` - an API key for local indexer, or for PureStake
* `-p` - if using the PureStake API to access an indexer instead of a local instance
* `-o` - output directory to write .csv files (defaults to current directory)

The `flag.String` calls should be clear. `flag.Var` is where we specify that the accounts variable of type `accountList` should be used instead. `flag.Var` expects its passed type to conform to the `flag.Value` interface. It needs to implement `String()` and `Set(string)` error which our already defined `accountList` type does.

The `exporter.Formats()` call is to a simple wrapper package we will explore later. Since there are multiple formats supported, a simple way of adding support for multiple formats was needed. 

``` go
func main() {
var (
accounts accountList
formatFlag = flag.String("f", exporter.Formats()[0], fmt.Sprintf("Format to export: [%s]", strings.Join(exporter.Formats(), ", ")))
hostAddrFlag = flag.String("s", "localhost:8980", "Index server to connect to")
apiKey = flag.String("api", "", "Optional API Key for local indexer, or for PureStake")
pureStakeApiFlag = flag.Bool("p", false, "Use PureStake API - ignoring -s argument")
outDirFlag = flag.String("o", "", "output directory path for exported files")
)
flag.Var(&accounts, "a", "Account or list of comma delimited accounts to export")
flag.Parse()

if len(accounts) == 0 {
fmt.Println("One or more account addresses to export must be specified.")
flag.Usage()
os.Exit(1)
}
```

## Getting formatter implementation

The formatter name is used to get an instance of an exporter for the specified format. These exporters are defined in an exporter/ sub-package and are set up to 'register' themselves. `GetFormatter` will return: 

``` go
var export = exporter.GetFormatter(*formatFlag)
if export == nil {
fmt.Println("Unable to find formatter for:", *formatFlag)
fmt.Println("Valid formats are:\n", strings.Join(exporter.Formats(), "\n "))
os.Exit(1)
}
```

## Connecting to an indexer node

To keep things a little cleaner and because we're supporting two different indexer API connections, I've moved getting a connection to an indexer instance into a separate helper function. 

``` go
client, err := getClient(*hostAddrFlag, *apiKey, *pureStakeApiFlag)
if err != nil {
fmt.Println(err)
os.Exit(1)
}
```

The `getClient` helper function is passed the `-s [server address]` flag (as `hostAddrFlag`), the api key if specified, and whether or not to use the PureStake API.

If not using the PureStake API, the `hostAddrFlag` passed in is assumed to be something that can be appended to http:// and parsed by the built-in go `url.Parse` function. This is then passed to the built-in Go v2 `indexer.MakeClient` API including the api key (if needed).

If using the PureStake API, then we have to use the algorand-sdk `common.MakeClientWithHeaders` function to construct our client. We pass in a hardcoded URL, and passed in API Key (as `apiKey` variable) setting that into PureStake's required `X-API-Key` header field. The returned `*common.Client` is converted to an `*indexer.Client` and returned. 

``` go
func getClient(serverFlag string, apiKey string, usePureStake bool) (*indexer.Client, error) {
var (
client *indexer.Client
serverAddr *url.URL
err error
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
```

## Fetching Account transactions

After the retrieval of our client, we call another function, `exportAccounts` to export the list of accounts. We pass the client we just retrieved, the 'export' instance (which implements our exporter interface), and the list of accounts to export. Any errors are returned as-is.

``` go
os.MkdirAll(*outDirFlag, 0666)
if err := exportAccounts(client, export, accounts, *outDirFlag); err != nil {
fmt.Println(err)
os.Exit(1)
}
```

The export function loads the saved configuration from the last time the program is run...

``` go
func exportAccounts(client *indexer.Client, export exporter.Interface, accounts accountList, outDir string) error {
state := LoadConfig()
fmt.Println("Exporting accounts:")
```

then iterates over the provided accounts...

``` go
for _, accountAddress := range accounts {
// accountAddress contains the non-checksummed internal version - String() provides the
// version users know - the base32 pubkey w/ checksum
account := accountAddress.String()

startRound := state.ForAccount(export.Name(), account).LastRound + 1
fmt.Println(account, "starting at:", startRound)
```

The indexer API provides a method, `LookupAccountTransactions`, which fetches transactions for the specified account with optional parameters to filter the results. The only filtering we request is that we only want transactions occurring after the last block the program processed the last time it ran for this given 'format' and 'account.'

``` go
lookupTx := client.LookupAccountTransactions(account)
lookupTx.MinRound(startRound)
transactions, err := lookupTx.Do(context.TODO())
if err != nil {
return fmt.Errorf("error looking up transactions: %w", err)
}
endRound := transactions.CurrentRound
state.ForAccount(export.Name(), account).LastRound = endRound

fmt.Printf(" %v transactions\n", len(transactions.Transactions))
if len(transactions.Transactions) == 0 {
continue
}
```

If there are transactions to process, we create a csv file in the output directory named after the format, the account, and the start and end block number. This way, we only create files when there are transactions, and always create files containing only the new transactions since there were transactions to export. For platforms like Cointrack that don't properly handle importing duplicate records, this is required.

Once the file is created, we use the export implementation to write out its format-specific CSV header.We then iterate through all the transaction records, pass it through a generic `FilterTransaction` function, and then pass that through to the export implementation's `WriteRecord` method to write out the appropriate CSV for that record type.
``` go
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
```

## Filtering Transactions

Skipping some of the setup code, let's walk through the `FilterTransaction` function in the `exporter/` sub-package.

The function accepts a single `Transaction` instance (returned from the indexer) and the full string version of the account being exported. It returns a slice of `ExportRecord` structs, which are then passed to the exporter implementation for writing out.

Because Algorand accounts accrue 'pending' rewards automatically, but those rewards are only *applied* to the account when a transaction occurs, it's possible that a single transaction might need to be exported as multiple transactions. A receive to an account, or send from an account might have pending rewards. Those pending rewards should be added as a new 'staking reward' transaction immediately proceeding the transaction itself. We'll discuss this again when we get to that code.

``` go
func FilterTransaction(tx models.Transaction, account string) []ExportRecord {
var (
blockTime = time.Unix(int64(tx.RoundTime), 0).UTC()
recvAmount uint64
sendAmount uint64
rewards uint64
records []ExportRecord
)
```

The function next handles the various transaction types currently expressed in Algorand. The primary transaction is a 'pay' transaction. Sending from one account to another.

The first block of code checks to see if the account being exported is the account 'receiving' ALGO. Accounts can receive ALGO either via a simple Sender->Receiver transaction or via Sender->Receiver AND Sender->CloseRemainderTo. CloseRemainderTo is used to 'close' the sending account and ensures that any 'remaining' balance in the sender account is sent to the specified close-to account. The Receiver and Close To are often the same account, but can be different!

The 'rewards' variable is used for tracking rewards to apply later on for the account being exported. The comments are hopefully self-explanatory.

``` go
switch tx.Type {
case "pay":
if tx.PaymentTransaction.Receiver == account || tx.PaymentTransaction.CloseRemainderTo == account {
// We could potentially be receiver, AND close-to account so check independently
// We could be sender as well - so handle appropriately.
if tx.PaymentTransaction.Receiver == account {
recvAmount += tx.PaymentTransaction.Amount
rewards += tx.ReceiverRewards
}
if tx.PaymentTransaction.CloseRemainderTo == account {
recvAmount += tx.PaymentTransaction.CloseAmount + tx.ClosingAmount
rewards += tx.CloseRewards
}
// ...we could've sent to ourselves!
if tx.Sender == account {
sendAmount = tx.PaymentTransaction.Amount + tx.Fee
rewards += tx.SenderRewards
}
```

The `appendPostFilter` function acts like go's append function in that it returns a new slice, appending what's passed in. The key is that the `ExportRecord` we construct and pass in might get slightly modified by the `postFilter`, and possibly ignored. This simplifies some of the logic. See the full code for the details.

The result of this call is a new record which we'll return containing data about the amount received.

``` go
records = appendPostFilter(records, ExportRecord{
blockTime: blockTime,
txid: tx.Id,
recvQty: recvAmount,
receiver: account,
sentQty: sendAmount,
sender: tx.Sender,
})
```

This `else` block will handle pay transactions where the account is the sender. The sending case is a bit more involved as this is where we want to account for transaction fees. We also have to handle the case where we:

* Send to a single receiver
* Send to a receiver & close-to the same recipient
* Send to a receiver & close-to a different recipient

The recipient can also be the sender itself! This particular case is handled by the `appendPostFilter` function. Because these send operations may effectively involve sends to different accounts, we account for those as independent export records. This way, a tracking application will be able to match the send amounts with matching receive amounts if the receive account is an account also exported into that tracking system. This way it will be correctly tracked as an inter-wallet transfer.

``` go
} else {
// only choice at this point are sending transactions
rewards = tx.SenderRewards

// handle case where we close-to an account and it's not same as receiver so treat as if two sends for export purposes
// so receives can be matched in different accounts if user has both
if tx.PaymentTransaction.CloseRemainderTo != "" && tx.PaymentTransaction.Receiver != tx.PaymentTransaction.CloseRemainderTo {
// Frist, add transaction for close-to... (without fee)
records = appendPostFilter(records, ExportRecord{
blockTime: blockTime,
txid: tx.Id,
receiver: tx.PaymentTransaction.CloseRemainderTo,
sentQty: tx.PaymentTransaction.CloseAmount + tx.ClosingAmount,
sender: account,
})
// then add an extra transaction 1-sec later to base receiver (with fee)
records = appendPostFilter(records, ExportRecord{
blockTime: blockTime.Add(1 * time.Second),
txid: tx.Id,
receiver: tx.PaymentTransaction.Receiver,
sentQty: tx.PaymentTransaction.Amount + tx.Fee,
sender: account,
})
} else {
// either a regular receive or a receive and close-remainder-to but to same account.
records = appendPostFilter(records, ExportRecord{
blockTime: blockTime,
txid: tx.Id,
receiver: tx.PaymentTransaction.Receiver,
sentQty: tx.PaymentTransaction.Amount + tx.PaymentTransaction.CloseAmount + tx.ClosingAmount + tx.Fee,
sender: account,
})
}
}
```

All that's left are non-payment transactions, which are participation key registration (for running a participation node, which is highly recommended!), and ASA (Algorand Standard Asset) operations. This solution will skip ASA operations for now.

For these remaining types, we just want to ensure we handle processing rewards that may have been added as part of the new on-chain transaction, as well as any fees.

``` go
case "keyreg", "acfg", "afrz", "axfer":
// Just track the fees and rewards for now as a result of the transaction
// Ignore the ASA activity.
if tx.AssetTransferTransaction.Receiver == account {
rewards += tx.ReceiverRewards
}
if tx.Sender == account {
records = appendPostFilter(records, ExportRecord{
blockTime: blockTime,
txid: tx.Id,
sentQty: tx.Fee,
sender: account,
})
rewards = tx.SenderRewards
}
default:
log.Fatalln("unknown transaction type:", tx.Type)
}
```

All that's left is adding a transaction for any 'rewards' that may have been added to this account as part of receiving or sending. Because we want to ensure the balance tracked by these tracking sites includes the full balance, we fake a timestamp for the reward transaction by using the block timestamp and subtracting 1 second. This should be sufficient.

``` go
// now handle rewards (effectively us receiving them - either we sent and received pending rewards
// or received a payment and also were assigned the pending rewards. Treat both as a standalone receive.
// The transaction is exported with a timestamp 1 second before the real on-chain transaction
// so the extra balance is there for deductions and we don't go negative. The transaction is defined as a
// rewards so it can be tracked as income by the tax tracker.
if rewards != 0 {
// Apply rewards 'first' (earlier timestamp)
records = appendPostFilter(records, ExportRecord{
blockTime: blockTime.Add(-1 * time.Second),
txid: tx.Id,
reward: true,
recvQty: rewards,
receiver: account,
})
}
return records
}
```

All that remains is the actual 'export' code. Below is the main `WriteRecord` method for Koinly.

As you can see, it's quite simple. A number of fields available in `ExportRecord` are ignored (like sender and receiver) because Koinly has no provision for CSV files which contain the full on-chain details. The code is primitive, but straight-forward. The `algoFmt` function is a simple helper in `exporter/exporter.go` which expresses microAlgos as algos (ie: 1000 becomes .001). For koinly, we use the transaction id as the `TxHash` field. Koinly uses this to uniquely identify the imported record so it won't import it twice. Because we synthesize 'reward' transactions, we simply prepend 'reward-' to the transaction id for reward transactions. We also use the special 'staking' label for rewards so Koinly treats the reward as a staking reward (income).

``` go
func (k *koinlyExporter) WriteRecord(writer io.Writer, record ExportRecord) {
//a Date,Sent Amount,Sent Currency,Received Amount,Received Currency,Fee Amount,Fee Currency,Net Worth Amount,Net Worth Currency,Label,Description,TxHash
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
``` 

# Building the program

This solution assumed you're a go developer, or at least somewhat familiar with it. Even if not, if you'd just like to use the program it's simple to build yourself.

First, download the go compiler from https://golang.org/

Run the following commands:

1. `git clone https://github.com/pbennett/algo-export.git`
2. `cd algo-export`
3. `go build`

You will be left with a binary called `algo-export` in the current directory.

To build the program, simply clone the code into a new directory: 

`git clone https://github.com/pbennett/algo-export.git` 

and from that new directory, either: 

- run `go build` (to create the `algo-export` in the current directory), 
- or, run `go install` to create `algo-export` in `~/go/bin` (the default if you haven't defined your own explicit GOPATH)

Running `algo-export -h` will show you the options as discussed in the beginning:

```text
algo-export -h
Usage of algo-export:
-a value
Account or list of comma delimited accounts to export
-api string
Optional API Key for local indexer, or for PureStake
-f string
Format to export: [cointracker, koinly] (default "cointracker")
-o string
output directory path for exported files
-p	Use PureStake API - ignoring -s argument
-s string
Index server to connect to (default "localhost:8980")
```

Refer back to [Example use](#example-use) for examples of how the program is used.

Enjoy!
