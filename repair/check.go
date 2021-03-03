package repair

import (
	"github.com/AkvicorEdwards/util"
	"log"
	"os"
	"via/def"
)

func CheckPath() {
	if stat := util.FileStat(def.Path); stat == 0 {
		err := os.MkdirAll(def.Path, 0700)
		if err != nil {
			log.Println(err)
			os.Exit(-1)
		}
	} else if stat == 2 {
		log.Println("Path is file")
		os.Exit(-1)
	}
}


