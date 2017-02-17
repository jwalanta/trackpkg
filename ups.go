package main

import (
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Ups UPS handler
type Ups struct {
}

// Name returns carrier name
func (c Ups) Name() string {
	return "ups"
}

// GetTracking get tracking details
func (c Ups) GetTracking(trackingNumber string) (Shipment, error) {

	trackingURL := "https://wwwapps.ups.com/WebTracking/processInputRequest?sort_by=status&tracknums_displayed=1&TypeOfInquiryNumber=T&loc=en_US&track.x=0&track.y=0&InquiryNumber1="
	trackingURL += trackingNumber

	doc, err := goquery.NewDocument(trackingURL)

	if err != nil {
		return Shipment{}, err
	}

	shipment := Shipment{}
	shipment.TrackingNumber = trackingNumber
	shipment.StatusItems = []Status{}

	spaceRegexp := regexp.MustCompile("[[:space:]]+")
	timestampFormat := "01/02/2006 3:04 PM"

	var location, localTime, description string

	// find statuses
	doc.Find(".dataTable td").Each(func(i int, s *goquery.Selection) {

		status := Status{}

		text := strings.TrimSpace(s.Text())
		text = spaceRegexp.ReplaceAllString(text, " ")

		// there are four columns in UPS tracking: location, date, time, description
		switch i % 4 {
		case 0:
			// if the location is empty, use the previous location
			if text != "" {
				location = text
			}
		case 1:
			localTime = text
		case 2:
			localTime = localTime + " " + text
		case 3:
			description = text

			// parse time and convert to unixtime
			localTime = strings.Replace(localTime, ".", "", -1) // A.M. to AM
			timestamp, _ := time.Parse(timestampFormat, localTime)

			// put everything in status
			status.Timestamp = timestamp.Unix()
			status.Location = location
			status.Description = description

			shipment.StatusItems = append(shipment.StatusItems, status)

		}

	})

	// find delivery date
	doc.Find(".gradientGroup4 dl").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		text = spaceRegexp.ReplaceAllString(text, " ")

		if shipment.DeliveryTimestamp == 0 && strings.Contains(text, "day,") {
			dateRegexp := regexp.MustCompile(`[0-9/]+`)
			dateStr := dateRegexp.FindString(text)

			timestamp, _ := time.Parse(timestampFormat, dateStr+" 00:00 AM")
			shipment.DeliveryTimestamp = timestamp.Unix()
		}
	})

	if len(shipment.StatusItems) > 0 && strings.Contains(shipment.StatusItems[0].Description, "Delivered") {
		shipment.Delivered = true
		shipment.DeliveryTimestamp = shipment.StatusItems[0].Timestamp
	}

	return shipment, nil
}
