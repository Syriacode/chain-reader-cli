package main

import (
	"math/big"
	"strings"
)

// ethDecimals is how many decimal places ETH has: 1 ETH = 10^18 wei.
const ethDecimals = 18

// weiPerETH is 10^18. Amounts are kept in integer wei, so no precision is lost.
var weiPerETH = new(big.Int).Exp(big.NewInt(10), big.NewInt(ethDecimals), nil)

// FormatETH turns a wei amount into a decimal ETH string, using integer math only.
func FormatETH(wei *big.Int) string {
	// Split into whole ETH and leftover wei: wei = quotient*10^18 + remainder.
	quotient, remainder := new(big.Int).QuoRem(wei, weiPerETH, new(big.Int))

	// No leftover wei means a whole number of ETH, so there is no decimal point.
	if remainder.Sign() == 0 {
		return quotient.String()
	}

	// The fraction must be exactly 18 digits: 1 wei is "000000000000000001", not "1".
	frac := remainder.String()
	frac = strings.Repeat("0", ethDecimals-len(frac)) + frac

	// Drop trailing zeros: "500000000000000000" becomes "5".
	frac = strings.TrimRight(frac, "0")

	return quotient.String() + "." + frac
}
