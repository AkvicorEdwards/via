package repair

import (
	"github.com/AkvicorEdwards/util"
	"log"
	"os"
	"via/db"
	"via/def"
	"via/permission"
)

const (
	sql = `
create table inc
(
    name text,
    val  integer
);
` + `
create table file
(
    fid           integer constraint file_pk primary key,
    title         text,
    comment       text,
    filename      text,
    filepath      integer default 0,
    md5           text,
    sha256        text,
    size          integer default 0,
    password      text,
    permission    integer default 0,
    created       integer default 0,
    modified      integer default 0,
    accessed      integer default 0,
    views         integer default 0,
    priority      integer default 0
);
` + `
create table path
(
    pid           integer constraint path_pk primary key,
    ppid          integer default 0,
    title         text,
    comment       text,
    size          integer default 0,
    password      text,
    permission    integer default 0,
    created       integer default 0,
    modified      integer default 0,
    accessed      integer default 0,
    views         integer default 0
);
` + `
create table filepath
(
    pid           integer constraint filepath_pk primary key,
    size          integer default 0
);
` + `
create table relation
(
    rid           integer constraint relation_pk primary key,
    pid           integer default 0,
    fid           integer default 0
);

` + `
INSERT INTO inc (name, val) VALUES ('file', 0);
INSERT INTO inc (name, val) VALUES ('filepath', 0);
INSERT INTO inc (name, val) VALUES ('path', 0);
INSERT INTO inc (name, val) VALUES ('relation', 0);
`

)

func InitDatabase() {
	stat := util.FileStat(def.DBFilename)
	if stat == 0 || stat == 2 {
		if stat == 2 {
			err := os.Remove(def.DBFilename)
			if err != nil {
				log.Println("Cannot delete file")
				os.Exit(-4)
			}
		}
		f, err := os.Create(def.DBFilename)
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
		log.Printf("%s is a directory\n", def.DBFilename)
		os.Exit(-3)
	}

	err := db.Exec(sql).Error
	if err != nil {
		log.Println(err)
		os.Exit(-5)
	}
	db.AddPath(0, "ROOT", "", "", permission.ParseString("rw------rw"))

	log.Println("Init DB Finished")
}
