package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"math/big"
	"strconv"
)

type HashedTimelock struct{}

// riddleHash --> LockContract{...}
type LockContract struct {
	sender        string
	receiver      string
	amount        string
	expiredHeight string
	riddleHash    string
}

func (c *HashedTimelock) Initialize(ctx code.Context) code.Response {
	return code.OK(nil)
}

//Open a riddle
func (c *HashedTimelock) Open(ctx code.Context) code.Response {
	receiver, ok := ctx.Args()["receiver"]
	if !ok {
		return code.Errors("missing receiver")
	}
	riddleHash, ok := ctx.Args()["riddle_hash"]
	if !ok {
		return code.Errors("missing riddle_hash")
	}
	oldValue, _ := ctx.GetObject(riddleHash)
	if oldValue != nil {
		return code.Errors("duplicated riddle hash:" + string(riddleHash))
	}
	expiredHeight, ok := ctx.Args()["expired_height"]
	if !ok {
		return code.Errors("missing expired_height")
	}
	nHeight, _ := strconv.Atoi(string(expiredHeight))
	if nHeight <= 0 {
		return code.Errors("expired_height must be greater than zero")
	}
	sender := ctx.Initiator()
	amount, _ := ctx.TransferAmount()
	if amount.Cmp(big.NewInt(0)) < 1 {
		return code.Errors("amount should be greater than zero")
	}
	value := fmt.Sprintf("%s %s %s %s", sender, string(receiver), amount.String(), expiredHeight)
	err := ctx.PutObject(riddleHash, []byte(value))
	if err != nil {
		return code.Errors("put riddle failed:" + err.Error())
	}
	return code.OK(nil)
}

func (c *HashedTimelock) lookupRiddle(ctx code.Context) (*LockContract, error) {
	riddle, ok := ctx.Args()["riddle"]
	if !ok {
		return nil, errors.New("missing riddle")
	}
	gotHash := sha256.Sum256(riddle)
	gotHashHex := fmt.Sprintf("%x", gotHash)
	value, err := ctx.GetObject([]byte(gotHashHex))
	if err != nil {
		return nil, errors.New("no riddle_hash in state db:" + err.Error())
	}
	HashedTimelockItem := &LockContract{}
	n, err := fmt.Sscanf(string(value), "%s %s %s %s", &HashedTimelockItem.sender, &HashedTimelockItem.receiver, &HashedTimelockItem.amount, &HashedTimelockItem.expiredHeight)
	if n != 4 || err != nil {
		return nil, errors.New("wrong format value:" + string(value))
	}
	HashedTimelockItem.riddleHash = gotHashHex
	return HashedTimelockItem, nil
}

// Withdraw assets from contract for receiver, if the correct riddle provided
func (c *HashedTimelock) Withdraw(ctx code.Context) code.Response {
	HashedTimelockItem, err := c.lookupRiddle(ctx)
	if err != nil {
		return code.Error(err)
	}
	amount, _ := big.NewInt(0).SetString(HashedTimelockItem.amount, 10)
	err = ctx.Transfer(HashedTimelockItem.receiver, amount)
	if err != nil {
		return code.Errors("trasfer failed:" + err.Error() + ", amount:" + HashedTimelockItem.amount)
	}
	err = ctx.DeleteObject([]byte(HashedTimelockItem.riddleHash))
	if err != nil {
		return code.Errors("delete riddle failed:" + err.Error())
	}
	return code.OK(nil)
}

// Refund: refund assets from contract, if the original sender can prove the height of ledger has already gone beyond the expired block-height
func (c *HashedTimelock) Refund(ctx code.Context) code.Response {
	HashedTimelockItem, err := c.lookupRiddle(ctx)
	if err != nil {
		return code.Error(err)
	}
	if HashedTimelockItem.sender != string(ctx.Initiator()) {
		return code.Errors("not original sender, expected:" + HashedTimelockItem.sender)
	}
	blockid, ok := ctx.Args()["blockid"]
	if !ok {
		return code.Errors("args missing blockid")
	}
	block, err := ctx.QueryBlock(string(blockid))
	if err != nil {
		return code.Errors("find block err:" + err.Error())
	}
	expiredHeight, _ := strconv.Atoi(HashedTimelockItem.expiredHeight)
	if block.Height >= int64(expiredHeight) {
		amount, _ := big.NewInt(0).SetString(HashedTimelockItem.amount, 10)
		err := ctx.Transfer(HashedTimelockItem.sender, amount)
		if err != nil {
			return code.Errors("trasfer failed:" + err.Error())
		}
		err = ctx.DeleteObject([]byte(HashedTimelockItem.riddleHash))
		if err != nil {
			return code.Errors("delete riddle failed:" + err.Error())
		}
	} else {
		return code.Errors("still frozen until expired, " + "proof height:" + fmt.Sprintf("%d", block.Height) + ",expected:" + HashedTimelockItem.expiredHeight)
	}
	return code.OK(nil)
}

func main() {
	driver.Serve(new(HashedTimelock))
}
