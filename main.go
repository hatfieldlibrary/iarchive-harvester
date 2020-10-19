package main
/*
Takes a tab-delimited files as input.  Required fields are title, Internet Archive ID, and OCLC number.

The title field is used in logging output because the title in Internet Archive metadata is usually incomplete.
If title is unavailable modify the program to log the Internet Archive title instead.

The OCLC number is used to harvest additional metadata from WorldCat. This is optional. To harvest WorldCat metadata,
you need to provide a WorldCat Search API key. Besides an authoritative title, WorldCat metadata includes
information that may be worth adding to a digital repository record.
*/
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
		noKey := fmt.Sprintf("WARNING: If this is not what you want, stop program and create the %s file. ",
			apikeyfile)
		fmt.Println(noKey)
	}
	if apiKey == "" {
		fmt.Println("the api key field is an empty string, harvesting Internet Archive records only")
	}
	harvestResult, err := harvest.HarvestData(jsonFile, outputDirectory, apiKey)
	fmt.Println(harvestResult)
}
