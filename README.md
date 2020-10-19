# IArchives Harvester

A Go project created to harvest metadata and files from Intenet Archive. Accepts tab-delimited file containing IArchive identifier and OCLC number.

If a Worldcat Search WSKey is provided, Worldcat metadata will be included in the harvested record. The WSKey is defined
in a configuration file.  Add this file to the project root directory.

File name: "wskey.json"
Content:

    {
        "comment": "Description of the api key",
        "key": ""
    }

Worldcat records are not harvested if the file is missing, or the key is an empty string.

Writes output to subdirectories. Each directory contains a PDF, a full-text file, and separate files for Internet Archive and Worldcat metadata.

Output also includes an audit file.

## Input File

Takes a tab-delimited file as input.  Required fields are title, Internet Archive ID, and OCLC number.

The Internet Archive title is usually incomplete so this program logs the title provided in the tab-delimited
input. If the title is unavailable, modify the program to log the Internet Archive title instead.

The program retrieves complete metadata via the WorldCat search API. This is optional. To harvest 
WorldCat metadata, provide a WorldCat Search API key as described earlier, and include the OCLC number in 
 your input. 