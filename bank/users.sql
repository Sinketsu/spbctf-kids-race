CREATE DATABASE IF NOT EXISTS bank;
create table IF NOT EXISTS bank.users
(
    id     int auto_increment primary key,
    login   varchar(50) not null unique,
    password text,
    session text,
    money int default 100,
    shared int default 0
);

CREATE USER IF NOT EXISTS 'bank'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON bank.* TO 'bank'@'%';

