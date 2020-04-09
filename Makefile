
go:
	go build -o ./two-pc

start:
	pg_ctl -D /usr/local/var/postgres start

stop: 
	pg_ctl -D /usr/local/var/postgres stop

open:
	sudo psql -U mari postgres

help:
	@echo "\l -- open list of DBs"
	@echo "\c -- open DB"
	@echo "\dt -- open list of relations"