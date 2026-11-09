# chain-reader-cli

A Go command-line tool that reads the latest block and account balances from an Ethereum RPC endpoint using go-ethereum

Part of my [learning roadmap](https://github.com/Syriacode/Mini-Blockchain/blob/main/ROADMAP.md).

| | |
|---|---|
| Language | go |
| Claude Code technique practiced | learning a new language with Claude |

## Run it

```bash
# Both need Anvil running on 127.0.0.1:8545, or pass --rpc <url>
go run . block
go run . balance 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
go test ./...
go vet ./...
gofmt -l .
```

## Design decisions

- **Integer wei, no floats.** Balances and fees stay as `*big.Int` wei end to end. `FormatETH` turns wei into an ETH string with integer division and remainder, so there's no float rounding.
- **`ChainReader` interface.** Defined in `reader.go`, next to the code that uses it. It lists only the four methods this tool needs, so `*ethclient.Client` satisfies it and tests can pass in a fake.
- **`HeaderByNumber` + `TransactionCount` instead of `BlockByNumber`.** The header plus a count is all the output needs. Downloading every transaction body would be wasted bandwidth.
- **One hash for both calls.** The header is fetched once, its hash is computed once, and that hash is what `TransactionCount` asks about. The count is guaranteed to belong to the same block as the header, even if a new block lands between the two calls.
- **Validate the address first.** `common.IsHexAddress` rejects bad input before any RPC request goes out. (Dialing an HTTP endpoint is lazy and doesn't touch the network.)
- **10 second timeout.** One `context.WithTimeout` covers dialing and every RPC call, so a dead node can't hang the CLI.
- **`run()` returns an exit code.** `main` is just `os.Exit(run(...))`. `run` takes args and writers so tests can drive it. It returns 0 on success, 1 on a runtime failure, and 2 on wrong usage.

## Known limitations

- Only reads the latest block. You can't ask for a specific block number or hash.
- `latest` can be reorged. A production service would use the `safe` or `finalized` tag.
- It trusts the RPC provider and doesn't verify Merkle proofs (e.g. `eth_getProof`).
- No retries, backoff or rate-limit handling.
- `common.IsHexAddress` doesn't verify the EIP-55 mixed-case checksum, so a mistyped checksummed address is still accepted.
- The block hash is computed locally from the header. On a chain whose headers carry extra fields go-ethereum doesn't know about, the computed hash could differ from the node's, and `TransactionCount` would ask for a block the node doesn't recognise.
