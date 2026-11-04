# Week 1 Wed — Chain reader CLI: latest block, balances (step-by-step guide)

**Goal:** build a small Go command-line tool that talks to any Ethereum node over JSON-RPC, prints the latest
block, and reads an address's balance without losing a single wei. You'll write one core function by hand, and
you'll run the finished tool against a local chain and then against real mainnet.

**Why this project:** this is your first Go project, and Go is what most blockchain backend teams use. Nearly
every service on the roadmap starts this way: connect to a node, read data, handle failure. The Week 2 indexer
and the Week 3 reserve service both grow out of what you build today.

| Job requirement | Where you practice it |
|---|---|
| Go backend services (Coinbase, Anchorage, Robinhood) | Every step |
| Reading onchain data through RPC | Steps 3, 6 |
| Correct money handling | Step 2 (integer math, float break test) |
| Testable services, *well-tested code* | Steps 2, 3, 4 (table tests, fakes, simulated chain) |
| *Uses generative AI responsibly, with human oversight* | "You check" in every step, Step 2 hand-written function |

**Claude Code technique today: learning a new language with Claude.** Claude can write Go faster than you can
read it, which is exactly the trap. Today you practice four habits that turn it into a tutor:

1. **Put your teaching style in `CLAUDE.md`,** so every session explains Go by comparing it to what you know.
2. **Predict first:** say what you think code does, then let Claude correct you.
3. **Type the core function yourself,** and ask Claude for hints and review, not the answer.
4. **Read the library's source** with Claude instead of trusting a summary of it.

**Time:** about 90 minutes, including the first go-ethereum build. It's free. The only real
network you touch is read-only, and no real wallet or key is involved.

---

## How to use this guide with Claude Code

Every step has the same five parts:

- **Learn:** the concept, in a few sentences. Read it before prompting.
- **Prompt:** what to type into Claude Code. Use **plan mode** (`Shift+Tab` until the status bar shows
  `⏸ plan mode on`), read the plan, then approve it.
- **You check:** what *you* verify before accepting. Each check names the file and function to open.
- **Done when:** a command whose output proves the step works.
- **Commit:** the commit message to use. Ask Claude to commit, or run it yourself.

A leading `!` in the Claude Code prompt runs a shell command, as in `!go test ./...`.

Your permission rules already allow `go build`, `go test`, `go vet` and `gofmt -l`. Claude will ask before
`go get`, `go mod tidy` and `go run`. Approve those.

---

## Step 0 — Check Go, create the repo, add this guide (10 min)

**Learn.** Go compiles to a single program and comes with its formatter (`gofmt`), test runner (`go test`),
and dependency manager (`go.mod`), so there's almost nothing to set up. Your `/new-project` skill already has
Go templates.

**Do it:** in Git Bash, check your tools:

```bash
go version
anvil --version
cast --version
```

You need Go **1.21 or newer**. go-ethereum needs Go 1.25, but from 1.21 on, Go downloads the version a project
needs on its own. If `go` isn't found, run this, then open a **new** Git Bash window and check again:

```bash
winget install GoLang.Go
```

Now go to the folder where you keep your project repos (**not** inside another git repo), then:

```bash
claude
```

**Prompt:**
> /new-project Week 1 Wed – Chain reader CLI

When it proposes a name, accept **`chain-reader`**. Confirm the language is **`go`**. Approve the
`gh repo create` and `git push` prompts when they appear.

When it finishes and reports green CI, quit Claude Code (`/exit`), then:

```bash
cd chain-reader
```

Save **this file** as `GUIDE.md` in the `chain-reader` folder. Then:

```bash
claude
```

**Prompt:**
> Commit GUIDE.md.

**You check:** `git log --oneline` shows two commits. Open `go.mod`: the first line must be
`module github.com/Syriacode/chain-reader`.

**Done when:** `go test ./...` prints a line starting with `ok`.

**Commit:** `Add GUIDE.md`

---

## Step 1 — Turn Claude into your Go tutor (10 min)

**Learn (Claude Code).** Claude reads `CLAUDE.md` at the start of every session. It's normally used for project
rules, but it's just as good for saying *how you want to be taught*. A few lines here apply to every prompt, so
you never have to retype "please explain this".

**Prompt (no plan mode needed):**
> Add this section to @CLAUDE.md, word for word, under "How to work with me":
>
> ### Learning Go
> - I know Python and Solidity well, and I'm new to Go. Explain Go ideas by comparing them to those two.
> - When you introduce a Go feature for the first time, explain it in one or two sentences.
> - When I ask for a hint, give one hint only. Don't write the solution.
> - When I review code, ask me what I expect it to do before you tell me.

**Now use it.** Open `main.go` and `main_test.go` from the template in your editor, and read them yourself.
Make a guess at each question below **before** you send it.

**Prompt:**
> Using @main.go and @main_test.go, explain: (1) what `package main` and `func main` mean and how that compares
> with a Python script, (2) why `Println` has a capital P and `greeting` doesn't, (3) what `:=` does here,
> (4) how `go test` finds `TestGreeting`, and (5) what the `%q` in `t.Errorf` prints.
> Then ask me 3 quick questions to check that I understood.

**You check:** answer Claude's 3 questions in your own words. Your notes should say that in Go, **a capital
first letter means exported** (public outside the package), and that `:=` declares a new variable and infers
its type.

**Done when:** `git diff --stat` shows only `CLAUDE.md` changed.

**Commit:** `Add Go learning instructions to CLAUDE.md`

---

## Step 2 — Write `FormatETH` yourself (20 min)

**Learn (blockchain).** Balances are whole numbers of **wei**, and 1 ETH = 10¹⁸ wei. A node returns them as
256-bit integers. `int64` tops out around 9.2 ETH in wei, so it's too small, and `float64` has only about
15–17 significant digits, so it silently rounds. Money code uses **`*big.Int`**, Go's whole number of any size.

**Learn (Go).**
- `big.Int` has no `+` or `/` operators. You call methods, and the first value receives the result:
  `new(big.Int).Add(a, b)` means "make a new big.Int and store a + b in it".
- Go tests are usually **table-driven**: a list of cases and one loop, with `t.Run` naming each case.
- A file needs a **`package` line**, and all `.go` files in one folder share it. Functions in `format.go` are
  visible in `main.go` with no import.

**Prompt (plan mode):**
> Create format_test.go only. Don't create format.go. It's a table-driven test, TestFormatETH, for
> `func FormatETH(wei *big.Int) string`, using t.Run for each case. Store wei as a string in the table and
> parse it with `new(big.Int).SetString(value, 10)`. Cases (name, wei → want):
> - zero: 0 → "0"
> - one wei: 1 → "0.000000000000000001"
> - half ETH: 500000000000000000 → "0.5"
> - one ETH: 1000000000000000000 → "1"
> - one ETH and one wei: 1000000000000000001 → "1.000000000000000001"
> - anvil account: 10000000000000000000000 → "10000"
> - beyond float64: 123456789012345678901234567890 → "123456789012.34567890123456789"
>
> Run go test ./... and show me the failure.

**You check:** the output contains `undefined: FormatETH`. That's a compile error, not a test failure. Unlike
Python, Go checks the whole package before running anything.

**Now type the code yourself.** Don't ask Claude to write it. Create `format.go` in your editor and start from this:

```go
package main

import (
	"math/big"
	"strings"
)

// weiPerETH is 10^18. Amounts are kept in integer wei, so no precision is lost.
var weiPerETH = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

// FormatETH turns a wei amount into a decimal ETH string, using integer math only.
func FormatETH(wei *big.Int) string {
	// your code here
}
```

The plan:

1. Split `wei` into whole ETH and leftover wei:
   `quotient, remainder := new(big.Int).QuoRem(wei, weiPerETH, new(big.Int))`
2. If `remainder.Sign() == 0`, there's no fraction, so return `quotient.String()`.
3. Turn the remainder into a string. It must be **exactly 18 digits**, so add zeros on the left. Use
   `strings.Repeat("0", n)`, and work out `n` from `len(...)`.
4. Remove the trailing zeros with `strings.TrimRight(s, "0")`.
5. Join the parts with `"."`.

Run it with:

```text
!go test -run TestFormatETH -v ./...
```

If you get stuck, ask for **one hint**:
> Give me one hint for the failing test. Don't write the code.

**Once it passes, get a review:**
> Review @format.go. Before telling me anything, ask me what happens in FormatETH if the remainder has
> fewer than 18 digits, and why the "one wei" test would catch a mistake there. Then tell me one thing a
> senior Go reviewer would change. Don't edit the file.

**You check:**
- `format.go`, function `FormatETH`: there's no `float64` anywhere.
- `format_test.go`, case `one ETH and one wei`: you can explain why this is the case a float can't pass.
- Run `!gofmt -l .` and make sure it prints nothing. If it lists `format.go`, run `!gofmt -w format.go`.

**Done when:** `go test -run TestFormatETH -v ./...` shows all 7 sub-tests as `--- PASS`.

**Commit:** `Add FormatETH with table-driven tests`

Commit **now**, before the experiment below.

**Break it on purpose (5 min).** Predict first: which tests will fail?

**Prompt (no plan mode needed):**
> Break @format.go on purpose: rewrite FormatETH to convert wei to a float64, divide by 1e18, and format it with
> strconv.FormatFloat(value, 'f', -1, 64). Keep weiPerETH. Don't touch format_test.go. Run the tests and explain
> the result.

You should see:

```text
FormatETH(1000000000000000001) = "1", want "1.000000000000000001"
FormatETH(123456789012345678901234567890) = "123456789012.34567", want "123456789012.34567890123456789"
```

The other five cases still pass. A float version looks right on round numbers and only loses money on the
amounts nobody checks by eye. Now undo it, and confirm:

```text
!git restore format.go
!git status
!go test ./...
```

**Done when:** `git status` says `nothing to commit, working tree clean` and the tests print `ok`.

**Commit:** nothing for the experiment.

---

## Step 3 — Read the chain, behind an interface (15 min)

**Learn (blockchain).**
- A node speaks **JSON-RPC**: you POST JSON naming a method such as `eth_getBalance`, and it answers in JSON.
  Numbers come back as hex strings like `"0x1bc16d674ec80000"`.
- **`go-ethereum`** (geth) is the most widely used Ethereum client, and it's written in Go. Its `ethclient`
  package wraps those JSON-RPC calls in typed Go methods.
- **The block header** is the summary of a block: number, parent hash, timestamp, gas used, base fee.
- The **chain ID** says which network you're on. Printing it stops you from reading the wrong chain by accident.

**Learn (Go).**
- Functions usually return `(result, error)`, and callers check `if err != nil`. There are no exceptions.
  `fmt.Errorf("get balance: %w", err)` adds context and keeps the original error findable with `errors.Is`.
- **Interfaces are implicit:** any type with the right methods fits, with no `implements` keyword. That lets you
  define a small interface where you *use* it, and hand your function a fake in tests.
- **`context.Context`** carries a deadline, so a hung node can't freeze your program forever.

**Prompt 1 (plan mode):**
> Run `go get github.com/ethereum/go-ethereum@v1.17.10`. Then create reader.go in package main with:
> - `var ErrInvalidAddress = errors.New("invalid address")`
> - interface `ChainReader` with exactly the four methods I need, using the same signatures as `*ethclient.Client`:
>   ChainID, HeaderByNumber, TransactionCount, BalanceAt
> - struct `BlockSummary` with ChainID *big.Int, Number *big.Int, Hash common.Hash, Time time.Time, TxCount uint,
>   GasUsed uint64, BaseFee *big.Int (with a comment that BaseFee is nil on chains or blocks without EIP-1559)
> - `func LatestBlock(ctx context.Context, r ChainReader) (BlockSummary, error)`: call ChainID, then HeaderByNumber
>   with nil, then compute `hash := header.Hash()` and call TransactionCount(ctx, hash). Convert the time with
>   `time.Unix(int64(header.Time), 0).UTC()`. Wrap each error with %w and these messages: "get chain ID",
>   "get latest header", "count transactions". Don't use BlockByNumber.
> - `func Balance(ctx context.Context, r ChainReader, address string) (*big.Int, error)`: reject bad input with
>   common.IsHexAddress before any network call, returning an error that wraps ErrInvalidAddress and shows the
>   input with %q. Then call BalanceAt with nil. Wrap its error with "get balance".
>
> Then run `go mod tidy`, go build ./..., and go vet ./.... The first build compiles go-ethereum and can take a
> few minutes.

You may see `go: downloading go1.25.0`. That's Go fetching the version go-ethereum needs, and `go.mod` is
updated to match.

**Prompt 2 (plan mode):**
> Create reader_test.go with a hand-written `fakeReader` struct with fields headerErr error, balanceErr error,
> balanceAddress common.Address. Its methods use pointer receivers. ChainID returns 31337. HeaderByNumber returns
> headerErr if set, otherwise `&types.Header{Number: big.NewInt(7), Time: 1700000000, GasUsed: 21000}`.
> TransactionCount returns 3. BalanceAt stores the account in balanceAddress, returns balanceErr if set, otherwise 5.
> Tests:
> 1. TestLatestBlockSummarizesHeader: checks number 7, 3 transactions, gas used 21000, time
>    2023-11-14T22:13:20Z, and BaseFee is nil
> 2. TestLatestBlockWrapsNodeError: with headerErr set to a "connection refused" error, errors.Is finds it in the result
> 3. TestBalanceRejectsBadAddressWithoutCallingNode: for "", "hello", "0x1234" and
>    "0xZZZZ6BF26964aF9D7eEd9e03E53415D37aA96045", errors.Is finds ErrInvalidAddress and balanceAddress stays zero
>
> Run gofmt -l ., then go test -v ./... and show me the output.

**Read the source instead of guessing.** go-ethereum is on your machine now.

**Prompt 3 (no plan mode needed):**
> Find the go-ethereum source with `go list -m -f '{{.Dir}}' github.com/ethereum/go-ethereum`. In ethclient/ethclient.go,
> read BlockByNumber, HeaderByNumber, TransactionCount, BalanceAt and toBlockNumArg. Show me the lines, then answer:
> (1) which JSON-RPC method does each use? (2) what does a nil block number become? (3) what's the difference
> between the last argument of eth_getBlockByNumber in BlockByNumber and in HeaderByNumber? (4) what does
> `ErrTxTypeNotSupported` in core/types/transaction_marshalling.go mean for BlockByNumber on a chain with custom
> transaction types? Don't change any code.

**You check:**
- `reader.go`, `ChainReader`: exactly four methods. Ask: *"Why do I define this interface here instead of using
  ethclient.Client directly? How does the compiler know `*ethclient.Client` fits?"*
- `reader.go`, `LatestBlock`: `HeaderByNumber` gets `nil`, and `TransactionCount` gets the hash of *that same*
  header. Ask: *"What could go wrong if I asked for the count of 'latest' in a second call instead?"*
- `reader.go`, `Balance`: the `IsHexAddress` check comes first. Ask: *"Does IsHexAddress verify the checksum?
  If not, what could slip through?"* (Keep the answer for Known limitations.)
- `reader_test.go`, `fakeReader`: the methods use `(f *fakeReader)`. Ask: *"Why would a value receiver break
  the balanceAddress check?"*
- Prompt 3's answer: you can say why this tool calls `HeaderByNumber` rather than `BlockByNumber`. Keep it for
  Interview Q3.

**Done when:** `go test -v ./...` shows `--- PASS` for the three new tests, and `gofmt -l .` prints nothing.

**Commit:** `Add block and balance reader behind ChainReader interface`

---

## Step 4 — Test against a real chain in memory (5 min)

**Learn.** A fake only proves your code matches *your beliefs* about the node. go-ethereum's **simulated backend**
runs a real in-memory chain inside your test: no network, no Anvil, so it works in CI. Your code can't tell the
difference because its `Client()` also fits `ChainReader`. That's the interface paying off.

**Prompt (plan mode):**
> In @reader_test.go add TestBalanceAgainstSimulatedChain using github.com/ethereum/go-ethereum/ethclient/simulated.
> Fund address 0x000000000000000000000000000000000000bEEF with 3 ETH via
> `simulated.NewBackend(types.GenesisAlloc{addr: {Balance: funded}})`, where funded is 3 times weiPerETH. Defer
> backend.Close(), and call backend.Commit() once to mine a block. Then using backend.Client(), check that Balance
> returns exactly funded and LatestBlock returns block number 1.
> Run go mod tidy, gofmt -l ., and go test -v ./...

**You check:** in `reader_test.go`, the new test passes `backend.Client()` to the same `Balance` and `LatestBlock`
you use in production. No code in `reader.go` changed: confirm with `!git diff --stat reader.go` (it prints nothing).

**Done when:** `go test -v ./...` shows `--- PASS: TestBalanceAgainstSimulatedChain`.

**Commit:** `Add simulated chain test`

---

## Step 5 — The command-line program (10 min)

**Learn (Go).**
- `main` can't return anything, and `os.Exit` stops the program immediately, skipping `defer`. So the usual
  pattern is a small `main` that calls `run(args, stdout, stderr) int`, which tests can call directly.
- `defer client.Close()` runs when the function returns, like Python's `with`.
- The `flag` package stops at the first non-flag word, so `--rpc` must come **before** the subcommand.
- Exit codes matter to scripts: `0` success, `1` runtime failure, `2` wrong usage.

**Prompt (plan mode):**
> Rewrite @main.go and delete greeting() and its test. New design:
> - `func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }`
> - `func run(args []string, stdout, stderr io.Writer) int`
> - Use flag.NewFlagSet("chain-reader", flag.ContinueOnError) with SetOutput(stderr) and a string flag `rpc`. Its
>   default comes from env RPC_URL, or else const defaultRPC = "http://127.0.0.1:8545".
> - Usage text printed to stderr, exactly these two lines:
>   `usage: chain-reader [--rpc URL] block`
>   `       chain-reader [--rpc URL] balance <address>`
> - Valid commands are `block` with no extra arguments, and `balance <address>` with exactly one. Anything else
>   prints usage and returns 2. Check this before connecting.
> - Create a context with a 10 second timeout (defer cancel), call ethclient.DialContext (defer client.Close).
>   On any error print `error: ...` to stderr and return 1.
> - `printBlock(ctx, r ChainReader, w io.Writer) error` prints aligned lines: chain ID, block, hash, time (UTC) in
>   RFC3339, transactions, gas used, and `base fee: <n> wei` only if BaseFee isn't nil.
> - `printBalance(ctx, r ChainReader, address string, w io.Writer) error` prints `<wei> wei` and then `<eth> ETH`
>   using FormatETH.
>
> Replace @main_test.go with TestRunRejectsBadUsage: for args {}, {"nope"}, {"block", "extra"} and {"balance"},
> run returns 2 and stderr contains "usage:". Use bytes.Buffer for stdout and stderr. Run gofmt -l ., go vet ./...,
> go test -v ./...

**You check:**
- `main.go`, `run`: usage is checked before `DialContext`, and both `defer` lines come after their error check.
  Ask: *"Why would `defer client.Close()` be wrong above the `if err != nil`?"*
- `main.go`, `printBlock`: it takes `ChainReader`, not `*ethclient.Client`.
- Ask: *"`client` is an `*ethclient.Client` and printBlock wants a ChainReader. Where does the conversion happen?"*

**Done when:**

```bash
gofmt -l . && go vet ./... && go test ./...
```

prints only an `ok` line. All 6 test functions pass in `go test -v ./...`.

**Commit:** `Add block and balance commands`

---

## Step 6 — Run it for real: Anvil, then mainnet (10 min)

**Learn.** Anvil is your local chain. It gives you 10 funded accounts whose keys are **public and well-known**.
Never send real money to them. It mines a block only when a transaction arrives, so it starts at block 0.

**Failure first.** With no node running, in the Claude Code prompt:

```text
!go run . block
```

You should see `error: get chain ID: Post "http://127.0.0.1:8545": ...`, ending with Windows' wording for a refused
connection (`actively refused it`), then `exit status 1`. (That last line comes from `go run`.) The wrapped
message tells you exactly which call failed.

**Start the chain.** Open a **second** Git Bash window and leave it running:

```bash
anvil
```

**Back in your first Git Bash window** (quit Claude Code with `/exit`, or use the `!` prefix), in `chain-reader`:

```bash
go run . block
```

Expect `chain ID: 31337`, `block: 0`, `transactions: 0`, `gas used: 0`.

```bash
go run . balance 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
```

Expect `10000000000000000000000 wei` and `10000 ETH`.

**Send a transaction with `cast`,** using Anvil's first public test key:

```bash
cast send 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --value 1.5ether \
  --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
  --rpc-url http://127.0.0.1:8545
```

Look for a `status` line reading `1 (success)`. Now:

```bash
go run . block
go run . balance 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
go run . balance 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
cast balance 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
```

Expect `block: 1`, `transactions: 1`, `gas used: 21000`. The receiver has exactly `10001.5 ETH`, and your
wei value matches `cast balance`. The sender has a bit **less** than 9998.5 ETH (about `9998.49997...`).
That missing amount is the gas fee, and only an exact integer shows it to the wei.

**Now read real mainnet,** by changing nothing but the URL:

```bash
go run . --rpc https://ethereum-rpc.publicnode.com block
cast block-number --rpc-url https://ethereum-rpc.publicnode.com
go run . --rpc https://ethereum-rpc.publicnode.com balance 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
```

Expect `chain ID: 1`, a block number above 26 million, usually hundreds of transactions, and a block time within the last
minute. `cast` may show a slightly higher number, since a new block arrives every 12 seconds. That address
is Vitalik Buterin's public one. It's a free public node, so if you get a rate-limit error, wait a minute and retry.

**Bad input:**

```bash
go run . balance hello
```

Expect `error: invalid address: "hello"`.

**You check:** in `main.go`, nothing about Anvil or mainnet is hard-coded except `defaultRPC`. That same binary can
read any EVM chain, because the protocol is the same.

**Done when:** every "Expect" above matched. Close the Anvil window (`Ctrl+C`). Anvil keeps its chain in memory,
so a restart goes back to block 0.

**Commit:** nothing. No files changed in this step.

---

## Step 7 — README, INTERVIEW.md, push, green CI (10 min)

Start Claude Code again if you quit it.

**Prompt 1 (README and CLAUDE.md):**
> In @README.md and @CLAUDE.md, replace the "go run ." line in the run commands with these two:
> `go run . block` and `go run . balance 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266`. Add a comment saying both
> need Anvil running, or --rpc <url>.
>
> In @README.md replace the "Design decisions" TODO with short bullets on: integer wei with big.Int and no floats,
> the ChainReader interface and where it's defined, HeaderByNumber plus TransactionCount instead of
> BlockByNumber, one hash for both calls, validating the address before any network call, a 10 second context
> timeout, and the run() function with exit codes.
> Replace the "Known limitations" TODO with: only reads the latest block; "latest" can be reorged, so a production
> service would use the safe or finalized tag; it trusts the RPC provider and doesn't verify proofs; no retries or
> rate-limit handling; IsHexAddress doesn't verify the EIP-55 checksum; the hash is computed locally from the
> header, so a chain with extra header fields could produce a hash the node doesn't recognise.
> Don't touch the "How I used Claude Code" section.

**Prompt 2 (INTERVIEW.md):**
> Fill in @INTERVIEW.md with these 5 questions and short model answers (5–8 sentences each), based on this code:
> 1. Why are balances handled with big.Int, and what went wrong when I used float64? Give the numbers from my test.
> 2. Walk me through what happens from `chain-reader balance 0x...` to the printed number, including the JSON-RPC
>    method, the "latest" tag, the hex response, and every place it can fail.
> 3. Why does LatestBlock use HeaderByNumber and TransactionCount instead of BlockByNumber? Cover payload size,
>    custom transaction types on L2s, and why both calls use one hash.
> 4. Is the "latest" block final? Explain reorgs, the safe and finalized tags, and what a service that credits
>    customer deposits should do differently from this tool.
> 5. If this became a production service reading balances for a stablecoin team, what would you change? Cover
>    trusting a single RPC provider, retries and timeouts, rate limits, metrics, and how the ChainReader
>    interface helps.

**You check:** read every answer. Rewrite at least one in your own words. Q2 especially: you should be able to say
it out loud without looking. If an answer says something you can't explain, ask Claude about it until you can.

**Write "How I used Claude Code" yourself** in `README.md`: what the `CLAUDE.md` tutor rules changed, how writing
`FormatETH` by hand went, what you got wrong in your predictions, and anything Claude got wrong.

**Done when:**

```bash
gofmt -l . && go vet ./... && go test ./...
```

prints only an `ok` line. Then commit and push (Claude will ask before `git push`):

**Commit:** `Add README design notes and interview Q&A`

Then check CI:

```text
!gh run list --branch main --limit 1
```

Wait until the newest run shows success (a ✓, or `completed` / `success`). Re-run the command every few seconds
while it says `in_progress`. CI's first run compiles go-ethereum, so it takes a few minutes. If it fails, open it
with `!gh run view <id> --log-failed`, then ask Claude to fix it.

---

## When you're done

Tell the roadmap chat **"done"**. It will review the repo, run the tests, and tick this row in the roadmap.

## Level up

**Tell Claude how to teach you, once, in `CLAUDE.md`.** When you're learning a language, the fastest way to learn
nothing is to approve code you can't read. Write down what you already know, ask for comparisons to it, predict
before you ask, and hand-type the one function where the real idea lives. Then copy your "Learning Go" section into
the Rust project in Week 4, and just change the language name.
