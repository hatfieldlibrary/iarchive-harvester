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
	"theses/types"
)

const iArchiveOutputFile = "iarchive.json"
const worldCatOutputFile = "worldcat.xml"
const iArchiveType = "iArchiveFile"
const worldcatType = "worldcat"
const auditFile = "../audit.json"

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
func getIarchiveMetadata(iarchiveID string, oclcNumber string, outputdirectory string) []types.DataSource {
	url := strings.Replace(iarchiveID, "details", "metadata", 1)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	body := getResponseBody(resp.Body)
	writeFile(outputdirectory, iArchiveOutputFile, body)
	var dataSources = setDataSources(oclcNumber, iarchiveID, body)
	return dataSources
}

/*
setDataSources appends information for IArchive and OCLC data sources to data sources array.
 */
func setDataSources(oclcNumber string, iarchiveID string, body []byte) []types.DataSource {
	iArchiveSourceFiles := readIAJsonResponse(iarchiveID, oclcNumber, body)
	oclcSource := types.DataSource{File: oclcNumber, OclcNumber: oclcNumber, Source: worldcatType}
	iArchiveSourceFiles = append(iArchiveSourceFiles, oclcSource)
	return iArchiveSourceFiles
}

/*
readIAJsonResponse adds file information extracted from the IArchive json response to the data sources array.
 */
func readIAJsonResponse(iarchiveID string, oclcNumber string, body []byte) []types.DataSource {
	dataSourceArray := []types.DataSource{}
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Printf("warning: unable to retrieve metadata for: %v\n", iarchiveID)
		return dataSourceArray
	}
	files := dat["files"].([]interface{})
	for i := 0; i < len(files); i++ {
		file := files[i].(map[string]interface{})
		if file["format"] == "Text PDF" {
			source := types.DataSource{File: file["name"].(string), OclcNumber: oclcNumber, Source: iArchiveType, BaseUrl: iarchiveID}
			dataSourceArray = append(dataSourceArray, source)
		}
		if file["format"] == "DjVuTXT" {
			source := types.DataSource{File: file["name"].(string), OclcNumber: oclcNumber, Source: iArchiveType, BaseUrl: iarchiveID}
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
			if wskey != "" {
				fmt.Println(each.OclcNumber)
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
HarvestData exported function retrieves metadata and binary files from the Internet Archive and additional metadata from Worldcat.
Input data is a tab-delimited text file.  Output is written to subdirectories containing a json file for IArchive data,
a marcxml file for worldcat data, and binary files.
 */
func HarvestData(input string, outputdirectory string, apiKey string) (string, error) {
	if input == "" {
		return "", errors.New("no input file name")
	}
	if outputdirectory == "" {
		return "", errors.New("no output file name")
	}
	createDirectory(outputdirectory)
	theses, _ := readJsonInputFile(input)
	count := 1
	file, err := os.OpenFile(auditFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0775)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()
	for _, each := range theses {
		oclc := each.Oclc
		ia := each.IarchiveID
		title := each.Title
		auditEntry := types.Audit{Title: title, OCLCNumber: oclc, IArchiveID: ia}
		updateAudit(file, &auditEntry)
		var subdir = fmt.Sprintf("%05d", count)
		createDirectory(outputdirectory + "/" + subdir)
		dataSources := getIarchiveMetadata(each.IarchiveID, each.Oclc, outputdirectory + "/" + subdir)
		downloadDataSources(outputdirectory + "/" + subdir, dataSources, apiKey)
		count++

	}
	message := fmt.Sprintf("Data harvested and written to output directory: %v", string(outputdirectory))
	return message, nil
}