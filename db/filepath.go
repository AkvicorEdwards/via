package db

import (
	"gorm.io/gorm"
	"log"
	"via/def"
)

const TableFilePath = "filepath"

type FilePath struct {
	Pid        int64
	Size       int64
}

type FilePaths []FilePath

func CalculateFilePath() *FilePath {
	if !Connected {
		Connect()
	}
	lockFilePath.Lock()
	defer lockFilePath.Unlock()
	filepath := &FilePath{}
	err := db.Table(TableFilePath).Where("size<?", def.MaximumFilesPerDirectory).First(filepath).Error
	if err != nil {
		filepath = &FilePath{
			Pid:  GetInc(TableFilePath) + 1,
			Size: 1,
		}
		if err = db.Table(TableFilePath).Create(filepath).Error; err != nil {
			log.Println(err)
			return nil
		}
		UpdateInc(TableFilePath, filepath.Pid)
	} else {
		filepath.Size += 1
		err = db.Table(TableFilePath).Where("pid=?", filepath.Pid).Update("size", filepath.Size).Error
		if err != nil {
			log.Println(err)
			return nil
		}
	}
	return filepath
}

func DecreaseFilePathSize(pid int64) error {
	if !Connected {
		Connect()
	}
	lockFilePath.Lock()
	defer lockFilePath.Unlock()

	return db.Table(TableFilePath).Where("pid=?", pid).UpdateColumns(map[string]interface{}{
		"size":     gorm.Expr("size-1"),
	}).Error
}

func GetFilePaths() *FilePaths {
	if !Connected {
		Connect()
	}
	paths := make(FilePaths, 0)
	err := db.Table(TableFilePath).Find(&paths).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &paths
}
