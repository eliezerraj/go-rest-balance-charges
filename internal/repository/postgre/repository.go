package db_postgre

import (
	"context"
	"time"
	"errors"

	_ "github.com/lib/pq"

	"github.com/go-rest-balance-charges/internal/core"
	"github.com/go-rest-balance-charges/internal/erro"

)

type WorkerRepository struct {
	databaseHelper DatabaseHelper
}

func NewWorkerRepository(databaseHelper DatabaseHelper) WorkerRepository {
	childLogger.Debug().Msg("NewWorkerRepository")
	return WorkerRepository{
		databaseHelper: databaseHelper,
	}
}

func (w WorkerRepository) Ping() (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("Ping")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	ctx, cancel := context.WithTimeout(context.Background(), 1000)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)
	err := client.Ping()
	if err != nil {
		return false, erro.ErrConnectionDatabase
	}

	return true, nil
}

func (w WorkerRepository) Add(balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("Add")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	stmt, err := client.Prepare(`INSERT INTO balance_charge ( 	fk_balance_id, 
																type_charge,
																charged_at, 
																currency,
																amount,
																tenant_id) 
									VALUES($1, $2, $3, $4, $5, $6) `)
	if err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}
	_, err = stmt.Exec(	balanceCharge.FkBalanceID, 
						balanceCharge.Type,
						time.Now(),
						balanceCharge.Currency,
						balanceCharge.Amount,
						balanceCharge.TenantID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return nil, errors.New(err.Error())
	}

	return &balanceCharge , nil
}

func (w WorkerRepository) Get(balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("Get")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	result_query := core.BalanceCharge{}
	rows, err := client.Query(`SELECT id, fk_balance_id, type_charge, charged_at, currency, amount, tenant_id
								FROM balance_charge 
								WHERE id =$1`, balanceCharge.ID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&result_query.ID, 
							&result_query.FkBalanceID, 
							&result_query.Type, 
							&result_query.ChargeAt,
							&result_query.Currency,
							&result_query.Amount,
							&result_query.TenantID,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		return &result_query, nil
	}

	return nil, erro.ErrNotFound
}

func (w WorkerRepository) List(balanceCharge core.BalanceCharge) (*[]core.BalanceCharge, error){
	childLogger.Debug().Msg("List")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	result_query := core.BalanceCharge{}
	balance_list := []core.BalanceCharge{}

	rows, err := client.Query(`SELECT id, fk_balance_id, type_charge, charged_at, currency, amount, tenant_id
								FROM balance_charge 
								WHERE fk_balance_id =$1 order by charged_at desc`, balanceCharge.FkBalanceID)
	if err != nil {
		childLogger.Error().Err(err).Msg("SELECT statement")
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		err := rows.Scan( 	&result_query.ID, 
							&result_query.FkBalanceID, 
							&result_query.Type, 
							&result_query.ChargeAt,
							&result_query.Currency,
							&result_query.Amount,
							&result_query.TenantID,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		balance_list = append(balance_list, result_query)
	}

	return &balance_list , nil
}
