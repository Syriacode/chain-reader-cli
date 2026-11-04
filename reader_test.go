package main

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// fakeReader is a hand-written ChainReader that returns fixed data, so tests
// never need a real node.
type fakeReader struct {
	headerErr      error
	balanceErr     error
	balanceAddress common.Address // set by BalanceAt, so tests can see if the node was called
}

// Compile-time check: fails the build if *fakeReader stops matching ChainReader.
var _ ChainReader = (*fakeReader)(nil)

func (f *fakeReader) ChainID(ctx context.Context) (*big.Int, error) {
	return big.NewInt(31337), nil
}

func (f *fakeReader) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if f.headerErr != nil {
		return nil, f.headerErr
	}
	return &types.Header{Number: big.NewInt(7), Time: 1700000000, GasUsed: 21000}, nil
}

func (f *fakeReader) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	return 3, nil
}

func (f *fakeReader) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	f.balanceAddress = account
	if f.balanceErr != nil {
		return nil, f.balanceErr
	}
	return big.NewInt(5), nil
}

func TestLatestBlockSummarizesHeader(t *testing.T) {
	got, err := LatestBlock(context.Background(), &fakeReader{})
	if err != nil {
		t.Fatalf("LatestBlock() error = %v", err)
	}

	if got.Number.Cmp(big.NewInt(7)) != 0 {
		t.Errorf("Number = %s, want 7", got.Number)
	}
	if got.TxCount != 3 {
		t.Errorf("TxCount = %d, want 3", got.TxCount)
	}
	if got.GasUsed != 21000 {
		t.Errorf("GasUsed = %d, want 21000", got.GasUsed)
	}
	wantTime := time.Date(2023, 11, 14, 22, 13, 20, 0, time.UTC)
	if !got.Time.Equal(wantTime) {
		t.Errorf("Time = %s, want %s", got.Time.Format(time.RFC3339), wantTime.Format(time.RFC3339))
	}
	if got.BaseFee != nil {
		t.Errorf("BaseFee = %s, want nil", got.BaseFee)
	}
}

func TestLatestBlockWrapsNodeError(t *testing.T) {
	errConnRefused := errors.New("connection refused")

	_, err := LatestBlock(context.Background(), &fakeReader{headerErr: errConnRefused})
	if !errors.Is(err, errConnRefused) {
		t.Fatalf("LatestBlock() error = %v, want it to wrap %v", err, errConnRefused)
	}
}

func TestBalanceRejectsBadAddressWithoutCallingNode(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{name: "empty", address: ""},
		{name: "not hex", address: "hello"},
		{name: "too short", address: "0x1234"},
		{name: "bad hex chars", address: "0xZZZZ6BF26964aF9D7eEd9e03E53415D37aA96045"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeReader{}

			_, err := Balance(context.Background(), f, tc.address)
			if !errors.Is(err, ErrInvalidAddress) {
				t.Errorf("Balance(%q) error = %v, want ErrInvalidAddress", tc.address, err)
			}
			if f.balanceAddress != (common.Address{}) {
				t.Errorf("Balance(%q) called the node with %s", tc.address, f.balanceAddress)
			}
		})
	}
}
