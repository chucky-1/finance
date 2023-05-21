CREATE SCHEMA finance;

CREATE TABLE finance.users
(
    username varchar(15) PRIMARY KEY,
    password varchar(60)
)