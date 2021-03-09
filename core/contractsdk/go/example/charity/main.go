package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
)

type charityDonation struct {
}

const (
	USER_DONATE    = "UserDonate/"
	ALL_DONATE     = "AllDonate/"
	ALL_COST       = "AllCost/"
	TOTAL_RECEIVED = "TotalDonates"
	TOTAL_COSTS    = "TotalCosts"
	BALANCE        = "Balance"
	DONATE_COUNT   = "DonateCount"
	COST_COUNT     = "CostCount"
	ADMIN          = "admin"
	MAX_LIMIT      = 100
)

type costDetail struct {
	Id        string `json:"id"`
	To        string `json:"to"`
	Amount    string `json:"amount"`
	Timestamp string `json:"timestamp"`
	Comments  string `json:"comments"`
}

type donateDetail struct {
	Id        string `json:"id"`
	Donor     string `json:"donor"`
	Amount    string `json:"amount"`
	Timestamp string `json:"timestamp"`
	Comments  string `json:"comments"`
}

var (
	ErrLimitExceeded = errors.New("limit exceeded")
)

func (cd *charityDonation) Initialize(ctx code.Context) code.Response {
	args := &struct {
		Admin string `json:"admin" validate:"required"`
	}{}
	if err := code.Unmarshal(ctx.Args(), args); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(ADMIN), []byte(args.Admin)); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(TOTAL_RECEIVED), []byte("0")); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(TOTAL_COSTS), []byte("0")); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(BALANCE), []byte("0")); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(DONATE_COUNT), []byte("0")); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(COST_COUNT), []byte("0")); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok"))
}

func (cd *charityDonation) Donate(ctx code.Context) code.Response {
	err := cd.checkPermission(ctx)
	if err != nil {
		return code.Error(err)
	}
	args := struct {
		Donor     string   `json:"donor" validate:"required,excludes=/"`
		Amount    *big.Int `json:"amount" validate:"required,gt=0"`
		Timestamp string   `json:"timestamp" validate:"required"`
		Comments  string   `json:"comments" validate:"required"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	donateCountByte, err := ctx.GetObject([]byte(DONATE_COUNT))
	if err != nil {
		return code.Error(err)
	}

	donateCount, _ := big.NewInt(0).SetString(string(donateCountByte), 10)
	donateCount = donateCount.Add(donateCount, big.NewInt(1))
	donateID := fmt.Sprintf("%020s", donateCount.String())

	totalReceivedByte, err := ctx.GetObject([]byte(TOTAL_RECEIVED))
	if err != nil {
		return code.Error(err)
	}

	totalReceived, _ := big.NewInt(0).SetString(string(totalReceivedByte), 10)
	totalReceived = totalReceived.Add(totalReceived, args.Amount)

	balanceByte, err := ctx.GetObject([]byte(BALANCE))
	if err != nil {
		return code.Error(err)
	}
	balance, _ := big.NewInt(0).SetString(string(balanceByte), 10)
	balance = balance.Add(balance, args.Amount)
	donateDetailByte, _ := json.Marshal(donateDetail{
		Donor:     args.Donor,
		Amount:    args.Amount.String(),
		Timestamp: args.Timestamp,
		Comments:  args.Comments,
	})

	userDonateKey := USER_DONATE + args.Donor + "/" + donateID
	if err := ctx.PutObject([]byte(userDonateKey), donateDetailByte); err != nil {
		return code.Error(err)
	}

	allDonateKey := ALL_DONATE + donateID
	if err := ctx.PutObject([]byte(allDonateKey), donateDetailByte); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(DONATE_COUNT), []byte(donateCount.String())); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(TOTAL_RECEIVED), []byte(totalReceived.String())); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(BALANCE), []byte(balance.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(donateID))
}

func (cd *charityDonation) Cost(ctx code.Context) code.Response {
	if err := cd.checkPermission(ctx); err != nil {
		return code.Error(err)
	}
	args := struct {
		To        string   `json:"to" validate:"required"`
		Amount    *big.Int `json:"amount" validate:"required,gt=0"`
		Timestamp string   `json:"timestamp" validate:"required"`
		Comments  string   `json:"comments" validate:"required"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	costCountByte, err := ctx.GetObject([]byte(COST_COUNT))
	if err != nil {
		return code.Error(err)
	}
	costCount, _ := big.NewInt(0).SetString(string(costCountByte), 10)
	costCount = costCount.Add(costCount, big.NewInt(1))

	totalCostsByte, err := ctx.GetObject([]byte(TOTAL_COSTS))
	if err != nil {
		return code.Error(err)
	}
	totalCost, _ := big.NewInt(0).SetString(string(totalCostsByte), 10)
	totalCost = totalCost.Add(totalCost, args.Amount)

	balanceByte, err := ctx.GetObject([]byte(BALANCE))
	if err != nil {
		return code.Error(err)
	}
	balance, _ := big.NewInt(0).SetString(string(balanceByte), 10)
	if balance.Cmp(args.Amount) < 0 {
		return code.Error(code.ErrBalanceLow)
	}
	balance = balance.Sub(balance, args.Amount)
	data := costDetail{
		To:        args.To,
		Amount:    args.Amount.String(),
		Timestamp: args.Timestamp,
		Comments:  args.Comments,
	}
	dataByte, _ := json.Marshal(data)
	costId := fmt.Sprintf("%020d", costCount.Int64())
	allCostKey := ALL_COST + costId
	if err := ctx.PutObject([]byte(allCostKey), dataByte); err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(COST_COUNT), []byte(costCount.String())); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(TOTAL_COSTS), []byte(totalCost.String())); err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(BALANCE), []byte(balance.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(costId))
}

func (cd *charityDonation) Statistics(ctx code.Context) code.Response {
	totalReceived, err := ctx.GetObject([]byte(TOTAL_RECEIVED))
	if err != nil {
		return code.Error(err)
	}

	totalCost, err := ctx.GetObject([]byte(TOTAL_COSTS))
	if err != nil {
		return code.Error(err)
	}
	balance, err := ctx.GetObject([]byte(BALANCE))
	if err != nil {
		return code.Error(err)
	}

	return code.JSON(struct {
		TotalDonate string `json:"total_donate"`
		TotalCost   string `json:"total_cost"`
		FundBalance string `json:"fund_balance"`
	}{
		string(totalReceived),
		string(totalCost),
		string(balance),
	})
}

func (cd *charityDonation) QueryDonor(ctx code.Context) code.Response {
	args := struct {
		Donar string `json:"donor" validate:"required,excludes=/"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	prefix := USER_DONATE + args.Donar + "/"
	iter := ctx.NewIterator(code.PrefixRange([]byte(prefix)))
	donateCount := big.NewInt(0)
	defer iter.Close()

	donateDetails := []donateDetail{}

	for iter.Next() {
		donateCount = donateCount.Add(donateCount, big.NewInt(1))
		donateId := iter.Key()[len(prefix):]
		detail := donateDetail{}
		if err := json.Unmarshal(iter.Value(), &detail); err != nil {
			return code.Error(err)
		}
		detail.Id = string(donateId)
		donateDetails = append(donateDetails, detail)
	}

	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.JSON(donateDetails)
}

func (cd *charityDonation) QueryDonates(ctx code.Context) code.Response {
	args := struct {
		Start *big.Int `json:"start" validate:"required"`
		Limit *big.Int `json:"limit" validate:"required"`
	}{}

	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	if args.Limit.Cmp(big.NewInt(MAX_LIMIT)) > 0 {
		return code.Error(ErrLimitExceeded)
	}
	end := big.NewInt(0).Add(args.Start, args.Limit)
	iter := ctx.NewIterator([]byte(ALL_DONATE+fmt.Sprintf("%020s", args.Start.String())), []byte(ALL_DONATE+fmt.Sprintf("%020s", end.String())))
	defer iter.Close()

	donateDetails := []donateDetail{}
	selected := int64(0) // use selected is safe as we check limit before
	for iter.Next() {
		if selected >= args.Limit.Int64() {
			break
		}
		selected++
		detail := donateDetail{}
		if err := json.Unmarshal(iter.Value(), &detail); err != nil {
			return code.Error(err)
		}
		donateId := iter.Key()[len([]byte(ALL_DONATE)):]
		detail.Id = string(donateId)
		donateDetails = append(donateDetails, detail)
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.JSON(donateDetails)
}

func (cd *charityDonation) QueryCosts(ctx code.Context) code.Response {
	args := struct {
		Start *big.Int `json:"start" validate:"required"`
		Limit *big.Int `json:"limit" validate:"required"`
	}{}

	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if args.Limit.Cmp(big.NewInt(MAX_LIMIT)) > 0 {
		return code.Error(ErrLimitExceeded)
	}

	end := big.NewInt(0).Add(args.Start, args.Limit)

	iter := ctx.NewIterator([]byte(ALL_COST+fmt.Sprintf("%020s", args.Start.String())), []byte(ALL_COST+fmt.Sprintf("%020s", end.String())))

	defer iter.Close()

	selected := int64(0)
	details := []costDetail{}
	for iter.Next() {
		if selected >= args.Limit.Int64() {
			break
		}
		selected++
		costId := iter.Key()[len([]byte(ALL_COST)):]
		detail := costDetail{}
		json.Unmarshal(iter.Value(), &detail)
		detail.Id = string(costId)
		details = append(details, detail)
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.JSON(details)
}

func (cd *charityDonation) checkPermission(ctx code.Context) error {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.ErrMissingInitiator
	}
	admin, err := ctx.GetObject([]byte(ADMIN))
	if err != nil {
		return err
	}
	if initiator != string(admin) {
		return code.ErrPermissionDenied
	}
	return nil
}

func main() {
	driver.Serve(new(charityDonation))
}
