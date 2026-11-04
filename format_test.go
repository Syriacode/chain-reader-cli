package main

import (
	"math/big"
	"testing"
)

func TestFormatETH(t *testing.T) {
	tests := []struct {
		name string
		wei  string
		want string
	}{
		{name: "zero", wei: "0", want: "0"},
		{name: "one wei", wei: "1", want: "0.000000000000000001"},
		{name: "half ETH", wei: "500000000000000000", want: "0.5"},
		{name: "one ETH", wei: "1000000000000000000", want: "1"},
		{name: "one ETH and one wei", wei: "1000000000000000001", want: "1.000000000000000001"},
		{name: "anvil account", wei: "10000000000000000000000", want: "10000"},
		{name: "beyond float64", wei: "123456789012345678901234567890", want: "123456789012.34567890123456789"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wei, ok := new(big.Int).SetString(tc.wei, 10)
			if !ok {
				t.Fatalf("bad wei value %q in test table", tc.wei)
			}
			if got := FormatETH(wei); got != tc.want {
				t.Errorf("FormatETH(%s) = %q, want %q", tc.wei, got, tc.want)
			}
		})
	}
}
