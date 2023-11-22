package service

import (
	"context"
	"strconv"

	"github.com/go-rest-balance-charges/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"

)

func (s WorkerService) GetCache(ctx context.Context, balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("GetCache")

	_, root := xray.BeginSubsegment(ctx, "Service.GetCache")
	defer root.Close(nil)

	res, err := s.cache.Get(ctx, balanceCharge.AccountID)
	if err != nil {
		return nil, err
	}
	
	item, _ := strconv.ParseFloat(res.(string), 64)
	balanceCharge.Amount = item

	return &balanceCharge, nil
}
