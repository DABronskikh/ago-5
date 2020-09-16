-- табличка с пользователями и их паролями
CREATE TABLE users
(
    id       BIGSERIAL PRIMARY KEY,
    login    TEXT      NOT NULL UNIQUE,
    password TEXT      NOT NULL,
    roles    TEXT[]    NOT NULL DEFAULT '{}',
    created  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- табличка с токенами (чтобы пользователь каждый раз не присылал пароль и логин)
CREATE TABLE tokens (
    id       TEXT PRIMARY KEY,
    user_id   BIGINT NOT NULL REFERENCES users,
    created  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cards
(
    id       BIGSERIAL PRIMARY KEY,
    number   TEXT      NOT NULL,
    balance  BIGINT    NOT NULL DEFAULT 0,
    issuer   TEXT      NOT NULL CHECK (issuer IN ('Visa', 'MasterCard', 'MIR')),
    holder   TEXT      NOT NULL,
    user_id  BIGINT    NOT NULL REFERENCES users,
    status   TEXT      NOT NULL DEFAULT 'INACTIVE' CHECK (status IN ('INACTIVE', 'ACTIVE')),
    created  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
