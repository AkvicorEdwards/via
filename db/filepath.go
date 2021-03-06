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

type ResFilePath struct {
	FilePath
	Filename int64
}

type FilePaths []FilePath

func CalculateFilePath() *ResFilePath {
	if !Connected {
		Connect()
	}
	lockFilePath.Lock()
	defer lockFilePath.Unlock()
	filepath := &FilePath{}
	resFilePath := &ResFilePath{}
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
		resFilePath.Pid = filepath.Pid
		resFilePath.Size = filepath.Size
		resFilePath.Filename = filepath.Size
	} else {
		fis := GetFileInfosByFilepath(filepath.Pid)
		if fis == nil {
			filepath.Size += 1
			err = db.Table(TableFilePath).Where("pid=?", filepath.Pid).Update("size", filepath.Size).Error
			if err != nil {
				log.Println(err)
				return nil
			}
			resFilePath.Pid = filepath.Pid
			resFilePath.Size = filepath.Size
			resFilePath.Filename = filepath.Size
		} else {
			mp := map[int64]bool{}
			for _, v := range *fis {
				mp[v.Filename] = true
			}
			for i := int64(1); i <= int64(def.MaximumFilesPerDirectory); i++ {
				if _, ok := mp[i]; !ok {
					resFilePath.Pid = filepath.Pid
					resFilePath.Size = filepath.Size + 1
					resFilePath.Filename = i
					err = db.Table(TableFilePath).Where("pid=?", filepath.Pid).Update("size", filepath.Size+1).Error
					if err != nil {
						log.Println(err)
						return nil
					}
					break
				}
			}
		}
	}
	return resFilePath
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
