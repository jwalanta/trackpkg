package main

import (
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Usps USPS handler
type Usps struct {
}

// Name returns carrier name
func (c Usps) Name() string {
	return "usps"
}

// GetTracking get tracking details
func (c Usps) GetTracking(trackingNumber string) (Shipment, error) {

	shipment := Shipment{}

	trackingURL := "https://tools.usps.com/go/TrackConfirmAction?tLabels=" + trackingNumber

	doc, err := goquery.NewDocument(trackingURL)

	if err != nil {
		return shipment, err
	}

	shipment.TrackingNumber = trackingNumber
	shipment.StatusItems = []Status{}

	spaceRegexp := regexp.MustCompile("[[:space:]]+")
	timestampFormat := "January 2, 2006 , 3:04 pm"

	currentLocation := time.Now().Location()

	var location, localTime, desc string

	doc.Find("#tc-hits .detail-wrapper td").Each(func(i int, s *goquery.Selection) {

		status := Status{}

		text := strings.TrimSpace(s.Text())
		text = spaceRegexp.ReplaceAllString(text, " ")

		switch i % 3 {
		case 0:
			localTime = text
		case 1:
			desc = text
		case 2:
			location = text

			if len(localTime) > 0 && !strings.Contains(localTime, ":") {
				localTime += " , 12:00 am"
			}

			timestamp, _ := time.ParseInLocation(timestampFormat, localTime, currentLocation)

			// put everything in status
			status.Timestamp = timestamp.Unix()
			status.Location = location
			status.Description = desc

			shipment.StatusItems = append(shipment.StatusItems, status)

		}

	})

	doc.Find(".tracking-progress span").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		text = spaceRegexp.ReplaceAllString(text, " ")

		if shipment.DeliveryTimestamp == 0 && strings.Contains(text, "day, ") {

			timeStr := text[strings.Index(text, ", ")+2:] + " , 12:00 am"
			timeStamp, _ := time.ParseInLocation(timestampFormat, timeStr, currentLocation)

			shipment.DeliveryTimestamp = timeStamp.Unix()
		}
	})

	if len(shipment.StatusItems) > 0 && strings.Contains(shipment.StatusItems[0].Description, "Delivered") {
		shipment.Delivered = true
		shipment.DeliveryTimestamp = shipment.StatusItems[0].Timestamp
	}

	return shipment, nil
}
