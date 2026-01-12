package postgres

var tables = []string{productsTable,
	inventoryTable,
	sourcesTable}

var productsTable = `create table if not exists products (
                                          id                  uuid primary key,
    name                text                        NOT NULL,
    description         text                        NOT NULL,
    created_at          timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at          timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    type                text                        NOT NULL,
-- record_status поле для состояния самой записи активна/удалена
    record_status       text                        NOT NULL default (''),
    price               decimal                     NOT NULL DEFAULT (0),

    subscription_period text                        NOT NULL
);`

var sourcesTable = `create table if not exists products_sources
(
    id            uuid primary key,
    product_id    uuid references products (id),
    type          text                        NOT NULL,
    source        text                        NOT NULL,
    created_at    timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    record_status text                        NOT NULL default ('')
);`

var inventoryTable = `create table if not exists inventory
(
    id          uuid primary key,
    product_id  uuid references products (id),
    status      text                        NOT NULL default ('sale'),
    created_at  timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at  timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    description text                        NOT NULL
);`
