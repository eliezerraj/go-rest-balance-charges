package service

import (
	"errors"
	"context"
	"time"
	"github.com/rs/zerolog/log"
	"github.com/mitchellh/mapstructure"

	"github.com/go-rest-balance-charges/internal/core"
	"github.com/go-rest-balance-charges/internal/repository/postgre"
	"github.com/go-rest-balance-charges/internal/adapter/restapi"

)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 		*db_postgre.WorkerRepository
	restapi					*restapi.RestApiSConfig
}

func NewWorkerService(workerRepository *db_postgre.WorkerRepository, restapi *restapi.RestApiSConfig) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
		restapi:			restapi,
	}
}

func (s WorkerService) Add(balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("Add")

	rest_interface_data, err := s.restapi.GetData(balanceCharge.AccountID)
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
	res, err := s.workerRepository.Add(balanceCharge)
	if err != nil {
		return nil, err
	}

	balance_parsed.Amount = balance_parsed.Amount + balanceCharge.Amount
	childLogger.Debug().Interface("balance_parsed:",balance_parsed).Msg("")

	_, err = s.restapi.PostData(balanceCharge.AccountID, balance_parsed)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Get(balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("Get")

	res, err := s.workerRepository.Get(balanceCharge)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) List(balanceCharge core.BalanceCharge) (*[]core.BalanceCharge, error){
	childLogger.Debug().Msg("List")

	rest_interface_data, err := s.restapi.GetData(balanceCharge.AccountID)
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
	res, err := s.workerRepository.List(balanceCharge)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) AddCtx(balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("AddCtx")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	tx, err := s.workerRepository.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rest_interface_data, err := s.restapi.GetData(balanceCharge.AccountID)
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
	res, err := s.workerRepository.AddCtx(tx, balanceCharge)
	if err != nil {
		return nil, err
	}

	balance_parsed.Amount = balance_parsed.Amount + balanceCharge.Amount
	childLogger.Debug().Interface("balance_parsed:",balance_parsed).Msg("")

	_, err = s.restapi.PostData(balanceCharge.AccountID, balance_parsed)
	if err != nil {
		return nil, err
	}

	tx.Commit()
	return res, nil
}