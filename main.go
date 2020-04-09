package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/urfave/cli"
)

type (
	Configuration struct {
		Shema    string
		User     string
		Password string
		Host     string
		Port     string
	}
	FlyBooking struct {
		gorm.Model
		ClientName          string
		FlyNumber, From, To string
		Date                time.Time
	}
	HotelBooking struct {
		gorm.Model
		ClientName         string
		HotelName          string
		Arrival, Departure time.Time
	}
	Money struct {
		gorm.Model
		ClientName string  `gorm:"unique;not_null"`
		Amount     float32 `gorm"check(amount>0)`
	}
)

var (
	Version     string = "0.0.1"
	ServiceName string = "two-phase-commit"
	DBPostgres  string = "postgres"
	layoutISO   string = "2006-01-02"
)

// /usr/local/var/postgres/postgresql.conf

// /Users/mariiaandreevna/Projects/mygo/two-pc

func main() {

	app := cli.NewApp()
	app.Name = ServiceName
	app.Usage = fmt.Sprintf("%s command line client", ServiceName)
	app.Description = ""
	app.Version = Version
	app.Copyright = "2020, mariiatuzovska"
	app.Authors = []cli.Author{cli.Author{Name: "Mariia Tuzovska"}}
	app.Commands = []cli.Command{
		{
			Name:   "init",
			Usage:  "Initialize client",
			Action: inite,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "money-db",
					Usage: "Path to Holel Booking scheme Configuration",
					Value: "configs/money.json",
				},
				&cli.StringFlag{
					Name:  "ClientName",
					Usage: "Client Name",
				},
			},
		},
		{
			Name:    "booking",
			Usage:   "Booking command for Fly Booking & Hotel Booking",
			Aliases: []string{"b", "book"},
			Action:  book,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "fly-db",
					Usage: "Path to Fly Booking db Configuration",
					Value: "configs/fly.json",
				},
				&cli.StringFlag{
					Name:  "hotel-db",
					Usage: "Path to Holel Booking db Configuration",
					Value: "configs/hotel.json",
				},
				&cli.StringFlag{
					Name:  "money-db",
					Usage: "Path to Holel Booking db Configuration",
					Value: "configs/money.json",
				},
				&cli.StringFlag{
					Name:  "ClientName",
					Usage: "Client Name",
				},
				&cli.StringFlag{
					Name:  "FlyNumber",
					Usage: "Fly Number for Fly Booking",
				},
				&cli.StringFlag{
					Name:  "From",
					Usage: "Place From for Fly Booking",
				},
				&cli.StringFlag{
					Name:  "To",
					Usage: "Place To for Fly Booking",
				},
				&cli.StringFlag{
					Name:  "Date",
					Usage: fmt.Sprintf("Date of Fly Booking, for example %s", layoutISO),
				},
				&cli.StringFlag{
					Name:  "HotelName",
					Usage: "Hotel Name for Hotel Booking",
				},
				&cli.StringFlag{
					Name:  "Arrival",
					Usage: fmt.Sprintf("Arrival date of Hotel Booking, for example %s", layoutISO),
				},
				&cli.StringFlag{
					Name:  "Departure",
					Usage: fmt.Sprintf("Departure date of Hotel Booking, for example %s", layoutISO),
				},
			},
		},
		{
			Name:   "del",
			Usage:  "Initialize client",
			Action: del,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "fly-db",
					Usage: "Path to Fly Booking db Configuration",
					Value: "configs/fly.json",
				},
				&cli.StringFlag{
					Name:  "hotel-db",
					Usage: "Path to Holel Booking db Configuration",
					Value: "configs/hotel.json",
				},
				&cli.StringFlag{
					Name:  "money-db",
					Usage: "Path to Holel Booking db Configuration",
					Value: "configs/money.json",
				},
			},
		},
	}

	app.Run(os.Args)
}

func inite(c *cli.Context) error {
	if c.String("ClientName") == "" {
		log.Fatal("Nil clients")
	}
	moneyDB, err := newDBConnection(c.String("money-db"))
	if err != nil {
		log.Fatal(err)
	}
	moneyDB.AutoMigrate(&Money{})
	m := &Money{}
	if moneyDB.First(&m,
		&Money{
			ClientName: c.String("ClientName"),
		}).RecordNotFound() {
		err = moneyDB.Create(
			&Money{
				ClientName: c.String("ClientName"),
				Amount:     100,
			}).Error
		if err != nil {
			log.Fatal(err)
		}
		log.Println(fmt.Sprintf("DEBUG | created new clien ClientName=%s Amount=100", c.String("ClientName")))
	} else {
		m.Amount += 50.0
		err = moneyDB.First(&Money{}, m.ID).Update(m).Error
		if err != nil {
			log.Fatal(err)
		}
		log.Println(fmt.Sprintf("DEBUG | updated clien ClientName=%s Amount=%f", c.String("ClientName"), m.Amount))
	}
	return nil
}

func book(c *cli.Context) error {

	// connection and migration to Fly DB
	flyDB, err := newDBConnection(c.String("fly-db"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DEBUG | Fly DB connected")
	flyDB.AutoMigrate(&FlyBooking{})
	log.Println("DEBUG | FlyBooking migrated")

	// connection and migration to Money DB
	moneyDB, err := newDBConnection(c.String("money-db"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DEBUG | Money DB connected")
	moneyDB.AutoMigrate(&Money{})
	log.Println("DEBUG | Money migrated")

	// connection and migration to Hotel DB
	hotelDB, err := newDBConnection(c.String("hotel-db"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DEBUG | Hotel DB connected")
	hotelDB.AutoMigrate(&HotelBooking{})
	log.Println("DEBUG | HotelBooking migrated")

	flyRecord := flyFilling(c)
	// if moneyDB.First(&Money{}, &Money{ClientName: flyRecord.ClientName}).RecordNotFound() {
	// 	log.Fatal("\nNo money -- no honey\n")
	// }
	log.Println("DEBUG | Fly record filled")
	// begin a transaction for Fly Booking
	flyTX := flyDB.Begin()
	if err := flyTX.Error; err != nil {
		log.Println("DEBUG | flyTX has been begun with error")
		log.Fatal(err)
	}
	log.Println("DEBUG | flyTX has been begun")
	if err := flyTX.Create(flyRecord).Error; err != nil {
		// rollback the transaction in case of error
		flyTX.Rollback()
		log.Println("DEBUG | flyTX rollbecked")
		log.Fatal(err)
	}
	log.Println("DEBUG | flyTX has been created")
	if err := flyTX.Exec("PREPARE TRANSACTION 'foobar'").Error; err != nil {
		flyTX.Rollback()
		log.Println("DEBUG | flyTX has been prepared with error")
		log.Fatal(err)
	}
	log.Println("DEBUG | flyTX has been prepared")

	hotelRecord := hotelFilling(c)
	hotelRecord.ClientName = flyRecord.ClientName
	// begin a transaction for Hotel Booking
	hotelTX := hotelDB.Begin()
	log.Println("DEBUG | hotelTX has been begun")
	if err := hotelTX.Error; err != nil {
		log.Println("DEBUG | hotelTX has been begun with error")
		log.Fatal(err)
	}
	if err := hotelTX.Create(hotelRecord).Error; err != nil {
		// rollback the transaction in case of error
		log.Println("DEBUG | hotelTX create with error")
		hotelTX.Rollback()
		log.Fatal(err)
	}
	if err := hotelTX.Exec("PREPARE TRANSACTION 'foobar'").Error; err != nil {
		hotelTX.Rollback()
		log.Println("DEBUG | hotelTX has been prepared with error")
		log.Fatal(err)
	}
	log.Println("DEBUG | hotelTX has been prepared")

	// cash := Money{}
	// if moneyDB.First(&cash, &Money{
	// 	ClientName: flyRecord.ClientName,
	// }).RecordNotFound() {
	// 	// rollback the transaction in case of error
	// 	flyTX.Rollback()
	// 	log.Println("DEBUG | flyTX rollbecked")
	// 	hotelTX.Rollback()
	// 	log.Println("DEBUG | hotelTX rollbecked")
	// 	log.Fatal(err)
	// }
	// cash.Amount -= 100

	moneyTX := moneyDB.Begin()
	if err := moneyTX.Error; err != nil {
		log.Println("DEBUG | moneyTX has been begun with error")
		log.Fatal(err)
	}
	log.Println("DEBUG | moneyTX has been begun")
	if err := moneyTX.Update(&Money{}, gorm.Expr("amount - ?", 100)).Error; err != nil {
		moneyTX.Rollback()
		log.Println("DEBUG | moneyTX has been prepared with error")
		flyTX.Rollback()
		log.Println("DEBUG | flyTX rollbecked")
		hotelTX.Rollback()
		log.Println("DEBUG | hotelTX rollbecked")
		log.Fatal(err)
	}
	log.Println("DEBUG | moneyTX has been created")
	if err := moneyTX.Exec("PREPARE TRANSACTION 'foobar'").Error; err != nil {
		moneyTX.Rollback()
		log.Println("DEBUG | moneyTX has been prepared with error")
		flyTX.Rollback()
		log.Println("DEBUG | flyTX rollbecked")
		hotelTX.Rollback()
		log.Println("DEBUG | hotelTX rollbecked")
		log.Fatal(err)
	}
	log.Println("DEBUG | moneyTX has been prepared")

	err = flyTX.Commit().Error
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DEBUG | flyTX commited")
	err = hotelTX.Commit().Error
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DEBUG | hotelTX commited")
	err = moneyTX.Commit().Error
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DEBUG | moneyTX commited")

	return nil
}

func newDBConnection(path string) (*gorm.DB, error) {

	config := new(Configuration)
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return gorm.Open(DBPostgres, fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Shema, config.Password))
}

func flyFilling(c *cli.Context) *FlyBooking {
	var (
		err  error
		date string
	)
	fly := FlyBooking{
		ClientName: c.String("ClientName"),
		FlyNumber:  c.String("FlyNumber"),
		From:       c.String("From"),
		To:         c.String("To"),
	}
	if fly.ClientName == "" {
		fmt.Println("Fly Booking | Enter Client Name")
		fmt.Scan(&fly.ClientName)
	}
	if fly.FlyNumber == "" {
		fmt.Println("Fly Booking | Enter Fly Number")
		fmt.Scan(&fly.FlyNumber)
	}
	if fly.From == "" {
		fmt.Println("Fly Booking | Enter From")
		fmt.Scan(&fly.From)
	}
	if fly.To == "" {
		fmt.Println("Fly Booking | Enter To")
		fmt.Scan(&fly.To)
	}
	if c.String("Date") == "" {
		fmt.Println("Fly Booking | Enter Date")
		fmt.Scan(&date)
	} else {
		date = c.String("Date")
	}
	fly.Date, err = time.Parse(layoutISO, date)
	if err != nil {
		fly.Date = time.Now()
	}
	return &fly
}

func hotelFilling(c *cli.Context) *HotelBooking {
	var (
		err       error
		arrival   string
		departure string
	)
	hotel := HotelBooking{
		HotelName: c.String("HotelName"),
	}
	if hotel.HotelName == "" {
		fmt.Println("Hotel Booking | Enter Hotel Name")
		fmt.Scan(&hotel.HotelName)
	}
	if c.String("Arrival") == "" {
		fmt.Println("Hotel Booking | Enter Arrival")
		fmt.Scan(&arrival)
	} else {
		arrival = c.String("Arrival")
	}
	if c.String("Departure") == "" {
		fmt.Println("Hotel Booking | Enter Departure")
		fmt.Scan(&departure)
	} else {
		departure = c.String("Departure")
	}
	hotel.Arrival, err = time.Parse(layoutISO, arrival)
	if err != nil {
		hotel.Arrival = time.Now()
	}
	hotel.Departure, err = time.Parse(layoutISO, departure)
	if err != nil {
		hotel.Departure = time.Now().Add(24 * time.Hour)
	}
	return &hotel
}

func del(c *cli.Context) error {
	// connection and migration to Fly DB
	flyDB, err := newDBConnection(c.String("fly-db"))
	if err != nil {
		log.Fatal(err)
	}
	flyDB.DropTableIfExists(&FlyBooking{})

	// connection and migration to Money DB
	moneyDB, err := newDBConnection(c.String("money-db"))
	if err != nil {
		log.Fatal(err)
	}
	moneyDB.DropTableIfExists(&Money{})

	// connection and migration to Hotel DB
	hotelDB, err := newDBConnection(c.String("hotel-db"))
	if err != nil {
		log.Fatal(err)
	}
	hotelDB.DropTableIfExists(&HotelBooking{})
	return nil
}
