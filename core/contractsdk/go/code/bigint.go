package code

import "math/big"

var bigIntFunctions = map[string]func(*big.Int, string) bool{
	"lte": ilte,
	"lt":  ilt,
	"eq":  ieq,
	"neq": ineq,
	"gt":  igt,
	"gte": igte,
}

func ieq(a *big.Int, b string) bool {
	c, ok := big.NewInt(0).SetString(b, 10)
	if !ok {
		return false
	}
	return a.Cmp(c) == 0
}

func ineq(a *big.Int, b string) bool {
	c, ok := big.NewInt(0).SetString(b, 10)
	if !ok {
		return false
	}
	return a.Cmp(c) != 0
}

func igt(a *big.Int, b string) bool {
	c, ok := big.NewInt(0).SetString(b, 10)
	if !ok {
		return false
	}
	return a.Cmp(c) > 0
}
func igte(a *big.Int, b string) bool {

	c, ok := big.NewInt(0).SetString(b, 10)
	if !ok {
		return false
	}
	return a.Cmp(c) >= 0
}

func ilt(a *big.Int, b string) bool {

	c, ok := big.NewInt(0).SetString(b, 10)
	if !ok {
		return false
	}
	return a.Cmp(c) < 0
}

func ilte(a *big.Int, b string) bool {
	c, ok := big.NewInt(0).SetString(b, 10)
	if !ok {
		return false
	}
	return a.Cmp(c) <= 0
}
