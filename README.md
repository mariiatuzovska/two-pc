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

## two-phase commit running

`docker run -d -p 5432:5432 --name my-postgres -e POSTGRES_PASSWORD=mysecretpassword postgres`

`docker exec -it my-postgres bash`

`postgres=# CREATE DATABASE fly_booking;`

`postgres=# CREATE DATABASE hotel_booking;`

then, in the project package:

`go build`

`./two-pc booking --ClientName Mariia --FlyNumber fu123 --From Kyiv --To Odessa --Date 12-06-2020 --HotelName BeutyHotel --Arrival 12-06-2020 --Departure 13-06-2020`

`./two-pc book`

anyway