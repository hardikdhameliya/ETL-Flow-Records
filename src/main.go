package main

import (
	"gigamon.com/config"
	"gigamon.com/fetch"
	"gigamon.com/upload"
)

func main() {
	// Get the buffer size
	buffer, err := config.GetSectionParamInt("buffer.size")
	if err != nil {
		buffer = 3
	}
	// Set channels for data with buffer size
	data := make(chan map[string][]interface{}, buffer)

	//	Set the channel to terminate the execution
	done := make(chan bool)

	// New instance of fetch to fetch data files
	fch := fetch.NewFetch()

	// Executes go routines to fetch data files and upload the documents
	go fch.Run(data, done)
	go upload.Run(data)

	//Terminate the execution in case of timeout
	<-done
}
