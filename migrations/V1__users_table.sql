CREATE SCHEMA finance;

CREATE TABLE finance.users
(
    username varchar(15) PRIMARY KEY,
    password varchar(60)  NOT NULL,
    country  varchar(168) NOT NULL,
    timezone interval     NOT NULL
)