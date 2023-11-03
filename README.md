# go-rest-balance-charges

POC for encryption purposes.

CRUD a balance_charge data

## Database

    CREATE TABLE balance_charge (
        id              SERIAL PRIMARY KEY,
        fk_balance_id   integer REFERENCES balance(id),
        type_charge     varchar(200) NULL,
        charged_at      timestamptz NULL,
        currency        varchar(10) NULL,   
        amount          float8 NULL,
        tenant_id       varchar(200) NULL
    );

## Endpoints

+ POST /add

        {
            "account_id": "ACC-001",
            "type_charge": "DEBITO",
            "currency": "BRL",
            "amount": -120.00,
            "tenant_id": "TENANT-001"
        }

+ GET /get/1

+ GET /header

+ GET /list/26
