package main

import (
	"os"

	"github.com/labstack/gommon/log"
	"gopkg.in/yaml.v2"
)

// ReadConfig get config params from settings.yaml which is in path.
func ReadConfig(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := yaml.NewDecoder(f)

	//var m map[string]interface{}
	var m map[string]string

	if err := d.Decode(&m); err != nil {
		log.Fatal(err)
	}
	return m
}
