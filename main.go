package main

import (
	"fmt"
	"log"
	"theses/filereader"
	"theses/harvest"
)

// tab-delimited input file.
const inputFile = "../cst_theses.txt"
// json file created at runtime and used by the harvester.
const jsonFile = "theses-data.json"
// config file that provides the worldcat api key (See README).
const apikeyfile = "wskey.json"
// the output directory.
const outputDirectory = "/Users/michaelspalti/willamette/cst/cst_thesis_harvest"

func convertToJsonFile(input string, jsonFile string) {
	_, err := filereader.InputFileConverter(input, jsonFile)
	if (err != nil) {
		log.Fatal(err)
	}
}

func main() {
	log.SetPrefix("harvester: ")
	log.SetFlags(0)
	convertToJsonFile(inputFile, jsonFile)
	apiKey, err := filereader.ReadApiKey(apikeyfile)
	if err != nil {
		fmt.Println(err)
	}
	if apiKey == "" {
		fmt.Println("the api key field is an empty string, harvesting Internet Archive records only")
	}
	harvestResult, err := harvest.HarvestData(jsonFile, outputDirectory, apiKey)
	if (err != nil) {
		log.Fatal(err)
	}
	fmt.Println(harvestResult)
}
