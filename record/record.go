package record

import (
	"fmt"
	"github.com/AkvicorEdwards/util"
	"log"
	"os"
	"sync"
	"via/def"
)

var file *os.File = nil
var Lock = &sync.Mutex{}

func Enable() bool {
	if !def.RecordEnableFile {
		return false
	}
	CheckRecordFile()
	var err error
	file, err = os.OpenFile(def.RecordFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0600)
	if err != nil {
		return false
	}
	return true
}

func CloseFile() {
	if file != nil {
		_ = file.Close()
	}
}

func CheckRecordFile() {
	stat := util.FileStat(def.RecordFilename)
	if stat == 0 {
		if stat == 2 {
			err := os.Remove(def.RecordFilename)
			if err != nil {
				log.Println("Cannot delete file")
				os.Exit(-4)
			}
		}
		f, err := os.Create(def.RecordFilename)
		if err != nil {
			log.Println("Cannot create file")
			os.Exit(-1)
		}
		err = f.Close()
		if err != nil {
			log.Println("Create failed")
			os.Exit(-2)
		}
	} else if stat == 1 {
		log.Printf("%s is a directory\n", def.RecordFilename)
		os.Exit(-3)
	}
}

func Printf(format string, v ...interface{}) {
	if def.RecordEnable {
		log.Printf(format, v...)
	}
	if def.RecordEnableFile {
		r := []byte(AddTimeNow(fmt.Sprintf(format, v...)))
		Lock.Lock()
		defer Lock.Unlock()
		n, err := file.Write(r)
		if err != nil || n != len(r) {
			Enable()
		}
	}
}

func Println(v ...interface{}) {
	if def.RecordEnable {
		log.Println(v...)
	}
	if def.RecordEnableFile {
		r := []byte(AddTimeNow(fmt.Sprintln(v...)))
		Lock.Lock()
		defer Lock.Unlock()
		n, err := file.Write(r)
		if err != nil || n != len(r) {
			Enable()
		}
	}
}

func Print(v ...interface{}) {
	if def.RecordEnable {
		log.Print(v...)
	}
	if def.RecordEnableFile {
		r := []byte(AddTimeNow(fmt.Sprint(v...)))
		Lock.Lock()
		defer Lock.Unlock()
		n, err := file.Write(r)
		if err != nil || n != len(r) {
			Enable()
		}
	}
}
