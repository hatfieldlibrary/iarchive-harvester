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

func createDirectory(directory string) {
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(directory, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}
}

func createOutputSubDirectory(directory string, subdirectory string) {
	_, err := os.Stat(directory + "/" + subdirectory)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(directory + "/" + subdirectory, 0755)
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
	fmt.Println(f.Name())
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
func readJsonInputFile(input string) ([]types.Thesis, error) {
	var theses []types.Thesis
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
the output, and initializes the data source for later use.
 */
func getIarchiveMetadata(iarchiveID string, oclcNumber string, outputdirectory string) []types.DataSource {
	url := strings.Replace(iarchiveID, "details", "metadata", 1)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body := getResponseBody(resp.Body)
	createOutputSubDirectory(outputdirectory, oclcNumber)
	writeFile(outputdirectory + "/" + oclcNumber, iArchiveOutputFile, body)
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
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	files := dat["files"].([]interface{})
	dataSourceArray := []types.DataSource{}
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
downloadDataSources iterates over data sources, fetches data via http, and writes output.
 */
func downloadDataSources(outputdirectory string, sources []types.DataSource, wskey string) {
	var wg sync.WaitGroup
	for _, each := range sources {
		if each.Source == worldcatType {
			// non-empty worldcat api key required.
			if wskey != "" {
				wg.Add(1)
				url := createWorldCatMetadataUrl(each.OclcNumber, wskey)
				resp, err := getData(url, &wg)
				if err != nil {
					log.Fatal(err)
				}
				body := getResponseBody(resp)
				createOutputSubDirectory(outputdirectory, each.OclcNumber)
				writeFile(outputdirectory+"/"+each.OclcNumber, worldCatOutputFile, body)
			}
		}
		if each.Source == iArchiveType {
			wg.Add(1)
			url := createIArchiveFileUrl(each.BaseUrl, each.File)
			resp, err := getData(url, &wg)
			if err != nil {
				log.Fatal(err)
			}
			body := getResponseBody(resp)
			createOutputSubDirectory(outputdirectory, each.OclcNumber)
			writeFile(outputdirectory + "/" + each.OclcNumber, each.File, body)
		}
	}
	wg.Wait()
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
	for _, each := range theses {
		createDirectory(outputdirectory + "/" + each.Oclc)
		dataSources := getIarchiveMetadata(each.IarchiveID, each.Oclc, outputdirectory)
		downloadDataSources(outputdirectory, dataSources, apiKey)
	}
	message := fmt.Sprintf("Data harvested and written to output directory: %v", string(outputdirectory))
	return message, nil
}