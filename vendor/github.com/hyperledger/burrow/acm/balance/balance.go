package balance

import (
	"fmt"
	"math/big"
	"sort"
)

var (
	eth = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)

type Balances []Balance

func (b Balance) String() string {
	return fmt.Sprintf("{%v: %d}", b.Type, b.Amount)
}

func New() Balances {
	return []Balance{}
}

func (bs Balances) Sort() Balances {
	sort.Stable(bs)
	return bs
}

func (bs Balances) Len() int {
	return len(bs)
}

func (bs Balances) Less(i, j int) bool {
	if bs[i].Type < bs[j].Type {
		return true
	}
	return bs[i].Type == bs[j].Type && bs[i].Amount < bs[j].Amount
}

func (bs Balances) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func (bs Balances) Add(ty Type, amount uint64) Balances {
	return append(bs, Balance{
		Type:   ty,
		Amount: amount,
	})
}

func (bs Balances) Native(amount uint64) Balances {
	return bs.Add(TypeNative, amount)
}

func (bs Balances) Power(amount uint64) Balances {
	return bs.Add(TypePower, amount)
}

func (bs Balances) Sum(bss ...Balances) Balances {
	return Sum(append(bss, bs)...)
}

func Sum(bss ...Balances) Balances {
	sum := New()
	sumMap := make(map[Type]uint64)
	for _, bs := range bss {
		for _, b := range bs {
			sumMap[b.Type] += b.Amount
		}
	}
	for k, v := range sumMap {
		sum = sum.Add(k, v)
	}
	sort.Stable(sum)
	return sum
}

func Native(native uint64) Balance {
	return Balance{
		Type:   TypeNative,
		Amount: native,
	}
}

func Power(power uint64) Balance {
	return Balance{
		Type:   TypePower,
		Amount: power,
	}
}

func (bs Balances) Has(ty Type) bool {
	for _, b := range bs {
		if b.Type == ty {
			return true
		}
	}
	return false
}

func (bs Balances) Get(ty Type) *uint64 {
	for _, b := range bs {
		if b.Type == ty {
			return &b.Amount
		}
	}
	return nil
}

func (bs Balances) GetFallback(ty Type, fallback uint64) uint64 {
	for _, b := range bs {
		if b.Type == ty {
			return b.Amount
		}
	}
	return fallback
}

func (bs Balances) GetNative(fallback uint64) uint64 {
	return bs.GetFallback(TypeNative, fallback)
}

func (bs Balances) GetPower(fallback uint64) uint64 {
	return bs.GetFallback(TypePower, fallback)
}

func (bs Balances) HasNative() bool {
	return bs.Has(TypeNative)
}

func (bs Balances) HasPower() bool {
	return bs.Has(TypePower)
}

func NativeToWei(n uint64) *big.Int {
	// 1 native unit to 1 ether (wei)
	x := new(big.Int).SetUint64(n)
	return new(big.Int).Mul(x, eth)
}

func WeiToNative(n []byte) *big.Int {
	x := new(big.Int).SetBytes(n)
	return new(big.Int).Div(x, eth)
}
