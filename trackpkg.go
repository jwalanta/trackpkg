package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	"os"
	"strings"
)

const trackingRepo = "~/.trackingrepo"

var trackingRepoPath, _ = homedir.Expand(trackingRepo)

var supportedCarriers = []Carrier{Fedex{}, Ups{}, Usps{}}

var repository = ShipmentsRepository{Path: trackingRepoPath}

var commands = []*cli.Command{
	{
		Name:         "add",
		Usage:        "adds new tracking number",
		ArgsUsage:    "<tracking_number> <description>",
		Action:       Add,
		BashComplete: TrackpkgAutocomplete,
	},
	{
		Name:         "remove",
		Usage:        "removes tracking number",
		ArgsUsage:    "<tracking_number> | <item_number>",
		Action:       Remove,
		BashComplete: TrackpkgAutocomplete,
	},
	{
		Name:         "update",
		Usage:        "updates status of all tracked numbers",
		Action:       Update,
		BashComplete: TrackpkgAutocomplete,
	},
	{
		Name:         "list",
		Usage:        "lists status of a given or all tracked number",
		ArgsUsage:    "[<tracking_number> | <item_number>]",
		Action:       List,
		BashComplete: TrackpkgAutocomplete,
	},
	{
		Name:         "detail",
		Usage:        "prints detail tracking status of a given or all tracked numbers",
		ArgsUsage:    "[<tracking_number> | <item_number>]",
		Action:       Detail,
		BashComplete: TrackpkgAutocomplete,
	},
	{
		Name:         "clean",
		Usage:        "removes delivered packages",
		Action:       Clean,
		BashComplete: TrackpkgAutocomplete,
	},
}

// Add adds new tracking number
func Add(context *cli.Context) error {

	trackingNumber := context.Args().First()
	description := strings.Join(context.Args().Tail(), " ")

	if trackingNumber == "" {
		return nil
	}

	shipments, _ := repository.load()
	shipment := Shipment{TrackingNumber: trackingNumber, Description: description, Delivered: false}
	shipments.addItem(shipment)
	err := repository.save(shipments)

	return err
}

// Remove removes tracking number
func Remove(context *cli.Context) error {
	trackingNumber := context.Args().First()

	if trackingNumber == "" {
		return nil
	}

	shipments, _ := repository.load()
	shipments.removeItem(trackingNumber)

	err := repository.save(shipments)

	return err
}

// Update updates status of all tracked numbers
func Update(context *cli.Context) error {
	shipments, err := repository.load()

	shipments.updateTracking(supportedCarriers)

	err = repository.save(shipments)

	return err
}

// List lists status of a given or all tracked number
func List(context *cli.Context) error {

	shipments, err := repository.load()
	shipments.list("", false)

	return err
}

// Detail prints detail tracking status of a given or all tracked numbers
func Detail(context *cli.Context) error {

	trackingNumber := context.Args().First()

	shipments, err := repository.load()
	shipments.list(trackingNumber, true)

	return err
}

// Clean removes delivered packages
func Clean(context *cli.Context) error {

	shipments, err := repository.load()

	shipments.removeDelivered()

	err = repository.save(shipments)

	return err
}

// TrackpkgAutocomplete bash autocomplete
func TrackpkgAutocomplete(context *cli.Context) {
}

func main() {

	app := cli.NewApp()

	app.Name = "Trackpkg"
	app.Usage = "Track Shipment Packages"
	app.Version = "0.1"
	// app.Authors = []cli.Author{
	// 	cli.Author{
	// 		Name:  "Jwalanta Shrestha",
	// 		Email: "jwalanta@gmail.com",
	// 	},
	// }
	app.EnableBashCompletion = true
	app.Commands = commands

	err := app.Run(os.Args)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
