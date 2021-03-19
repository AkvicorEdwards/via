package def

import (
	"fmt"
	"path"
)

var PORT = 8080
var ADDR = "0.0.0.0"
var SessionDomain = ""
var SessionPath = "/"
var SessionName = "user"

var MaximumFilesPerDirectory = 1024

var RecordEnable = true
var RecordEnableFile = true

var DeleteButton = true
var EditButton = true
var UpdateButton = true

var Username = "Akvicor"
var Password = "password"

var DBFilename = "via.db"
var RecordFilename = "via.log"

var Path = "Via"

func CalPath(p, f int64) string {
	return path.Join(Path, fmt.Sprint(p), fmt.Sprint(f))
}

