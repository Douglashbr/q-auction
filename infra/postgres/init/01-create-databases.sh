#!/bin/sh
# Database-per-service: uma instância, um database e um usuário por serviço.
# Roda apenas na primeira inicialização (volume pg-data vazio).
set -e

create_db() {
  db="$1"; user="$2"; password="$3"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER $user WITH PASSWORD '$password';
    CREATE DATABASE $db OWNER $user;
    REVOKE ALL ON DATABASE $db FROM PUBLIC;
EOSQL
}

create_db auction_db auction_user "$AUCTION_DB_PASSWORD"
create_db bidding_db bidding_user "$BIDDING_DB_PASSWORD"
create_db user_db    user_user    "$USER_DB_PASSWORD"
