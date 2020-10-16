package main

import (
	"fmt"
	"log"
	"theses/filereader"
	"theses/harvest"
)

// A tab-delimited file.
const input = "../test_cst_file.txt"
// This json file will be created at runtime.
const jsonFile = "theses-data.json"
// Create this file to load worldcat data (See README)
const apikeyfile = "wskey.json"
// Set the the output directory here.
const outputDirectory = "/Users/michaelspalti/willamette/cst/cst_harvest_test"

func convertToJsonFile(input string, jsonFile string) {
	_, err := filereader.InputFileConverter(input, jsonFile)
	if (err != nil) {
		log.Fatal(err)
	}
}

func main() {
	log.SetPrefix("harvester: ")
	log.SetFlags(0)
	convertToJsonFile(input, jsonFile)
	apiKey, err := filereader.ReadApiKey(apikeyfile)
	if err != nil {
		fmt.Println(err)
	}
	harvestResult, err := harvest.HarvestData(jsonFile, outputDirectory, apiKey)
	if (err != nil) {
		log.Fatal(err)
	}
	fmt.Println(harvestResult)
}
