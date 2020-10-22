package harvest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"theses/filereader"
	"theses/types"
)

type harvest struct {
	createCSV          bool
	inputCsv           string
	outputDirectory    string
	apiKey             string
	auditFile 		   *os.File
}

type dataDownload struct {
	file string
	response io.ReadCloser
	error error
}

const iArchiveOutputFileName = "iarchive.json"
const worldCatOutputFileName = "worldcat.xml"
const iArchiveType = "iArchiveFile"
const worldcatType = "worldcat"
const auditFileLocation = "../audit.log"
const jsonInputFile = "../records.json"

/*
createDirectory makes an output directory.
 */
func createDirectory(directory string) {
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(directory, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}
}

/*
writeFile writes data to the supplied directory and file name.
 */
func writeFile(directory string, fileName string, data []byte) {
	f, err := os.Create(directory + "/" + fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err2 := f.WriteString(string(data))
	if err2 != nil {
		log.Fatal(err2)
	}
}

/*
createIArchiveFileUrl creates a metadata file download URL from the provided Internet Archive ID.
 */
func createIArchiveFileUrl(iarchiveID string, name string) string {
	url := strings.Replace(iarchiveID, "details", "download", 1)
	url += "/" + name
	return url
}

/*
createWorldCatMetadataUrl creates a WorldCat Search API query with the provide accession number and API key.
 */
func createWorldCatMetadataUrl(accession string, wskey string) string {
	return "http://www.worldcat.org/webservices/catalog/content/" + accession + "?wskey=" + wskey
}

/*
readJsonInputFile reads the json input file. This file is created from the csv input file.
 */
func readJsonInputFile(input string) ([]types.Record, error) {
	var theses []types.Record
	dat, err := ioutil.ReadFile(input)
	if err != nil {
		log.Fatal(err)
	}
	_ = json.Unmarshal([]byte(dat), &theses)
	return theses, nil
}

/*
getResponseBody reads the HTTP response body.
 */
func getResponseBody(response io.ReadCloser) []byte {
	body, err := ioutil.ReadAll(response)
	if err != nil {
		log.Fatal(err)
	}
	response.Close()
	return body
}

/*
getIarchiveMetadata fetches IArchive metadata for single item and returns the response body.
 */
func getIarchiveMetadata(iarchiveID string) []byte {
	url := strings.Replace(iarchiveID, "details", "metadata", 1)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	body := getResponseBody(resp.Body)
	return body
}

/*
readIAJsonResponse converts the response buffer to a map.
 */
func readIAJsonResponse(iarchiveID string, body []byte) map[string]interface{} {
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Printf("warning: unable to retrieve metadata for: %v\n", iarchiveID)
		return nil
	}
	return dat
}

/*
getDataSources creates an array of DataSource types from Internet Archive file metadata. Appends
an OCLC DataSource to the list.
*/
func getDataSources(files []interface{}, iarchiveID string, oclcNumber string) []types.DataSource {
	var dataSourceArray []types.DataSource
	for i := 0; i < len(files); i++ {
		file := files[i].(map[string]interface{})
		// See IArchive metadata record for available file types.
		if file["format"] == "Text PDF" {
			source := types.DataSource{File: file["name"].(string), OclcNumber: oclcNumber, Source: iArchiveType,
				BaseUrl: iarchiveID}
			dataSourceArray = append(dataSourceArray, source)
		}
		if file["format"] == "DjVuTXT" {
			source := types.DataSource{File: file["name"].(string), OclcNumber: oclcNumber, Source: iArchiveType,
				BaseUrl: iarchiveID}
			dataSourceArray = append(dataSourceArray, source)
		}
	}
	oclcSource := types.DataSource{File: oclcNumber, OclcNumber: oclcNumber, Source: worldcatType}
	dataSourceArray = append(dataSourceArray, oclcSource)
	return dataSourceArray
}

/*
getData executes a single HTTP GET request.
 */
func getData(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

/*
downloadDataSources iterates over data sources, fetches via http, and writes to output files. A wskey is required
for WorldCat Search API queries.
 */
func downloadDataSources(sources []types.DataSource, wskey string, outputdirectory string, ) {
	ch := make(chan *dataDownload, len(sources))
	for _, each := range sources {
		go func(each types.DataSource) {
			if each.Source == worldcatType {
				// non-empty worldcat api key required.
				if wskey != "" && len(each.OclcNumber) > 4 {
					url := createWorldCatMetadataUrl(each.OclcNumber, wskey)
					resp, err := getData(url)
					if err != nil {
						log.Println(err)
					}
					ch <- &dataDownload{worldCatOutputFileName, resp, err}
				}
			}
			if each.Source == iArchiveType {
				url := createIArchiveFileUrl(each.BaseUrl, each.File)
				resp, err := getData(url)
				if err != nil {
					log.Println(err)
				}
				ch <- &dataDownload{each.File, resp, err}
			}
		}(each)

	}
	// Read from channels.
	for range sources {
		data := <-ch
		body := getResponseBody(data.response)
		writeFile(outputdirectory, data.file, body)
	}
}

/*
updateAudit writes an new entry to the json auditFile
 */
func updateAudit(harvester harvest, title string, iarchiveID string, oclcNumber string, outputDirectory string,
	metadata map[string]interface{}) {
	creator := metadata["creator"].(string)
	description := metadata["description"].(string)
	date := metadata["date"].(string)
	auditEntry := types.Audit{Title: title, Author: creator, Date: date, Description: description,
		OCLCNumber: oclcNumber, IArchiveID: iarchiveID, OutputDirectory: outputDirectory}
	jsondata, err := json.Marshal(auditEntry) // convert to JSON
	if err != nil {
		fmt.Println("error marshalling audit record")
		fmt.Println(err)
		os.Exit(1)
	}
	_, err2 := harvester.auditFile.WriteString(string(jsondata) + "\n")
	if err2 != nil {
		log.Printf("failed writing to audit file.\n")
		err3 := harvester.auditFile.Close()
		if err3 != nil {
			log.Fatal(err3)
		}
		log.Fatal(err2)
	}
}

/*
Converts the csv input file to Json.
 */
func convertToJsonFile(csvFile string, jsonFile string) {
	_, err := filereader.InputFileConverter(csvFile, jsonFile)
	if (err != nil) {
		log.Fatal(err)
	}
}

/*
Removes current log files if the exist.
 */
func removeAuditFileIfExists(harvester harvest) {
	var _, err = os.Stat(auditFileLocation)
	if err == nil {
		err1 := os.Remove(auditFileLocation)
		if err1 != nil {
			message := fmt.Sprintf("Unable to remove log file: %s", auditFileLocation)
			fmt.Println(message)
		}
	}
	if harvester.createCSV {
		csvFileName := filereader.CreateCsvLogFileName(auditFileLocation)
		var _, err2 = os.Stat(csvFileName)
		if err2 == nil {
			err3 := os.Remove(csvFileName)
			if err3 != nil {
				message := fmt.Sprintf("Unable to remove log file: %s", csvFileName)
				fmt.Println(message)
			}
		}

	}
}

/*
Creates the output log file and adds the file pointer to the harvester struct.
 */
func createAuditFile(harvester harvest) *os.File {
	removeAuditFileIfExists(harvester)
	auditFile, err := os.OpenFile(auditFileLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0775)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	return auditFile
}

/*
Harvests data for each record Record type and writes to output directories. Returns the final count
of records processed.
 */
func processRecords(harvester harvest, records [] types.Record) string {
	count := 1
	for _, each := range records {
		var subdir = fmt.Sprintf("%05d", count)
		recordOutputDir := harvester.outputDirectory + "/" + subdir
		createDirectory(recordOutputDir)
		body := getIarchiveMetadata(each.IarchiveID)
		writeFile(recordOutputDir, iArchiveOutputFileName, body)
		iaResponse := readIAJsonResponse(each.IarchiveID, body)
		if iaResponse != nil {
			metadata := iaResponse["metadata"].(map[string]interface{})
			files := iaResponse["files"].([]interface{})
			updateAudit(harvester, each.Title, each.IarchiveID, each.Oclc, subdir, metadata)
			dataSources := getDataSources(files, each.IarchiveID, each.Oclc)
			downloadDataSources(dataSources, harvester.apiKey, recordOutputDir)
		}
		count++
	}
	message := fmt.Sprintf("%d records harvested and written to output directory: %s",
		count - 1, harvester.outputDirectory)
	return message
}

/*
FetchData retrieves metadata and binary files from the Internet Archive and marcxml metadata
from WorldCat. The input is a tab-delimited text file.  Output is written to subdirectories containing a json,
marcxml, and binary files.
*/
func FetchData(harvester harvest) (string, error) {
	if harvester.outputDirectory == "" {
		return "", errors.New("no output file name")
	}
	auditFile := createAuditFile(harvester)
	defer auditFile.Close()
	harvester.auditFile = auditFile
	convertToJsonFile(harvester.inputCsv, jsonInputFile)
	createDirectory(harvester.outputDirectory)
	records, _ := readJsonInputFile(jsonInputFile)
	message := processRecords(harvester, records)
	if harvester.createCSV {
		filereader.ConvertLogToCsv(auditFileLocation)
	}
	return message, nil
}

/*
Initializes and returns harvester struct.
 */
func New(inputCsvFile string, outputDirectory string, apiKey string, createCsv bool) harvest {
	h := harvest{inputCsv: inputCsvFile, outputDirectory: outputDirectory, apiKey: apiKey, createCSV: createCsv}
	return h
}