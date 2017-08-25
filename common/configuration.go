package common

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/fatih/color"
)

type Configuration struct {
	Name            string `json:"name"`
	Interval        int    `json:"repeat-interval"`
	WaitForComplete bool   `json:"wait-for-complete"`
	Enabled         bool   `json:"enabled"`
}

func GetConfiguration() ([]Configuration, error) {
	defer color.Unset()

	color.Set(color.FgGreen)

	path := os.Getenv("PARSER_CONFIG_LOCATION")
	if len(path) < 1 {
		path = "./config.json"
	}
	log.Printf("Getting configuration from %v", path)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		color.Set(color.FgHiRed)
		log.Printf("There was an error: %v", err.Error())
		return []Configuration{}, err
	}
	log.Printf("File read.")

	toReturn := []Configuration{}

	err = json.Unmarshal(b, &toReturn)
	if err != nil {
		color.Set(color.FgHiRed)
		log.Printf("There was an error: %v", err.Error())
		return []Configuration{}, err
	}

	log.Printf("Done.")
	return toReturn, nil
}
