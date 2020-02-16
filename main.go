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
		ClientName, FlyNumber, From, To string
		Date                            time.Time
	}
	HotelBooking struct {
		gorm.Model
		ClientName, HotelName string
		Arrival, Departure    time.Time
	}
)

var (
	Version     string = "0.0.1"
	ServiceName string = "two-phase-commit"
	DBPostgres  string = "postgres"
	layoutISO   string = "2006-01-02"
)

func main() {

	app := cli.NewApp()
	app.Name = ServiceName
	app.Usage = fmt.Sprintf("%s command line client", ServiceName)
	// app.Description = ""
	app.Version = Version
	app.Copyright = "2020, mariiatuzovska"
	app.Authors = []cli.Author{cli.Author{Name: "Mariia Tuzovska"}}
	app.Commands = []cli.Command{
		{
			Name:    "booking",
			Usage:   "Booking command for Fly Booking & Hotel Booking",
			Aliases: []string{"b", "book"},
			Action:  book,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "fly-scheme",
					Usage: "Path to Fly Booking scheme Configuration",
					Value: "configs/fly.json",
				},
				&cli.StringFlag{
					Name:  "hotel-scheme",
					Usage: "Path to Holel Booking scheme Configuration",
					Value: "configs/hotel.json",
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
	}

	app.Run(os.Args)
}

func book(c *cli.Context) error {

	// connection and migration for Fly Booking
	flyDB, err := newDBConnection(c.String("fly-scheme"))
	if err != nil {
		log.Fatal(err)
	}
	flyDB.AutoMigrate(&FlyBooking{})
	flyRecord := flyFilling(c)

	// begin a transaction for Fly Booking
	flyTX := flyDB.Begin()
	if err := flyTX.Error; err != nil {
		log.Fatal(err)
	}
	if err := flyTX.Create(flyRecord).Error; err != nil {
		// rollback the transaction in case of error
		flyTX.Rollback()
		log.Fatal(err)
	}

	var command string

	fmt.Println("Commit Fly Booking transaction? Enter yes/y to commit the Fly Booking transaction or book the hotel! Enter n/no to stop.")
	fmt.Scan(&command)
	if command == "y" || command == "yes" {
		// commit the transaction
		err := flyTX.Commit().Error
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Fly Booking transaction is committed")
	} else if command == "n" || command == "no" {
		// rollback the transaction
		flyTX.Rollback()
		fmt.Println("Fly Booking transaction is rollbacked")
	} else {
		// connection and migration for Hotel Booking
		hotelDB, err := newDBConnection(c.String("hotel-scheme"))
		if err != nil {
			log.Fatal(err)
		}
		hotelDB.AutoMigrate(&HotelBooking{})
		hotelRecord := hotelFilling(c)
		hotelRecord.ClientName = flyRecord.ClientName
		// begin a transaction for Hotel Booking
		hotelTX := hotelDB.Begin()
		if err := hotelTX.Error; err != nil {
			log.Fatal(err)
		}
		if err := hotelTX.Create(hotelRecord).Error; err != nil {
			// rollback the transaction in case of error
			hotelTX.Rollback()
			log.Fatal(err)
		}
		command = ""
		fmt.Println("Commit Fly Booking & Hotel Booking transactions? Enter yes/y to commit the transactions.")
		fmt.Scan(&command)
		if command == "y" || command == "yes" {
			// commit the transaction
			err := flyTX.Commit().Error
			if err != nil {
				log.Fatal(err)
			}
			err = hotelTX.Commit().Error
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Fly Booking & Hotel Booking transactions are committed")
		}
	}

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

	return gorm.Open(DBPostgres,
		DBPostgres+"://"+config.User+":"+config.Password+"@"+config.Host+":"+config.Port+"/"+config.Shema+"?sslmode=disable")
}

func flyFilling(c *cli.Context) *FlyBooking {

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
	var (
		err  error
		date string
	)
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

	hotel := HotelBooking{
		// ClientName: c.String("ClientName"),
		HotelName: c.String("HotelName"),
	}
	// if hotel.ClientName == "" {
	// 	fmt.Println("Hotel Booking | Enter Client Name")
	// 	fmt.Scan(&hotel.ClientName)
	// }
	if hotel.HotelName == "" {
		fmt.Println("Hotel Booking | Enter Hotel Name")
		fmt.Scan(&hotel.HotelName)
	}
	var (
		err       error
		arrival   string
		departure string
	)
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
