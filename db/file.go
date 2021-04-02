package db

import (
	"encoding/json"
	"fmt"
	"github.com/AkvicorEdwards/util"
	"gorm.io/gorm"
	"log"
	"os"
	"path"
	"sort"
	"via/def"
)

const TableFile = "file"

type File struct {
	Fid        int64  `json:"vid"`
	Title      string `json:"title"`
	Comment    string `json:"comment"`
	Filename   int64
	Filepath   int64
	MD5        string `json:"md5"`
	SHA256     string `json:"sha256"`
	Size       int64  `json:"size"`
	Password   string
	Permission int64 `json:"permission"`
	Created    int64 `json:"created"`
	Modified   int64 `json:"modified"`
	Accessed   int64 `json:"accessed"`
	Views      int64 `json:"views"`
	Priority   int64 `json:"priority"`
}

func (f *File) Deny(p byte) bool {
	return (f.Permission>>p)&1 == 0
}

func (f *File) Permit(p byte) bool {
	return (f.Permission>>p)&1 == 1
}

func (f *File) GetPassword() string {
	return f.Password
}

func (f *File) Type() int {
	return 1
}

type Files []File

func (f *File) JSON() string {
	res, err := json.Marshal(f)
	if err != nil {
		return ""
	}
	return string(res)
}

func (f *Files) JSON() string {
	res, err := json.Marshal(f)
	if err != nil {
		return ""
	}
	return string(res)
}

// 1: Fid DESC
// 2: Size DESC
// 3: Created DESC
// 4: Modified DESC
// 5: Accessed DESC
// 6: Views ASC
// 7: Priority DESC
func (f *Files) Sort(by int, order bool) {
	s := func(less func(i, j int) bool) { sort.Slice((*f)[:], less) }
	switch by {
	case 1:
		if order {
			s(func(i, j int) bool { return (*f)[i].Fid > (*f)[j].Fid })
		} else {
			s(func(i, j int) bool { return (*f)[i].Fid < (*f)[j].Fid })
		}
	case 2:
		if order {
			s(func(i, j int) bool { return (*f)[i].Size > (*f)[j].Size })
		} else {
			s(func(i, j int) bool { return (*f)[i].Size < (*f)[j].Size })
		}
	case 3:
		if order {
			s(func(i, j int) bool { return (*f)[i].Created > (*f)[j].Created })
		} else {
			s(func(i, j int) bool { return (*f)[i].Created < (*f)[j].Created })
		}
	case 4:
		if order {
			s(func(i, j int) bool { return (*f)[i].Modified > (*f)[j].Modified })
		} else {
			s(func(i, j int) bool { return (*f)[i].Modified < (*f)[j].Modified })
		}
	case 5:
		if order {
			s(func(i, j int) bool { return (*f)[i].Accessed > (*f)[j].Accessed })
		} else {
			s(func(i, j int) bool { return (*f)[i].Accessed < (*f)[j].Accessed })
		}
	case 6:
		if order {
			s(func(i, j int) bool { return (*f)[i].Views < (*f)[j].Views })
		} else {
			s(func(i, j int) bool { return (*f)[i].Views > (*f)[j].Views })
		}
	case 7:
		if order {
			s(func(i, j int) bool { return (*f)[i].Priority > (*f)[j].Priority })
		} else {
			s(func(i, j int) bool { return (*f)[i].Priority < (*f)[j].Priority })
		}
	}
}

func AddFile(title, comment, md5, sha256, password string, filename, filepath, size, permission, priority int64) int64 {
	if !Connected {
		Connect()
	}
	lockFile.Lock()
	defer lockFile.Unlock()

	file := File{
		Fid:        GetInc(TableFile) + 1,
		Title:      title,
		Comment:    comment,
		Filename:   filename,
		Filepath:   filepath,
		MD5:        md5,
		SHA256:     sha256,
		Size:       size,
		Password:   password,
		Permission: permission,
		Created:    UnixTime(),
		Modified:   0,
		Accessed:   0,
		Views:      0,
		Priority:   priority,
	}

	if err := db.Table(TableFile).Create(&file).Error; err != nil {
		return -1
	}

	UpdateInc(TableFile, file.Fid)

	return file.Fid
}

func UpdateFileInfo(fi *File) bool {
	if !Connected {
		Connect()
	}
	lockFile.Lock()
	defer lockFile.Unlock()

	res := db.Table(TableFile).Where("fid=?", fi.Fid).UpdateColumns(map[string]interface{}{
		"title":      fi.Title,
		"comment":    fi.Comment,
		"md5":        fi.MD5,
		"sha256":     fi.SHA256,
		"size":       fi.Size,
		"password":   fi.Password,
		"permission": fi.Permission,
		"priority":   fi.Priority,
	})
	if res.Error != nil {
		return false
	}
	return true
}

// TODO
func CompleteDeleteFile(fid int64) bool {
	return false
}

func DeleteFileFromAllPath(fid int64) bool {
	return false
}

func DeleteFileFromPath(fid, pid int64) bool {
	return false
}

func EraseAndUpdate(fid int64) bool {
	return false
}

func Update(fid, filepath int64) bool {
	return false
}

func CompleteDelFile(fid int64) bool {
	if !Connected {
		Connect()
	}
	lockFile.Lock()
	defer lockFile.Unlock()
	file := &File{}
	err := db.Table(TableFile).Where("fid=?", fid).First(file).Error
	if err != nil {
		log.Println(err)
		return false
	}
	DelRelationByFile(fid)
	err = os.Remove(path.Join(def.Path, fmt.Sprint(file.Filepath), fmt.Sprint(file.Filename)))
	if err != nil {
		log.Println(err)
	}
	err = DecreaseFilePathSize(file.Filepath)
	if err != nil {
		log.Println(err)
	}
	err = db.Table(TableFile).Where("fid=?", fid).Delete(file).Error
	if err != nil {
		log.Println(err)
	}

	return true
}

func GetFileInfo(fid int64) *File {
	if !Connected {
		Connect()
	}
	file := &File{}
	err := db.Table(TableFile).Where("fid=?", fid).First(file).Error
	if err != nil {
		return nil
	}

	return file
}

func GetFileInfos() *Files {
	if !Connected {
		Connect()
	}
	files := make(Files, 0)
	err := db.Table(TableFile).Find(&files).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &files
}

func GetFileInfosNotInFids(fids []int64) *Files {
	if !Connected {
		Connect()
	}
	files := make(Files, 0)
	err := db.Table(TableFile).Where("fid NOT IN (?)", fids).Find(&files).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &files
}

func GetFileInfosByPath(pid int64) *Files {
	if !Connected {
		Connect()
	}

	files := make(Files, 0)
	err := db.Table(TableFile).Where("fid IN (?)", GetRelationsByPath(pid).Fid()).Find(&files).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &files
}

func GetFileInfosByFilepath(filepath int64) *Files {
	if !Connected {
		Connect()
	}

	files := make(Files, 0)
	err := db.Table(TableFile).Where("filepath = ?", filepath).Find(&files).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &files
}

func FileAccessed(fid int64) {
	if !Connected {
		Connect()
	}
	lockFile.Lock()
	defer lockFile.Unlock()
	res := db.Table(TableFile).Where("fid=?", fid).UpdateColumns(map[string]interface{}{
		"accessed": UnixTime(),
		"views":    gorm.Expr("views+1"),
	})
	if res.Error != nil {
		log.Println(res.Error)
	}
}

func FileSearchBySHA256(s6 string) *Files {
	if !Connected {
		Connect()
	}
	files := make(Files, 0)
	err := db.Table(TableFile).Where("sha256 = ?", s6).Find(&files).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &files
}

func FileSearchByMD5(m5 string) *Files {
	if !Connected {
		Connect()
	}
	files := make(Files, 0)
	err := db.Table(TableFile).Where("md5 = ?", m5).Find(&files).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &files
}

func FileSearchByTitle(title string) *Files {
	if !Connected {
		Connect()
	}
	files := make(Files, 0)
	err := db.Table(TableFile).Where("title LIKE ?", fmt.Sprintf("%%%s%%", title)).Find(&files).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &files
}

func FileExist(m5, s6 string) *Files {
	if !Connected {
		Connect()
	}
	files := make(Files, 0)
	err := db.Table(TableFile).Where("md5 = ? AND sha256 = ?", m5, s6).Find(&files).Error
	if err != nil {
		return nil
	}
	if len(files) == 0 {
		return nil
	}
	return &files
}

func VerifyMD5(fid int64) bool {
	if !Connected {
		Connect()
	}
	fi := GetFileInfo(fid)
	if fi == nil {
		return false
	}
	md5, err := util.MD5File(def.CalPath(fi.Filepath, fi.Filename))
	if err != nil {
		return false
	}
	return fi.MD5 == fmt.Sprintf("%X", md5)
}

func VerifySHA256(fid int64) bool {
	if !Connected {
		Connect()
	}
	fi := GetFileInfo(fid)
	if fi == nil {
		return false
	}
	sha256, err := util.SHA256File(def.CalPath(fi.Filepath, fi.Filename))
	if err != nil {
		return false
	}
	return fi.SHA256 == fmt.Sprintf("%X", sha256)
}
