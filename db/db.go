package db

import (
	"gorm.io/gorm/logger"
	"log"
	"sync"
	"time"
	"via/def"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Inc struct {
	Name string // table name
	Val  int64
}

var (
	db        *gorm.DB
	Connected = false
	lockFile = sync.Mutex{}
	lockPath = sync.Mutex{}
	lockRelation = sync.Mutex{}
	lockFilePath = sync.Mutex{}
)

func Connect() {
	if Connected {
		return
	}
	var err error
	db, err = gorm.Open(sqlite.Open(def.DBFilename), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Println(err)
		return
	}
	Connected = true
}

func Exec(sql string, values ...interface{}) *gorm.DB {
	if !Connected {
		Connect()
	}
	return db.Exec(sql, values...)
}

func Raw(sql string, values ...interface{}) *gorm.DB {
	if !Connected {
		Connect()
	}
	return db.Raw(sql, values...)
}

func GetInc(name string) int64 {
	if !Connected {
		Connect()
	}
	ic := Inc{}
	db.Table("inc").Where("name=?", name).First(&ic)
	return ic.Val
}

func UpdateInc(name string, val int64) bool {
	if !Connected {
		Connect()
	}
	return db.Table("inc").Where("name=?", name).Update("val", val).Error == nil
}

func Lock() {
	lockFile.Lock()
	lockPath.Lock()
	lockRelation.Lock()
	lockFilePath.Lock()
}

func UnixTime() int64 {
	return time.Now().Unix()
}

type PathFile interface {
	Permit(byte) bool
	Deny(byte) bool
	GetPassword() string
	//	1 File
	//	2 Path
	Type() int
}
