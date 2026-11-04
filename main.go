package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// defaultRPC is a local node such as Anvil, used when --rpc and RPC_URL are both unset.
const defaultRPC = "http://127.0.0.1:8545"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the whole program. It returns the exit code: 0 success, 1 runtime failure, 2 wrong usage.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("chain-reader", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { printUsage(stderr) }

	rpcDefault := os.Getenv("RPC_URL")
	if rpcDefault == "" {
		rpcDefault = defaultRPC
	}
	rpcURL := fs.String("rpc", rpcDefault, "Ethereum JSON-RPC endpoint")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0 // -h or --help: usage was printed on purpose
		}
		return 2 // the flag package already printed the problem and the usage
	}

	// Check the command before connecting, so bad usage never touches the network.
	rest := fs.Args()
	switch {
	case len(rest) == 1 && rest[0] == "block":
	case len(rest) == 2 && rest[0] == "balance":
	default:
		printUsage(stderr)
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, *rpcURL)
	if err != nil {
		fmt.Fprintf(stderr, "error: connect to %s: %v\n", *rpcURL, err)
		return 1
	}
	defer client.Close()

	if rest[0] == "block" {
		err = printBlock(ctx, client, stdout)
	} else {
		err = printBalance(ctx, client, rest[1], stdout)
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: chain-reader [--rpc URL] block")
	fmt.Fprintln(w, "       chain-reader [--rpc URL] balance <address>")
}

// printBlock prints a summary of the latest block.
func printBlock(ctx context.Context, r ChainReader, w io.Writer) error {
	b, err := LatestBlock(ctx, r)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "%-14s%s\n", "chain ID:", b.ChainID)
	fmt.Fprintf(w, "%-14s%s\n", "block:", b.Number)
	fmt.Fprintf(w, "%-14s%s\n", "hash:", b.Hash.Hex())
	fmt.Fprintf(w, "%-14s%s\n", "time (UTC):", b.Time.Format(time.RFC3339))
	fmt.Fprintf(w, "%-14s%d\n", "transactions:", b.TxCount)
	fmt.Fprintf(w, "%-14s%d\n", "gas used:", b.GasUsed)
	if b.BaseFee != nil {
		fmt.Fprintf(w, "%-14s%s wei\n", "base fee:", b.BaseFee)
	}
	return nil
}

// printBalance prints an account's balance in wei and in ETH.
func printBalance(ctx context.Context, r ChainReader, address string, w io.Writer) error {
	wei, err := Balance(ctx, r, address)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "%s wei\n", wei)
	fmt.Fprintf(w, "%s ETH\n", FormatETH(wei))
	return nil
}
