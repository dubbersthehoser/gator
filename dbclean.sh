#!/bin/bash

set -e

DBCONS="postgres://postgres:postgres@localhost:5432/gator"

cd sql/schema/
goose postgres "$DBCON" down
goose postgres "$DBCON" up

