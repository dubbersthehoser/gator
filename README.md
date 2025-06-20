
# Gator

A boot.dev project. An blog aggregator.

I didn't care for this project, not going to go through to much trouble with docs.

May change the hole design: remove multiple users, use SQLite, overall simplifying the hole thing.

# Dependency 

- Goose: `go install github.com/pressly/goose/v3/cmd/goose@latest`

## Development

- Sqlc:  `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

# Setup Postres (Linux)

1. Install: `sudo apt update; sudo apt install postgresql postgresql-contrib`
1. Set Password: `sudo passwd postgres`
1. Start Service: `sudo service postgresql start`
1. Connect: `sudo -u postgres psql`
1. Create DB: `CREATE DATABASE gator;`
1. Add User: `ALTER USER postgres PASSWORD 'XXXXXX';`

# Migration

1. Change database connection string in `dbclean.sh` e.g. `$DBCONS="postgres://postgres:postgres@localhost:5432/gator"`
1. Run migration `bash dbclean.sh`
