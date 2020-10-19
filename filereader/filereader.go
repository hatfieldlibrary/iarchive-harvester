package filereader

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"theses/types"
)

/*
Columns in tab-delimited file.
 */
const titleColumn = 0
const iarchiveColumn = 9
const oclcColumn = 24

/*
ReadApiKey returns the key set in the json configuration file. This file is described in README.
 */
func ReadApiKey(input string) (string, error) {
	if (input == "") {
		return "", errors.New("no api configuration file name, harvesting Internet Archive records only")
	}
	dat, err := ioutil.ReadFile(input)
	if (err != nil) {
		return "", errors.New(fmt.Sprintf("unable to open api key file %s, harvesting Internet Archive records " +
			"only", input))
	}
	key := types.ApiKey{}
	_ = json.Unmarshal([]byte(dat), &key)
	return key.Key, nil
}

/*
writeCsvLine creates a tab-delimited line from the json entry and writes that line to the provided file handle.
 */
func writeCsvLine(line []byte, csvFile *os.File) {
	jsonFormat := types.Audit{}
	_ = json.Unmarshal([]byte(line), &jsonFormat)
	title := jsonFormat.Title
	author := jsonFormat.Author
	date := jsonFormat.Date
	description := jsonFormat.Description
	iarchiveID := jsonFormat.IArchiveID
	oclcNumber  := jsonFormat.OCLCNumber
	outputdir := jsonFormat.OutputDirectory
	fields := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", title, author, date, description, iarchiveID,
		oclcNumber, outputdir)
	_, err := csvFile.WriteString(fields + "\n")
	if err != nil {
		log.Fatal(err)
	}
}
/*
Each line in the log file is a json object. For some projects it is useful to have the same information
 in tab-delimited format. Calling ConvertLogToCsv creates a csv file.
 */
func ConvertLogToCsv(logFile string) {
	file, err := os.Open(logFile)
	if err != nil {
		log.Fatal(err)
	}
	csvFileName := strings.Replace(logFile, "log", "csv", 1)
	csvFile, err1 := os.OpenFile(csvFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0775)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer csvFile.Close()
	defer file.Close()
	scanner := bufio.NewScanner(file)
	// set generous buffer capacity
	const maxCapacity = 10*1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	for scanner.Scan() {
		writeCsvLine(scanner.Bytes(), csvFile)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
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
	var oneRecord types.Record
	var allRecords []types.Record
	for _, each := range csvData {
		oneRecord.Title = each[titleColumn]
		oneRecord.IarchiveID = each[iarchiveColumn]
		oneRecord.Oclc = each[oclcColumn]
		allRecords = append(allRecords, oneRecord)
	}
	start := fmt.Sprintf("Processing %v records.", len(allRecords))
	fmt.Println(start)
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
