# Interview questions: chain-reader-cli

## Q1. Why are balances handled with big.Int, and what went wrong when you used float64?

Answer: Ethereum balances are whole numbers of wei (1 ETH = 10^18 wei), stored onchain as `uint256`, so the
natural Go type is `*big.Int`, which holds an integer of any size exactly. `int64` is too small: it tops out
around 9.22 ETH when counted in wei. `float64` is big enough in range but only has a 53-bit mantissa, which is
about 15–17 significant digits, so it silently rounds anything longer. When I rewrote `FormatETH` to use
`float64` on purpose, five table cases still passed, but `1000000000000000001` wei came out as `"1"` instead of
`"1.000000000000000001"`. The large case `123456789012345678901234567890` came out as `"123456789012.34567"`
instead of `"123456789012.34567890123456789"`. The dangerous part is that round numbers look correct, so the
bug only shows up on the exact amounts nobody checks by eye. The real `FormatETH` uses `QuoRem` to split wei
into whole ETH and leftover wei, then zero-pads the remainder to 18 digits, so no float is involved anywhere.

## Q2. Walk me through `chain-reader balance 0x...` from command line to printed number.

Answer: `main` calls `run`, which parses `--rpc` (default from `RPC_URL`, else `http://127.0.0.1:8545`).
A bad flag or the wrong number of arguments prints usage and returns exit code 2 before anything touches the
network. `run` then creates one 10-second `context.WithTimeout` and calls `ethclient.DialContext`. For an HTTP
URL that only builds the client, but an unsupported URL scheme fails here with exit code 1. `Balance` checks
the string with `common.IsHexAddress` and returns `ErrInvalidAddress` before any request goes out. It then
calls `BalanceAt(ctx, addr, nil)`, and `nil` becomes the `"latest"` block tag in a JSON-RPC request:
`{"method":"eth_getBalance","params":["0xf39F...","latest"]}`. The node replies with a hex quantity such as
`"0x21e19e0c9bab2400000"`, which go-ethereum decodes into a `*big.Int`. That request can fail with a refused
connection, a timeout (`context deadline exceeded`), a JSON-RPC error from the node such as a rate limit, or a
response that doesn't decode as hex. Every one of those is wrapped as `get balance: ...`, printed to stderr,
and returns exit code 1. On success, `printBalance` prints the raw wei and the `FormatETH` string, and the
program exits 0.

## Q3. Why does LatestBlock use HeaderByNumber and TransactionCount instead of BlockByNumber?

Answer: The output only needs header fields (number, time, gas used, base fee) plus a transaction count.
`BlockByNumber` calls `eth_getBlockByNumber` with full transactions, which can be hundreds of kilobytes per
block on mainnet, while `HeaderByNumber` asks for the header only and `TransactionCount` returns a single
number from `eth_getBlockTransactionCountByHash`. `BlockByNumber` also has to decode every transaction into
go-ethereum's `types.Transaction`. L2s like OP Stack chains and Arbitrum add their own transaction types (for
example OP's deposit type `0x7e`), and upstream go-ethereum rejects those with "transaction type not
supported", so the whole call fails even though we never needed the transactions. Asking only for a count
avoids decoding them at all. Both calls use one hash: I fetch the header for "latest", compute `header.Hash()`
once, and ask for the transaction count of that exact hash. If I asked for "latest" twice, a new block could
arrive between the calls, and I'd print block N's header with block N+1's transaction count.

## Q4. Is the "latest" block final?

Answer: No. "latest" is just the head of the chain the node currently follows, and a reorg can replace it
when the network settles on a different fork, so a transaction or balance you saw can disappear or change.
Since the Merge, Ethereum also exposes two safer tags. `safe` is a block the beacon chain has justified, which
is very unlikely to be reorged. `finalized` is a block whose checkpoint has been finalized, which takes about
two epochs (roughly 13 minutes), and reverting it would require a third of all staked ETH to be slashed. This
tool is fine reading "latest" because it's a read-only display. A service that credits customer deposits
should only credit once the deposit's block is at or below `finalized` (or N confirmations on chains without
finality, and L1 finality for most L2s), record the block hash, and recheck that hash so it notices a reorg
before finality. Credits should also be idempotent, keyed by transaction hash and log index, so reprocessing a
block never pays someone twice.

## Q5. If this became a production service reading balances for a stablecoin team, what would you change?

Answer: First, stablecoin balances are ERC-20 state, so I'd read `balanceOf` through `eth_call` (or index
`Transfer` events) instead of `eth_getBalance`, and pin every read to a specific `finalized` block number so
all reads agree with each other. I wouldn't trust one RPC provider. I'd use two or more (ideally including our
own node), compare block hashes and results, and alert when they disagree. For high-value numbers I'd check
`eth_getProof` against a state root we trust. Each call would get its own timeout, plus retries with
exponential backoff and jitter, but only for errors that are worth retrying, like timeouts, 5xx and 429, and
never for a bad address. For rate limits I'd add a client-side limiter, batch requests and cache results by
block number, since a finalized block never changes. Metrics would cover latency and error rate per provider
and method, retry counts, and how far each provider's head lags the others. The `ChainReader` interface makes
this clean: retry, metrics and failover can each be a wrapper that implements `ChainReader` and holds another
one inside it, so `LatestBlock` and `Balance` don't change at all, and the same fakes still test them.
