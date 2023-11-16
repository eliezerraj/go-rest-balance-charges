package service

import (
	"errors"
	"context"
	"github.com/rs/zerolog/log"
	"github.com/mitchellh/mapstructure"

	"github.com/go-rest-balance-charges/internal/erro"
	"github.com/go-rest-balance-charges/internal/core"
	"github.com/go-rest-balance-charges/internal/repository/postgre"
	"github.com/go-rest-balance-charges/internal/adapter/restapi"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/sony/gobreaker"

)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 		*db_postgre.WorkerRepository
	restapi					*restapi.RestApiSConfig
	circuitBreaker			*gobreaker.CircuitBreaker
}

func NewWorkerService(workerRepository *db_postgre.WorkerRepository, 
						restapi *restapi.RestApiSConfig,
						circuitBreaker	*gobreaker.CircuitBreaker) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
		restapi:			restapi,
		circuitBreaker: 	circuitBreaker,
	}
}

func (s WorkerService) Add(ctx context.Context, balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("Add")

	_, root := xray.BeginSubsegment(ctx, "Service.Add")
	defer root.Close(nil)

	rest_interface_data, err := s.restapi.GetData(ctx, balanceCharge.AccountID)
	if err != nil {
		return nil, err
	}

	var balance_parsed core.Balance
	err = mapstructure.Decode(rest_interface_data, &balance_parsed)
    if err != nil {
		childLogger.Error().Err(err).Msg("error parse interface")
		return nil, errors.New(err.Error())
    }

	childLogger.Debug().Interface("balance_parsed:",balance_parsed).Msg("")
	
	balanceCharge.FkBalanceID = balance_parsed.ID
	res, err := s.workerRepository.Add(ctx, balanceCharge)
	if err != nil {
		return nil, err
	}

	balance_parsed.Amount = balance_parsed.Amount + balanceCharge.Amount
	childLogger.Debug().Interface("balance_parsed:",balance_parsed).Msg("")

	_, err = s.restapi.PostData(ctx, balanceCharge.AccountID, balance_parsed)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Get(ctx context.Context, balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("Get")

	_, root := xray.BeginSubsegment(ctx, "Service.Get")
	defer root.Close(nil)

	res, err := s.workerRepository.Get(ctx,balanceCharge)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) GetCb(ctx context.Context, balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("GetCb")

	// tracer
	_, root := xray.BeginSubsegment(ctx, "Service.GetCb")
	defer root.Close(nil)

	// Business rule with CB
	res_cb, err := s.circuitBreaker.Execute(func() (interface{}, error) {
		
		res, err := s.workerRepository.Get(ctx,balanceCharge)
		if err != nil {
			return nil, err
		}
	
		return res, nil
	})

	if (err != nil) {
		if (err != erro.ErrNotFound) {
			childLogger.Debug().Msg("Circuit Breaker OPEN !!!")
			return nil, erro.ErrPending
		} else {
			return nil, err
		}
	}
	
	// Assertion for Cb
	var balance_charge_assertion core.BalanceCharge
	err = mapstructure.Decode(res_cb, &balance_charge_assertion)
    if err != nil {
		childLogger.Error().Err(err).Msg("error parse interface")
		return nil, errors.New(err.Error())
    }

	return &balance_charge_assertion, nil
}

func (s WorkerService) List(ctx context.Context, balanceCharge core.BalanceCharge) (*[]core.BalanceCharge, error){
	childLogger.Debug().Msg("List")

	_, root := xray.BeginSubsegment(ctx, "Service.List")
	defer root.Close(nil)

	rest_interface_data, err := s.restapi.GetData(ctx, balanceCharge.AccountID)
	if err != nil {
		return nil, err
	}

	var balance_parsed core.Balance
	err = mapstructure.Decode(rest_interface_data, &balance_parsed)
    if err != nil {
		childLogger.Error().Err(err).Msg("error parse interface")
		return nil, errors.New(err.Error())
    }

	balanceCharge.FkBalanceID = balance_parsed.ID
	res, err := s.workerRepository.List(ctx,balanceCharge)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) AddCtx(ctx context.Context, balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("AddCtx")

	_, root := xray.BeginSubsegment(ctx, "Service.AddCtx")
	defer root.Close(nil)

	tx, err := s.workerRepository.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rest_interface_data, err := s.restapi.GetData(ctx, balanceCharge.AccountID)
	if err != nil {
		return nil, err
	}

	var balance_parsed core.Balance
	err = mapstructure.Decode(rest_interface_data, &balance_parsed)
    if err != nil {
		childLogger.Error().Err(err).Msg("error parse interface")
		return nil, errors.New(err.Error())
    }

	childLogger.Debug().Interface("balance_parsed:",balance_parsed).Msg("")

	balanceCharge.FkBalanceID = balance_parsed.ID
	res, err := s.workerRepository.AddCtx(ctx, tx, balanceCharge)
	if err != nil {
		return nil, err
	}

	balance_parsed.Amount = balance_parsed.Amount + balanceCharge.Amount
	childLogger.Debug().Interface("balance_parsed:",balance_parsed).Msg("")

	_, err = s.restapi.PostData(ctx, balanceCharge.AccountID, balance_parsed)
	if err != nil {
		return nil, err
	}

	tx.Commit()
	return res, nil
}