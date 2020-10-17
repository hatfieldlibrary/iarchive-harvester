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