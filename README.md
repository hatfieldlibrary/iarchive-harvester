# IArchives Harvester

A Go project created to harvest metadata and files from Intenet Archive. Accepts tab-delimited file containing IArchive identifier and OCLC number.

Worldcat metadata is included in the harvested records if a Worldcat Search API WSKey is provided. The WSKey is defined
in a configuration file that's added to the project root directory.

File name: "wskey.json"

    {
        "comment": "Description of the api key",
        "key": ""
    }

If the file is missing, or the key is an empty string, the harvester does not query the WorldCat Search API.

The harvester writes output to subdirectories. Each directory contains a PDF, a full-text file, and separate files 
for Internet Archive and Worldcat metadata.

Output includes an audit file with JSON entries (deault) and optional tab-delimited log output.

## Input File

Required fields are in the tab-delimited input file are title, Internet Archive ID, and OCLC number.

The Internet Archive title is usually incomplete. This program logs the title provided in the tab-delimited
input. If the title is unavailable then modify the program to log the Internet Archive title instead.

The program can retrieve marcxml metadata via the WorldCat search API. This is optional. To harvest 
WorldCat metadata, provide the OCLC number in your input file and configure a WorldCat Search API key as described 
earlier. 

## Output
For each record, the program creates a subdirectory with 3 files: IArchive metadata, PDF, and full text files. If 
you provided a `wskey.json`, the directory also includes the WorldCat marcxml record.