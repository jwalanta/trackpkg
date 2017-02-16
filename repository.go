package main

import (
	"encoding/json"
	"io/ioutil"
)

// RepositoryInterface interface for handling shipments repository
type RepositoryInterface interface {
	load() (Shipments, error)
	save(s Shipments) error
	archive(s Shipments)
}

// ShipmentsRepository struct to read the file
type ShipmentsRepository struct {
	Path string
}

func (r ShipmentsRepository) load() (Shipments, error) {

	shipments := Shipments{}

	jsonString, err := ioutil.ReadFile(r.Path)

	err = json.Unmarshal(jsonString, &shipments)

	if err != nil {
		return shipments, err
	}

	return shipments, nil
}

func (r ShipmentsRepository) save(s Shipments) error {

	b, err := json.MarshalIndent(s, "", "    ")

	if err == nil {
		err = ioutil.WriteFile(r.Path, b, 0644)
	}

	return err

}

func (r ShipmentsRepository) archive() error {
	return nil
}
