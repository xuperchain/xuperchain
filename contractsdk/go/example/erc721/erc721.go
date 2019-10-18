package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/driver"
)

// erc721 struct totolSupply: total digitAsset
// DigitAsset uniqueness
// database key: totalsupply
type erc721 struct {
	totalSupply int64
	balanceOf   map[string]*[]int64
	approvalOf  map[string]*[]int64
	ctx         code.Context
}

func newERC721() *erc721 {
	return &erc721{
		balanceOf:  map[string]*[]int64{},
		approvalOf: map[string]*[]int64{},
	}
}

func (e *erc721) makeBalanceOfKey(from string) string {
	return "balanceOf_" + from
}

// makeApprovalOfKey a_b a allow b to spend a's asset
func (e *erc721) makeApprovalOfKey(from string, to string) string {
	return "approvalOf_" + from + "_" + to
}

func (e *erc721) fillBalanceOf(addr string) {
	key := e.makeBalanceOfKey(addr)
	vals := e.getObject(key)
	e.balanceOf[key] = vals
	log.Printf("fillBalanceOf: key: %v, vals: %v", key, vals)
}

func (e *erc721) fillApprovalOf(from string, to string) {
	key := e.makeApprovalOfKey(from, to)
	vals := e.getObject(key)
	e.approvalOf[key] = vals
	log.Printf("fillApprovalOf: key: %v, vals: %v", key, vals)
}

func (e *erc721) getObject(key string) *[]int64 {
	value, err := e.ctx.GetObject([]byte(key))
	if err != nil {
		return &[]int64{}
	}

	log.Printf("getObject exist")
	vals := &[]int64{}
	json.Unmarshal(value, vals)
	return vals
}

func (e *erc721) commitBalanceOf(addr string) {
	key := e.makeBalanceOfKey(addr)
	valsJSON, _ := json.Marshal(e.balanceOf[key])
	e.ctx.PutObject([]byte(key), valsJSON)
	log.Printf("commitBalanceOf: key: %v, val: %v", key, string(valsJSON))
}

func (e *erc721) commitApprovalOf(from string, to string) {
	key := e.makeApprovalOfKey(from, to)
	valsJSON, _ := json.Marshal(e.approvalOf[key])
	e.ctx.PutObject([]byte(key), valsJSON)
	log.Printf("commitApprovalOf: key: %v, val: %v", key, string(valsJSON))
}

func (e *erc721) ownerOf(tokenID int64, from string) bool {
	key := e.makeBalanceOfKey(from)
	for _, tid := range *e.balanceOf[key] {
		if tokenID == tid {
			log.Printf("ownerOf: tokenID: %v in tids %v", tokenID, e.balanceOf[key])
			return true
		}
	}

	log.Printf("ownerOf: tokenID: %v not in tids %v", tokenID, e.balanceOf[key])
	return false
}

func (e *erc721) transfer(from string, to string, tokenID int64) error {
	if !e.ownerOf(tokenID, from) {
		log.Printf("transfer: from donot tokenID: %v", tokenID)
		return fmt.Errorf("transfer: tokenID: %v  not belong to from", tokenID)
	}

	e.sub(from, tokenID)
	e.add(to, tokenID)
	return nil
}

func (e *erc721) sub(from string, tokenID int64) {
	key := e.makeBalanceOfKey(from)
	for i, tid := range *e.balanceOf[key] {
		if tokenID == tid {
			*e.balanceOf[key] = append((*e.balanceOf[key])[:i], (*e.balanceOf[key])[i+1:]...)
			log.Printf("sub: from contains: %v, toke_id: %v", *e.balanceOf[key], tokenID)
			break
		}
	}
}

func (e *erc721) add(to string, tokenID int64) {
	key := e.makeBalanceOfKey(to)
	*e.balanceOf[key] = append(*e.balanceOf[key], tokenID)
	log.Printf("add: to contains: %v, toke_id: %v", *e.balanceOf[key], tokenID)
}

func (e *erc721) isApproved(from string, caller string, tokenID int64) bool {
	akey := e.makeApprovalOfKey(from, caller)
	for _, old := range *e.approvalOf[akey] {
		if tokenID == old {
			return true
		}
	}
	return false
}

func (e *erc721) transferFrom(from string, caller string, to string, tokenID int64) error {
	if !e.isApproved(from, caller, tokenID) {
		return fmt.Errorf("from is not authorized to caller")
	}

	err := e.transfer(from, to, tokenID)
	if err != nil {
		return err
	}

	akey := e.makeApprovalOfKey(from, caller)
	for i, tid := range *e.approvalOf[akey] {
		if tokenID == tid {
			*e.approvalOf[akey] = append((*e.approvalOf[akey])[:i], (*e.approvalOf[akey])[i+1:]...)
		}
	}

	return nil
}

func (e *erc721) approve(from string, to string, tokenID int64) error {
	if e.isApproved(from, to, tokenID) {
		return nil
	}

	if !e.ownerOf(tokenID, from) {
		return fmt.Errorf("approve: tokenID: %v  not belong to from: %v", tokenID, from)
	}

	akey := e.makeApprovalOfKey(from, to)
	*e.approvalOf[akey] = append(*e.approvalOf[akey], tokenID)

	return nil
}

func (e *erc721) approveAll(from string, to string) error {
	akey := e.makeApprovalOfKey(from, to)
	key := e.makeBalanceOfKey(from)
	if len(*e.balanceOf[key]) == 0 {
		return fmt.Errorf("from empty")
	}

	e.approvalOf[akey] = &[]int64{}
	for _, tid := range *e.balanceOf[key] {
		*e.approvalOf[akey] = append(*e.approvalOf[akey], tid)
	}

	return nil
}

func (e *erc721) setContext(ctx code.Context) {
	e.ctx = ctx
}

func (e *erc721) Initialize(ctx code.Context) code.Response {
	e.setContext(ctx)
	supplystr := string(ctx.Args()["supply"])
	if supplystr == "" {
		return code.Errors("Missing key: supply")
	}
	from := string(ctx.Args()["from"])
	if from == "" {
		return code.Errors("Missing key: supply")
	}

	vals := e.getObject("totalsupply")
	supply := []int64{}
	for _, s := range strings.Split(supplystr, ",") {
		num, _ := strconv.ParseInt(s, 10, 64)
		for _, o := range *vals {
			if num == o {
				break
			}
		}
		*vals = append(*vals, num)
		supply = append(supply, num)
	}

	supplyJSON, _ := json.Marshal(vals)
	ctx.PutObject([]byte("totalsupply"), supplyJSON)
	log.Printf("Initialize: totalSupply: %v", string(supplyJSON))
	log.Printf("Initialize: from: %v, vals: %v", from, supply)

	e.fillBalanceOf(from)
	key := e.makeBalanceOfKey(from)
	for _, s := range supply {
		*e.balanceOf[key] = append(*e.balanceOf[key], s)
	}
	e.commitBalanceOf(from)

	return code.OK(nil)
}

func (e *erc721) Invoke(ctx code.Context) code.Response {
	e.setContext(ctx)
	action := string(ctx.Args()["action"])
	if action == "" {
		return code.Errors("Missing key: action")
	}

	switch action {
	case "transfer":
		return e.Transfer(ctx)
	case "transferFrom":
		return e.TransferFrom(ctx)
	case "approve":
		return e.Approve(ctx)
	case "approveAll":
		return e.ApproveAll(ctx)
	default:
		return code.Errors("Invalid action " + action)
	}
}

func (e *erc721) Transfer(ctx code.Context) code.Response {
	from := string(ctx.Args()["from"])
	if from == "" {
		return code.Errors("Missing key: from")
	}
	to := string(ctx.Args()["to"])
	if to == "" {
		return code.Errors("Missing key: to")
	}
	tokenIDStr := string(ctx.Args()["tokenID"])
	if tokenIDStr == "" {
		return code.Errors("Missing key: from")
	}
	tokenID, _ := strconv.ParseInt(tokenIDStr, 10, 64)

	e.fillBalanceOf(from)
	e.fillBalanceOf(to)

	err := e.transfer(from, to, tokenID)
	if err != nil {
		log.Printf("Transfer tokenID:%v is not belong to from: %v", tokenID, from)
		return code.Errors("Token_id is not belong to from")
	}

	e.commitBalanceOf(from)
	e.commitBalanceOf(to)

	return code.OK(nil)
}

func (e *erc721) TransferFrom(ctx code.Context) code.Response {
	from := string(ctx.Args()["from"])
	if from == "" {
		return code.Errors("Missing key: from")
	}
	caller := string(ctx.Args()["caller"])
	if caller == "" {
		return code.Errors("Missing key: caller")
	}
	to := string(ctx.Args()["to"])
	if to == "" {
		return code.Errors("Missing key: to")
	}
	tokenIDStr := string(ctx.Args()["tokenID"])
	if tokenIDStr == "" {
		return code.Errors("Missing key: from")
	}
	tokenID, _ := strconv.ParseInt(tokenIDStr, 10, 64)

	e.fillBalanceOf(from)
	e.fillBalanceOf(to)
	e.fillApprovalOf(from, caller)

	err := e.transferFrom(from, caller, to, tokenID)
	if err != nil {
		log.Printf("TransferFrom: toke_id: %v is not authorized to caller: %v", tokenID, caller)
		return code.Errors("Token_id is not authorized to caller")
	}

	e.commitBalanceOf(from)
	e.commitBalanceOf(to)
	e.commitApprovalOf(from, caller)

	return code.OK(nil)
}

func (e *erc721) Approve(ctx code.Context) code.Response {
	from := string(ctx.Args()["from"])
	if from == "" {
		return code.Errors("Missing key: from")
	}
	to := string(ctx.Args()["to"])
	if to == "" {
		return code.Errors("Missing key: to")
	}
	tokenIDStr := string(ctx.Args()["tokenID"])
	if tokenIDStr == "" {
		return code.Errors("Missing key: from")
	}
	tokenID, _ := strconv.ParseInt(tokenIDStr, 10, 64)

	e.fillBalanceOf(from)
	e.fillApprovalOf(from, to)

	err := e.approve(from, to, tokenID)
	if err != nil {
		log.Printf("Approve: tokenID:%v is not belong to from: %v", tokenID, from)
		return code.Errors("Token_id is not belong to from")
	}

	e.commitApprovalOf(from, to)

	return code.OK(nil)
}

func (e *erc721) ApproveAll(ctx code.Context) code.Response {
	from := string(ctx.Args()["from"])
	if from == "" {
		return code.Errors("Missing key: from")
	}
	to := string(ctx.Args()["to"])
	if to == "" {
		return code.Errors("Missing key: to")
	}
	e.fillBalanceOf(from)

	err := e.approveAll(from, to)
	if err != nil {
		log.Printf("ApproveAll: from: %v empty", from)
		return code.Errors("from empty")
	}

	e.commitApprovalOf(from, to)

	return code.OK(nil)
}

func (e *erc721) Query(ctx code.Context) code.Response {
	action := string(ctx.Args()["action"])
	if action == "" {
		return code.Errors("Missing key: action")
	}

	switch action {
	case "totalSupply":
		return e.total(ctx)
	case "balanceOf":
		return e.balance(ctx)
	case "approvalOf":
		return e.approval(ctx)
	default:
		return code.Errors("Invalid action " + action)
	}
}

func (e *erc721) total(ctx code.Context) code.Response {
	value, err := ctx.GetObject([]byte("totalsupply"))
	if err != nil {
		log.Println("You need to do initialize method first")
		return code.Errors("You need to do initialize method first")
	}

	bvals := &[]int64{}
	err = json.Unmarshal(value, bvals)
	if err != nil {
		log.Println("Json unmarshal from database error")
	}
	log.Printf("totalSupply: vals: %v", bvals)

	resVal := strconv.Itoa(len(*bvals))

	return code.OK([]byte(resVal))
}

func (e *erc721) balance(ctx code.Context) code.Response {
	from := string(ctx.Args()["from"])
	if from == "" {
		return code.Errors("Missing key: from")
	}

	bkey := e.makeBalanceOfKey(from)
	value, err := ctx.GetObject([]byte(bkey))
	if err != nil {
		log.Printf("balance: get key:[%v] no exist", bkey)
		return code.Errors("from donot have digit asset")
	}

	bvals := &[]int64{}
	err = json.Unmarshal(value, bvals)
	if err != nil {
		log.Println("Json unmarshal from database error")
	}
	log.Printf("balance: from: %v, vals: %v", from, bvals)

	resVal := strconv.Itoa(len(*bvals))
	return code.OK([]byte(resVal))
}

func (e *erc721) approval(ctx code.Context) code.Response {
	caller := string(ctx.Args()["to"])
	if caller == "" {
		return code.Errors("Missing key: caller")
	}
	from := string(ctx.Args()["from"])
	if from == "" {
		return code.Errors("Missing key: from")
	}

	akey := e.makeApprovalOfKey(from, caller)
	value, err := ctx.GetObject([]byte(akey))
	if err != nil {
		log.Printf("approval: get key:[%v] no exist", akey)
		return code.Errors("[from] is not authorized to [to]")
	}

	avals := &[]int64{}
	err = json.Unmarshal(value, avals)
	if err != nil {
		log.Println("Json unmarshal from database error")
	}
	log.Printf("approvalOf: key: %v_%v, vals: %v", from, caller, avals)

	resVal := strconv.Itoa(len(*avals))
	return code.OK([]byte(resVal))
}

func main() {
	driver.Serve(newERC721())
}
