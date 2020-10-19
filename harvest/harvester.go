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
	"sync"
	"theses/filereader"
	"theses/types"
)

// if createCSV is true, a tab-delimited version of the log file is created in addition to json
const createCSV = true
const iArchiveOutputFile = "iarchive.json"
const worldCatOutputFile = "worldcat.xml"
const iArchiveType = "iArchiveFile"
const worldcatType = "worldcat"
const auditFileLocation = "../audit.log"

func createDirectory(directory string) {
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(directory, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}
}

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

func createIArchiveFileUrl(iarchiveID string, name string) string {
	url := strings.Replace(iarchiveID, "details", "download", 1)
	url += "/" + name
	return url
}

func createWorldCatMetadataUrl(accession string, wskey string) string {
	return "http://www.worldcat.org/webservices/catalog/content/" + accession + "?wskey=" + wskey
}

/*
readJsonInputFile reads the json input file provided as a input parameter to the exported function.
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
getResponseBody reads the response body.
 */
func getResponseBody(response io.ReadCloser) []byte {
	body, err := ioutil.ReadAll(response)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

/*
getIarchiveMetadata fetches IArchive metadata for single item via GET request, writes the response body to
the output, and initializes the data source for later use. Logs http errors.
 */
func getIarchiveMetadata(title, iarchiveID string, oclcNumber string, outputdirectory string, auditFile *os.File,
	directory string) []types.DataSource {
	url := strings.Replace(iarchiveID, "details", "metadata", 1)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	body := getResponseBody(resp.Body)
	writeFile(outputdirectory, iArchiveOutputFile, body)
	var dataSources = setDataSources(title, oclcNumber, iarchiveID, body, auditFile, directory)
	return dataSources
}

/*
setDataSources appends information for IArchive and OCLC data sources to data sources array.
 */
func setDataSources(title, oclcNumber string, iarchiveID string, body []byte, auditFile *os.File,
	directory string) []types.DataSource {
	iArchiveSourceFiles := readIAJsonResponse(title, iarchiveID, oclcNumber, body, auditFile, directory)
	oclcSource := types.DataSource{File: oclcNumber, OclcNumber: oclcNumber, Source: worldcatType}
	iArchiveSourceFiles = append(iArchiveSourceFiles, oclcSource)
	return iArchiveSourceFiles
}

/*
readIAJsonResponse adds file information extracted from the IArchive json response to the data sources array.
 */
func readIAJsonResponse(title, iarchiveID string, oclcNumber string, body []byte, auditFile *os.File,
	directory string) []types.DataSource {
	dataSourceArray := []types.DataSource{}
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Printf("warning: unable to retrieve metadata for: %v\n", iarchiveID)
		return dataSourceArray
	}
	metadata := dat["metadata"].(map[string]interface{})
	creator := metadata["creator"].(string)
	description := metadata["description"].(string)
	date := metadata["date"].(string)
	auditEntry := types.Audit{Title: title, Author: creator, Date: date, Description: description,
		OCLCNumber: oclcNumber, IArchiveID: iarchiveID, OutputDirectory: directory}
	updateAudit(auditFile, &auditEntry)
	files := dat["files"].([]interface{})
	for i := 0; i < len(files); i++ {
		file := files[i].(map[string]interface{})
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
	return dataSourceArray
}


/*
getData executes a single http GET request.
 */
func getData(url string, wg *sync.WaitGroup) (io.ReadCloser, error) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
/*
downloadDataSources iterates over data sources, fetches data via http, and writes output. Logs http errors.
 */
func downloadDataSources(outputdirectory string, sources []types.DataSource, wskey string) {
	var wg sync.WaitGroup
	for _, each := range sources {
		if each.Source == worldcatType {
			// non-empty worldcat api key required.
			if wskey != "" && len(each.OclcNumber) > 4 {
				wg.Add(1)
				url := createWorldCatMetadataUrl(each.OclcNumber, wskey)
				resp, err := getData(url, &wg)
				if err != nil {
					log.Println(err)
				}
				body := getResponseBody(resp)
				writeFile(outputdirectory, worldCatOutputFile, body)
			}
		}
		if each.Source == iArchiveType {
			wg.Add(1)
			url := createIArchiveFileUrl(each.BaseUrl, each.File)
			resp, err := getData(url, &wg)
			if err != nil {
				log.Println(err)
			}
			body := getResponseBody(resp)
			writeFile(outputdirectory, each.File, body)
		}
	}
	wg.Wait()
}

/*
updateAudit writes an new entry to the json auditFile
 */
func updateAudit(file *os.File, entry *types.Audit) {
	jsondata, err := json.Marshal(entry) // convert to JSON
	if err != nil {
		fmt.Println("error marshalling audit record")
		fmt.Println(err)
		os.Exit(1)
	}
	_, err2 := file.WriteString(string(jsondata) + "\n")
	if err2 != nil {
		log.Printf("failed writing to audint file: %s\n", err)
	}
}

/*
HarvestData exported function retrieves metadata and binary files from the Internet Archive and additional metadata
from Worldcat. Input data is a tab-delimited text file.  Output is written to subdirectories containing a json file
for IArchive data, a marcxml file for worldcat data, and binary files.
 */
func HarvestData(input string, outputdirectory string, apiKey string) (string, error) {
	if input == "" {
		return "", errors.New("no input file name")
	}
	if outputdirectory == "" {
		return "", errors.New("no output file name")
	}
	createDirectory(outputdirectory)
	records, _ := readJsonInputFile(input)
	count := 1
	// create the audit.log file.
	auditFile, err := os.OpenFile(auditFileLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0775)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer auditFile.Close()
	for _, each := range records {
		title := each.Title
		var subdir = fmt.Sprintf("%05d", count)
		createDirectory(outputdirectory + "/" + subdir)
		dataSources := getIarchiveMetadata(title, each.IarchiveID, each.Oclc, outputdirectory + "/" + subdir,
			auditFile, subdir)
		downloadDataSources(outputdirectory + "/" + subdir, dataSources, apiKey)
		count++

	}
	if (createCSV) {
		filereader.ConvertLogToCsv(auditFileLocation)
		if (err != nil) {
			log.Fatal(err)
		}
	}
	message := fmt.Sprintf("Data harvested and written to output directory: %v", string(outputdirectory))
	return message, nil
}