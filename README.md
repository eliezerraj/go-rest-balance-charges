# go-rest-balance-charges

POC for test purposes.

CRUD a balance_charge data. The ADD endpoint must be integrated with the service balance

## Depencies

The go-rest-balance service must be running

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

        curl --header "Content-Type: application/json" \
        --request POST \
        --data '{"account_id": "ACC-001","type_charge": "CRED", "currency": "BRL", "amount": 150.00, "tenant_id": "TENANT-001"}' \
        http://svc02.domain.com/add

        {
            "account_id": "ACC-001",
            "type_charge": "DEBITO",
            "currency": "BRL",
            "amount": -120.00,
            "tenant_id": "TENANT-001"
        }

+ GET /get/1

        curl svc02.domain.com/get/1 | jq

+ GET /header

        curl svc02.domain.com/header | jq

+ GET /list/ACC-001

        curl svc02.domain.com/list/ACC-001 | jq

+ POST /withdraw

        {
        "account_id": "ACC-201",
        "type_charge": "CRED",
        "currency": "BRL",
        "amount": -10.00,
        "tenant_id": "TENANT-001"
        }

## K8

Add in hosts file /etc/hosts the lines below in order to use ingress local 

    127.0.0.1   svc02.domain.com



