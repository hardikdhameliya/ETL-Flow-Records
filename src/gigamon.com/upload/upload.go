package upload

import (
	"os"
	"time"

	"gigamon.com/config"
)

var db DataStore

// Initialize datastore with configuration parameters
func init() {
	var app_name, url, db_name string
	var err error
	var timeout time.Duration

	if db_name, err = config.GetSectionParamString("upload.name"); err != nil {
		db_name = "flow"
	}
	if app_name, err = config.GetSectionParamString("upload.app"); err != nil {
		app_name = "flow_records"
	}

	if url, err = config.GetSectionParamString("upload.url"); err != nil {
		url = "mongodb://localhost:27017"
	}

	if timeout, err = config.GetSectionParamDuration("upload.timeout"); err != nil {
		timeout = 10 * time.Second
	}

	db = NewDb(db_name, app_name, url, timeout)
	if db == nil {
		os.Exit(1)
	}
}

// Receive data from the channel and store into the datastore
func Run(m chan map[string][]interface{}) {

	for {
		flows := <-m
		db.Set(flows)
	}
}
