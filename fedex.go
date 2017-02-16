package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Fedex fedex handler
type Fedex struct {
}

// FedexJSON fedex json
type FedexJSON struct {
	TrackPackagesResponse FedexResponse `json:"TrackPackagesResponse"`
}

// FedexResponse fedex json
type FedexResponse struct {
	FedexPackageList []FedexPackageList `json:"packageList"`
}

// FedexPackageList fedex json
type FedexPackageList struct {
	IsDelivered   bool                 `json:"isDelivered"`
	EstDeliveryDt string               `json:"estDeliveryDt"`
	ScanEventList []FedexScanEventList `json:"scanEventList"`
}

// FedexScanEventList fedex json
type FedexScanEventList struct {
	Date         string `json:"date"`
	Time         string `json:"time"`
	GmtOffset    string `json:"gmtOffset"`
	Status       string `json:"status"`
	ScanLocation string `json:"scanLocation"`
}

// Name returns carrier name
func (c Fedex) Name() string {
	return "fedex"
}

// GetTracking get tracking details
func (c Fedex) GetTracking(trackingNumber string) (Shipment, error) {

	shipment := Shipment{}
	fedexJSON := FedexJSON{}

	jsonString, err := c.getTrackingJSON(trackingNumber)

	if err != nil {
		return shipment, err
	}

	err = json.Unmarshal(jsonString, &fedexJSON)

	if err != nil {
		return shipment, err
	}

	if len(fedexJSON.TrackPackagesResponse.FedexPackageList[0].ScanEventList) == 0 || fedexJSON.TrackPackagesResponse.FedexPackageList[0].ScanEventList[0].Date == "" {
		return shipment, errors.New("No tracking statuses found")
	}

	shipment.TrackingNumber = trackingNumber
	shipment.StatusItems = []Status{}

	for _, event := range fedexJSON.TrackPackagesResponse.FedexPackageList[0].ScanEventList {

		status := Status{}

		timeStr := event.Date + " " + event.Time + " " + event.GmtOffset
		timeStamp, _ := time.Parse("2006-01-02 15:04:05 -07:00", timeStr)

		status.Timestamp = timeStamp.Unix()
		status.Location = event.ScanLocation
		status.Description = event.Status

		shipment.StatusItems = append(shipment.StatusItems, status)

	}

	if fedexJSON.TrackPackagesResponse.FedexPackageList[0].IsDelivered {
		shipment.Delivered = true
		shipment.DeliveryTimestamp = shipment.StatusItems[0].Timestamp
	} else {
		shipment.Delivered = false
		estDeliveryDt := fedexJSON.TrackPackagesResponse.FedexPackageList[0].EstDeliveryDt
		estDeliveryDt = estDeliveryDt[:strings.Index(estDeliveryDt, "T")]
		estDeliveryDt = strings.Replace(estDeliveryDt, "-", "/", -1)
		timeStamp, _ := time.Parse("2006/01/02", estDeliveryDt)

		shipment.DeliveryTimestamp = timeStamp.Unix()
	}

	return shipment, nil
}

func (c Fedex) getTrackingJSON(trackingNumber string) ([]byte, error) {

	urlStr := "https://www.fedex.com/trackingCal/track"

	data := url.Values{}
	data.Add("data", `{"TrackPackagesRequest":{"appType":"WTRK","appDeviceType":"DESKTOP","uniqueKey":"","processingParameters":{},"trackingInfoList":[{"trackNumberInfo":{"trackingNumber":"`+trackingNumber+`","trackingQualifier":"","trackingCarrier":""}}]}}`)
	data.Add("action", "trackpackages")
	data.Add("locale", "en_US")
	data.Add("version", "1")
	data.Add("format", "json")

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, _ := client.Do(r)

	if resp.Status != "200 OK" {
		return []byte{}, errors.New("Error downloading tracking JSON data")
	}

	body, _ := ioutil.ReadAll(resp.Body)

	return body, nil
}
