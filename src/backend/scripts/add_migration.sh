#!/bin/bash

if [ -z "$1" ]; then
  echo "Error: Migration name is required"
  echo "Usage: ./scripts/add_migration.sh <migration_name>"
  exit 1
fi

migrate create -ext sql -seq -dir ./sql/migrations "$1"

#usage ./scripts/add_migration.sh base_structure_migration