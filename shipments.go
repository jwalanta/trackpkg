package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Shipments collection of shipments
type Shipments struct {
	Items []Shipment
}

// Shipment data struct for individual Shipment
type Shipment struct {
	TrackingNumber    string
	Description       string
	Carrier           string
	Delivered         bool
	DeliveryTimestamp int64
	StatusItems       []Status
}

// Status individual status details of a shipment
type Status struct {
	Timestamp   int64
	Location    string
	Description string
}

func (shipments *Shipments) addItem(shipment Shipment) error {
	shipments.Items = append(shipments.Items, shipment)
	return nil
}

func (shipments *Shipments) removeItem(id string) error {
	for i, shipment := range shipments.Items {
		if strconv.Itoa(i+1) == id || shipment.TrackingNumber == id {
			shipments.Items = append(shipments.Items[:i], shipments.Items[i+1:]...)
			break
		}
	}
	return nil
}

// this is where the magic happens
func (shipments *Shipments) updateTracking(carriers []Carrier) error {

	for i, shipment := range shipments.Items {

		fmt.Printf("%3d %-25s ", (i + 1), shipment.TrackingNumber)

		// skip delivered shipment
		if shipment.Delivered {
			fmt.Printf("%-6s %s\n", strings.ToUpper(shipment.Carrier), "Already Delivered")
			continue
		}

		changed := false

		for _, carrier := range carriers {

			// if shipment's carrier has already been determined
			// and it not this one, then skip
			if shipment.Carrier != "" && carrier.Name() != shipment.Carrier {
				continue
			}

			// try getting tracking information
			updatedShipment, err := carrier.GetTracking(shipment.TrackingNumber)

			if err == nil && len(updatedShipment.StatusItems) > 0 {
				if len(updatedShipment.StatusItems) > len(shipment.StatusItems) ||
					shipment.Delivered != updatedShipment.Delivered ||
					shipment.DeliveryTimestamp != updatedShipment.DeliveryTimestamp {

					changed = true

					shipments.Items[i].Carrier = carrier.Name()
					shipments.Items[i].Delivered = updatedShipment.Delivered
					shipments.Items[i].DeliveryTimestamp = updatedShipment.DeliveryTimestamp

					if len(updatedShipment.StatusItems) > len(shipments.Items[i].StatusItems) {
						shipments.Items[i].StatusItems = updatedShipment.StatusItems
					}

				}

				break

			}

		}

		if shipments.Items[i].Carrier == "" {
			fmt.Printf("%-6s %s\n", "???", "Carrier cannot be determined")
		} else {

			if changed {
				status := shipments.Items[i].StatusItems[0]
				latestStatus := fmt.Sprintf(" - %s %s %s", time.Unix(status.Timestamp, 0).Format("01/02 03:04 PM"), status.Location, status.Description)
				fmt.Printf("%-6s %s\n", strings.ToUpper(shipments.Items[i].Carrier), "Updated"+latestStatus)
			} else {
				fmt.Printf("%-6s %s\n", strings.ToUpper(shipments.Items[i].Carrier), "No change in status")
			}

		}
	}

	return nil
}

// list the shipments
func (shipments *Shipments) list(id string, showDetail bool) error {

	listFormat := "%3d %-25s %-6s %-10s %s\n" // #, tracking number, carrier, delivery status, description
	detailFormat := "    - %-15s %-30s %s\n"  // date, location, status

	for i, shipment := range shipments.Items {

		if id != "" && !(strconv.Itoa(i+1) == id || shipment.TrackingNumber == id) {
			continue
		}

		deliveryStatus := ""

		if shipment.Delivered {
			deliveryStatus = "Delivered"
		} else {
			if shipment.DeliveryTimestamp > 0 {

				deliveryTimestamp := time.Unix(shipment.DeliveryTimestamp, 0).UTC()
				todayMidnight, _ := time.Parse("2006/01/02", time.Now().Format("2006/01/02"))

				daysRemaining := deliveryTimestamp.Sub(todayMidnight).Hours() / 24

				if daysRemaining < 0 {
					daysRemaining = 0
				}
				deliveryStatus += fmt.Sprintf("%s +%.0fd", deliveryTimestamp.Format("01/02"), daysRemaining)
			}
		}

		fmt.Printf(listFormat, (i + 1), shipment.TrackingNumber, strings.ToUpper(shipment.Carrier), deliveryStatus, shipment.Description)

		if showDetail {
			for _, status := range shipment.StatusItems {
				fmt.Printf(detailFormat, time.Unix(status.Timestamp, 0).Format("01/02 03:04 PM"), status.Location, status.Description)
			}
			if len(shipment.StatusItems) > 0 {
				fmt.Println()
			}
		}

	}

	return nil

}

// removes delivered entries
func (shipments *Shipments) removeDelivered() error {

	undeliveredShipments := []Shipment{}

	for _, shipment := range shipments.Items {
		if !shipment.Delivered {
			undeliveredShipments = append(undeliveredShipments, shipment)
		}
	}

	shipments.Items = undeliveredShipments

	return nil
}
