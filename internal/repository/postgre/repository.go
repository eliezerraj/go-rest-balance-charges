package db_postgre

import (
	"context"
	"time"
	"errors"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/go-rest-balance-charges/internal/core"
	"github.com/go-rest-balance-charges/internal/erro"
	"github.com/aws/aws-xray-sdk-go/xray"

)

var childLogger = log.With().Str("repository/postgre", "WorkerRepository").Logger()

type WorkerRepository struct {
	databaseHelper DatabaseHelper
}

func NewWorkerRepository(databaseHelper DatabaseHelper) WorkerRepository {
	childLogger.Debug().Msg("NewWorkerRepository")
	return WorkerRepository{
		databaseHelper: databaseHelper,
	}
}

func (w WorkerRepository) StartTx(ctx context.Context) (*sql.Tx, error) {
	childLogger.Debug().Msg("StartTx")

	client := w.databaseHelper.GetConnection()

	tx, err := client.BeginTx(ctx, &sql.TxOptions{})
    if err != nil {
        return nil, errors.New(err.Error())
    }

	return tx, nil
}

func (w WorkerRepository) Ping() (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("Ping")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	client := w.databaseHelper.GetConnection()
	err := client.Ping()
	if err != nil {
		return false, erro.ErrConnectionDatabase
	}

	return true, nil
}

func (w WorkerRepository) Add(ctx context.Context, balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("Add")

	_, root := xray.BeginSubsegment(ctx, "SQL.ADD-Balance-Charges")
	defer root.Close(nil)

	client := w.databaseHelper.GetConnection()

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
	_, err = stmt.ExecContext( ctx,
								balanceCharge.FkBalanceID, 
								balanceCharge.Type,
								time.Now(),
								balanceCharge.Currency,
								balanceCharge.Amount,
								balanceCharge.TenantID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return nil, errors.New(err.Error())
	}
	defer stmt.Close()

	return &balanceCharge , nil
}

func (w WorkerRepository) Get(ctx context.Context, balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("Get")

	_, root := xray.BeginSubsegment(ctx, "SQL.GET-Balance-Charges")
	defer root.Close(nil)

	client := w.databaseHelper.GetConnection()

	result_query := core.BalanceCharge{}
	rows, err := client.QueryContext(ctx, `SELECT id, fk_balance_id, type_charge, charged_at, currency, amount, tenant_id FROM balance_charge WHERE id =$1`, balanceCharge.ID)
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

func (w WorkerRepository) List(ctx context.Context, balanceCharge core.BalanceCharge) (*[]core.BalanceCharge, error){
	childLogger.Debug().Msg("List")
	
	_, root := xray.BeginSubsegment(ctx, "SQL.List-Balance-Charges")
	defer root.Close(nil)

	client := w.databaseHelper.GetConnection()

	result_query := core.BalanceCharge{}
	balance_list := []core.BalanceCharge{}

	rows, err := client.QueryContext(ctx, `SELECT id, fk_balance_id, type_charge, charged_at, currency, amount, tenant_id FROM balance_charge WHERE fk_balance_id =$1 order by charged_at desc`, balanceCharge.FkBalanceID)
	if err != nil {
		childLogger.Error().Err(err).Msg("SELECT statement")
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
		balance_list = append(balance_list, result_query)
	}

	return &balance_list , nil
}

func (w WorkerRepository) AddCtx(ctx context.Context, tx *sql.Tx, balanceCharge core.BalanceCharge) (*core.BalanceCharge, error){
	childLogger.Debug().Msg("AddCtx")

	_, root := xray.BeginSubsegment(ctx, "AddCtx.List-Balance-Charges")
	defer root.Close(nil)

	stmt, err := tx.Prepare(`INSERT INTO balance_charge ( 	fk_balance_id, 
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
	defer stmt.Close()

	_, err = stmt.ExecContext(	ctx,
								balanceCharge.FkBalanceID, 
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
