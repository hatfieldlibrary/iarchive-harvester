# IArchives Harvester

A Go project created to harvest theses from Intenet Archives.

If a Worldcat Search WSKey is provided, Worldcat metadata will be included in the result. The WSKey can be defined
in a configuration file.  Add this file to the root project directory.

File name: "wskey.json"
Content:

    {
        "comment": "Description of the api key",
        "key": ""
    }

Worldcat records are not harvested if the file is missing, or the key is an empty string.