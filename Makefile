
go:
	go build -o ./two-pc

open:
	sudo psql -U mari postgres

help:
	@echo "\l -- open list of DBs"
	@echo "\c -- open DB"
	@echo "\dt -- open list of relations"