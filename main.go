package main

import (
	"fmt"
	"log"
	"theses/filereader"
	"theses/harvest"
)

const input = "../test_cst_file.txt"
const jsonFile = "theses-data.json"
const outputDirectory = "/Users/michaelspalti/willamette/cst/cst_harvest_test"

func convertToJsonFile(input string, jsonFile string) {
	_, err := filereader.InputFileConverter(input, jsonFile)
	if (err != nil) {
		log.Fatal(err)
	}
}

func main() {
	// Set properties of the predefined Logger, including
	// the log entry prefix and a flag to disable printing
	// the time, source file, and line number.
	log.SetPrefix("file: ")
	log.SetFlags(0)
	convertToJsonFile(input, jsonFile)

	harvestResult, err := harvest.HarvestData(jsonFile, outputDirectory)
	if (err != nil) {
		log.Fatal(err)
	}
	fmt.Println(harvestResult)
}
