package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ErrInvalidAddress is returned when an address is not a 20-byte hex string.
var ErrInvalidAddress = errors.New("invalid address")

// ChainReader is the part of *ethclient.Client this tool uses. Keeping it small
// lets tests pass in a fake instead of a real node.
type ChainReader interface {
	ChainID(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
}

// BlockSummary is what the tool reports about a block.
type BlockSummary struct {
	ChainID *big.Int
	Number  *big.Int
	Hash    common.Hash
	Time    time.Time
	TxCount uint
	GasUsed uint64
	// BaseFee is nil on chains or blocks without EIP-1559.
	BaseFee *big.Int
}

// LatestBlock summarizes the node's latest block using only its header.
func LatestBlock(ctx context.Context, r ChainReader) (BlockSummary, error) {
	chainID, err := r.ChainID(ctx)
	if err != nil {
		return BlockSummary{}, fmt.Errorf("get chain ID: %w", err)
	}

	header, err := r.HeaderByNumber(ctx, nil) // nil means "latest"
	if err != nil {
		return BlockSummary{}, fmt.Errorf("get latest header: %w", err)
	}

	// Count transactions in this exact block, by hash, so a new block arriving
	// between the two calls can't mix data from two blocks.
	hash := header.Hash()
	txCount, err := r.TransactionCount(ctx, hash)
	if err != nil {
		return BlockSummary{}, fmt.Errorf("count transactions: %w", err)
	}

	return BlockSummary{
		ChainID: chainID,
		Number:  header.Number,
		Hash:    hash,
		Time:    time.Unix(int64(header.Time), 0).UTC(),
		TxCount: txCount,
		GasUsed: header.GasUsed,
		BaseFee: header.BaseFee,
	}, nil
}

// Balance returns an account's latest balance in wei.
func Balance(ctx context.Context, r ChainReader, address string) (*big.Int, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidAddress, address)
	}

	balance, err := r.BalanceAt(ctx, common.HexToAddress(address), nil) // nil means "latest"
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}
	return balance, nil
}
