package filereader

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"theses/types"
)

/*
ReadApiKey returns the key set in the json configuration file. This file is described in README.
 */
func ReadApiKey(input string) (string, error) {
	if (input == "") {
		return "", errors.New("no input file name")
	}
	dat, err := ioutil.ReadFile(input)
	if (err != nil) {
		return "", errors.New(fmt.Sprintf("unable to open api key file: %v", input))
	}
	key := types.ApiKey{}
	_ = json.Unmarshal([]byte(dat), &key)
	return key.Key, nil
}

/*
InputFileConverter extracts information from the tag-delimited input file and creates a json output
file that can be used in harvesting.
*/
func InputFileConverter(input string, output string) (string, error) {

	if (input == "") {
		return "", errors.New("no input file name")
	}
	if (output == "") {
		return "", errors.New("no output file name")
	}
	dat, err := os.Open(input)
	if (err != nil) {
		return "", errors.New(fmt.Sprintf("unable to open file: %v", input))
	}
	defer dat.Close()
	reader := csv.NewReader(dat)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	csvData, err := reader.ReadAll()
	if err != nil {
		fmt.Println("unable to read file")
		fmt.Println(err)
		os.Exit(1)
	}
	var oneRecord types.Thesis
	var allRecords []types.Thesis
	for _, each := range csvData {
		oneRecord.Title = each[0]
		oneRecord.IarchiveID = each[9]
		oneRecord.Oclc = each[24]
		allRecords = append(allRecords, oneRecord)
	}
	jsondata, err := json.Marshal(allRecords) // convert to JSON
	if err != nil {
		fmt.Println("error marshalling records")
		fmt.Println(err)
		os.Exit(1)
	}
	jsonFile, err := os.Create(output)

	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	jsonFile.Write(jsondata)
	jsonFile.Close()
	message := fmt.Sprintf("Written to json file: %v", string(output))
	return message, nil
}
