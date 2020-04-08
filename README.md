# two-phase commit protocol with postgres db on native golang

Trying to do this one:

```
BEGIN;
update mytable set a_col='something' where red_id=1000;
PREPARE TRANSACTION 'foobar';
COMMIT PREPARED 'foobar';

BEGIN;
update mytable set a_col='something' where red_id=1000;
PREPARE TRANSACTION 'foobar';
ROLLBACK PREPARED 'foobar';
```

[Fantastic GORM example](http://gorm.io/docs/transactions.html)

## two-phase commit command line client 

```
NAME:
   two-phase-commit - two-phase-commit command line client

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   0.0.1

AUTHOR:
   Mariia Tuzovska

COMMANDS:
   booking, b, book  Booking command for Fly Booking & Hotel Booking
   help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

COPYRIGHT:
   2020, mariiatuzovska
```

**two-phase commit command line client** has one command **booking** with alliases *book*, *b*:

```
NAME:
   main booking - Booking command for Fly Booking & Hotel Booking

USAGE:
   main booking [command options] [arguments...]

OPTIONS:
   --fly-scheme value    Path to Fly Booking scheme Configuration (default: "configs/fly.json")
   --hotel-scheme value  Path to Holel Booking scheme Configuration (default: "configs/hotel.json")
   --ClientName value    Client Name
   --FlyNumber value     Fly Number for Fly Booking
   --From value          Place From for Fly Booking
   --To value            Place To for Fly Booking
   --Date value          Date of Fly Booking, for example 2006-01-02
   --HotelName value     Hotel Name for Hotel Booking
   --Arrival value       Arrival date of Hotel Booking, for example 2006-01-02
   --Departure value     Departure date of Hotel Booking, for example 2006-01-02
```

## two-phase commit running with docker

```
$ mkdir -p custom_postgres
$ echo '#!/bin/bash' > custom_postgres/build.sh
$ echo 'sed -i "s/^.*max_prepared_transactions\s*=\s*\(.*\)$/max_prepared_transactions = 2/" "$PGDATA"/postgresql.conf' >> custom_postgres/build.sh
$ chmod +x custom_postgres/build.sh
```

PostgreSQL container in DOCKER:

`$docker run -d --name my-postgres -e POSTGRES_PASSWORD=mysecretpassword postgres -c 'max_prepared_transactions=2' -c 'shared_buffers=256MB' -c 'max_connections=10'`

`docker exec -it 6eed7d7278ac bash`

`root@6eed7d7278ac:/# psql -h localhost -p 5432 -U postgres -d postgres`

POSTRGES:

`postgres=# CREATE DATABASE fly_booking;`

`postgres=# CREATE DATABASE hotel_booking;`

`postgres=# CREATE DATABASE money;`

`exit`

`exit`

GO tools:

`go build`

`./two-pc book` or `./two-pc booking --ClientName Mariia --FlyNumber fu123 --From Kyiv --To Odessa --Date 12-06-2020 --HotelName BeutyHotel --Arrival 12-06-2020 --Departure 13-06-2020`

anyway... 

[https://github.com/docker-library](https://github.com/docker-library/docs/commit/b1d90dbda8c85e10e7cb010df35d4989039b700d?short_path=f04184d#diff-f04184d2552bd93f5562b6437f3627d1)