package db_postgre

import (
	"context"
	"time"
	"fmt"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/go-rest-balance-charges/internal/core"

)

var childLogger = log.With().Str("repository/postgre", "NewDatabaseHelper").Logger()

type DatabaseHelper interface {
	GetConnection(ctx context.Context) (*sql.DB, error)
	CloseConnection()
}

type DatabaseHelperImplementacion struct {
	client   	*sql.DB
	timeout		time.Duration
}

func NewDatabaseHelper(databaseRDS core.DatabaseRDS) (DatabaseHelper, error) {
	childLogger.Debug().Msg("NewDatabaseHelper")

	_ , cancel := context.WithTimeout(context.Background(), time.Duration(databaseRDS.Db_timeout)*time.Second)
	defer cancel()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
							databaseRDS.User, 
							databaseRDS.Password, 
							databaseRDS.Host, 
							databaseRDS.Port, 
							databaseRDS.DatabaseName) 
	
	client, err := sql.Open(databaseRDS.Postgres_Driver, connStr)
	if err != nil {
		return DatabaseHelperImplementacion{}, err
	}
	err = client.Ping()
	if err != nil {
		return DatabaseHelperImplementacion{}, err
	}

	return DatabaseHelperImplementacion{
		client: client,
		timeout:  time.Duration(databaseRDS.Db_timeout) * time.Second,
	}, nil
}

func (d DatabaseHelperImplementacion) GetConnection(ctx context.Context) (*sql.DB, error) {
	childLogger.Debug().Msg("GetConnection")
	return d.client, nil
}

func (d DatabaseHelperImplementacion) CloseConnection()  {
	childLogger.Debug().Msg("CloseConnection")
	defer d.client.Close()
}
