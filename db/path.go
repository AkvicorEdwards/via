package db

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"log"
	"sort"
)

const TablePath = "path"

type Path struct {
	Pid        int64  `json:"pid"`
	Ppid       int64  `json:"ppid"` // Parent path id
	Title      string `json:"title"`
	Comment    string `json:"comment"`
	Size       int64  `json:"size"`
	Password   string
	Permission int64 `json:"permission"`
	Created    int64 `json:"created"`
	Modified   int64 `json:"modified"`
	Accessed   int64 `json:"accessed"`
	Views      int64 `json:"views"`
}

func (p *Path) Deny(per byte) bool {
	return (p.Permission>>per)&1 == 0
}

func (p *Path) Permit(per byte) bool {
	return (p.Permission>>per)&1 == 1
}

func (p *Path) GetPassword() string {
	return p.Password
}

func (p *Path) Type() int {
	return 2
}

type Paths []Path

func (p *Path) JSON() string {
	res, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(res)
}

func (v *Paths) JSON() string {
	res, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(res)
}

// 1: Pid DESC
// 2: Size DESC
// 3: Created DESC
// 4: Modified DESC
// 5: Accessed DESC
// 6: Views ASC
func (v *Paths) Sort(by int, order bool) {
	s := func(less func(i, j int) bool) { sort.Slice((*v)[:], less) }
	switch by {
	case 1:
		if order {
			s(func(i, j int) bool { return (*v)[i].Pid > (*v)[j].Pid })
		} else {
			s(func(i, j int) bool { return (*v)[i].Pid < (*v)[j].Pid })
		}
	case 2:
		if order {
			s(func(i, j int) bool { return (*v)[i].Size > (*v)[j].Size })
		} else {
			s(func(i, j int) bool { return (*v)[i].Size < (*v)[j].Size })
		}
	case 3:
		if order {
			s(func(i, j int) bool { return (*v)[i].Created > (*v)[j].Created })
		} else {
			s(func(i, j int) bool { return (*v)[i].Created < (*v)[j].Created })
		}
	case 4:
		if order {
			s(func(i, j int) bool { return (*v)[i].Modified > (*v)[j].Modified })
		} else {
			s(func(i, j int) bool { return (*v)[i].Modified < (*v)[j].Modified })
		}
	case 5:
		if order {
			s(func(i, j int) bool { return (*v)[i].Accessed > (*v)[j].Accessed })
		} else {
			s(func(i, j int) bool { return (*v)[i].Accessed < (*v)[j].Accessed })
		}
	case 6:
		if order {
			s(func(i, j int) bool { return (*v)[i].Views < (*v)[j].Views })
		} else {
			s(func(i, j int) bool { return (*v)[i].Views > (*v)[j].Views })
		}
	}
}

func AddPath(ppid int64, title, comment, password string, permission int64) bool {
	if !Connected {
		Connect()
	}
	lockPath.Lock()
	defer lockPath.Unlock()

	path := Path{
		Pid:        GetInc(TablePath) + 1,
		Ppid:       ppid,
		Title:      title,
		Comment:    comment,
		Size:       0,
		Password:   password,
		Permission: permission,
		Created:    UnixTime(),
		Modified:   0,
		Accessed:   0,
		Views:      0,
	}

	if err := db.Table(TablePath).Create(&path).Error; err != nil {
		return false
	}

	UpdateInc(TablePath, path.Pid)

	res := db.Table(TablePath).Where("pid=?", ppid).UpdateColumns(map[string]interface{}{
		"size": gorm.Expr("size+1"),
	})
	if res.Error != nil {
		log.Println(res.Error)
	}

	return true
}

func UpdatePathInfo(ph *Path) bool {
	if !Connected {
		Connect()
	}
	lockPath.Lock()
	defer lockPath.Unlock()

	res := db.Table(TablePath).Where("pid=?", ph.Pid).UpdateColumns(map[string]interface{}{
		"ppid":       ph.Ppid,
		"title":      ph.Title,
		"comment":    ph.Comment,
		"size":       ph.Size,
		"password":   ph.Password,
		"permission": ph.Permission,
	})
	if res.Error != nil {
		return false
	}
	return true
}

func DelPath(pid int64) bool {
	if !Connected {
		Connect()
	}
	// Delete Child Path
	paths := GetPathInfosByParent(pid)
	if paths != nil {
		for _, v := range *paths {
			DelPath(v.Pid)
		}
	}

	// Delete Relations
	if !DelRelationByPath(pid) {
		return false
	}

	// Delete Path
	lockPath.Lock()
	defer lockPath.Unlock()
	err := db.Table(TablePath).Where("pid=?", pid).Delete(&Path{}).Error

	return err == nil
}

func GetPathInfo(pid int64) *Path {
	if !Connected {
		Connect()
	}
	if pid == 0 {
		return nil
	}
	path := &Path{}
	err := db.Table(TablePath).Where("pid=?", pid).First(path).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return path
}

func GetPathInfos() *Paths {
	if !Connected {
		Connect()
	}
	paths := make(Paths, 0)
	err := db.Table(TablePath).Find(&paths).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &paths
}

func GetPathInfosByParent(ppid int64) *Paths {
	if !Connected {
		Connect()
	}

	paths := make(Paths, 0)
	err := db.Table(TablePath).Where("ppid=?", ppid).Find(&paths).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &paths
}

func GetPathInfosByPid(pid []int64) *Paths {
	if !Connected {
		Connect()
	}

	paths := make(Paths, 0)
	err := db.Table(TablePath).Where("pid IN (?)", pid).Find(&paths).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &paths
}

func GetPathInfosByFile(fid int64) *Paths {
	if !Connected {
		Connect()
	}

	paths := make(Paths, 0)
	err := db.Table(TablePath).Find(&paths, GetRelationsByFile(fid).Pid()).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &paths
}

func PathAccessed(pid int64) {
	if !Connected {
		Connect()
	}
	lockPath.Lock()
	defer lockPath.Unlock()
	res := db.Table(TablePath).Where("pid=?", pid).UpdateColumns(map[string]interface{}{
		"accessed": UnixTime(),
		"views":    gorm.Expr("views+1"),
	})
	if res.Error != nil {
		log.Println(res.Error)
	}
}

func UpdateAddNewFile(pid int64) {
	if !Connected {
		Connect()
	}
	lockPath.Lock()
	defer lockPath.Unlock()
	res := db.Table(TablePath).Where("pid=?", pid).UpdateColumns(map[string]interface{}{
		"size": gorm.Expr("size+1"),
	})
	if res.Error != nil {
		log.Println(res.Error)
	}
}

func UpdateDelFile(pid int64) {
	if !Connected {
		Connect()
	}
	lockPath.Lock()
	defer lockPath.Unlock()
	res := db.Table(TablePath).Where("pid=?", pid).UpdateColumns(map[string]interface{}{
		"size": gorm.Expr("size-1"),
	})
	if res.Error != nil {
		log.Println(res.Error)
	}
}

func PathSearchByTitle(title string) *Paths {
	if !Connected {
		Connect()
	}
	paths := make(Paths, 0)
	err := db.Table(TablePath).Where("title LIKE ?", fmt.Sprintf("%%%s%%", title)).Find(&paths).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &paths
}
