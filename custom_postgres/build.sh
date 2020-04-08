#!/bin/bash
sed -i "s/^.*max_prepared_transactions\s*=\s*\(.*\)$/max_prepared_transactions = 2/" "$PGDATA"/postgresql.conf
