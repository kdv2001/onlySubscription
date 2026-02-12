create table if not exists orders
(
    id           uuid primary key,
    user_id      uuid                        NOT NULL,
    order_status varchar                     NOT NULL,
    created_at   timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at   timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    total_price  decimal                     not null,
    currency     varchar                     not null,
    product_id   uuid                        not null,
    ttl          timestamp WITHOUT TIME ZONE NOT NULL not null
);

create table if not exists invoices
(
    id             uuid primary key            not null,
    state          varchar                     NOT NULL,
    payment_method varchar                     NOT NULL,
    created_at     timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at     timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    order_id       uuid                        NOT NULL,
    amount         decimal                     NOT NULL,
    currency       varchar                     NOT NULL,
    provider_id    varchar
);

create table if not exists products
(
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
);

create table if not exists products_sources
(
    id            uuid primary key,
    product_id    uuid references products (id),
    type          text                        NOT NULL,
    source        text                        NOT NULL,
    created_at    timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    record_status text                        NOT NULL default ('')
);

create table if not exists inventory
(
    id          uuid primary key,
    product_id  uuid references products (id),
    status      text                        NOT NULL default ('sale'),
    created_at  timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at  timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    description text                        NOT NULL
);

create table if not exists products
(
    id          uuid primary key,
    user_id     uuid                        NOT NULL,
    order_id    uuid                        NOT NULL unique,
    description text                        NOT NULL,
    created_at  timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at  timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    state       text                        NOT NULL,
    deadline    timestamp WITHOUT TIME ZONE NOT NULL
);

create table if not exists users
(
    id    uuid primary key,
    state varchar NOT NULL
);

create table if not exists auth_bot_telegram
(
    id          uuid primary key,
    user_id     uuid                        NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id),
    telegram_id varchar                     NOT NULL UNIQUE,
    created_at  timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    state       varchar                     NOT NULL,
    chat_id     varchar                     NOT NULL
);