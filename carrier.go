package main

// Carrier interface common to all carrier types
type Carrier interface {
	Name() string
	GetTracking(string) (Shipment, error)
}
