package db

import "log"

const TableRelation = "relation"

type Relation struct {
	Rid int64
	Pid int64
	Fid int64
}

type Relations []Relation

func (r *Relations) Rid() []int64 {
	res := make([]int64, len(*r))
	for k, v := range *r {
		res[k] = v.Rid
	}
	return res
}

func (r *Relations) Fid() []int64 {
	res := make([]int64, len(*r))
	for k, v := range *r {
		res[k] = v.Fid
	}
	return res
}

func (r *Relations) Pid() []int64 {
	res := make([]int64, len(*r))
	for k, v := range *r {
		res[k] = v.Pid
	}
	return res
}

func AddRelation(pid, fid int64) bool {
	if !Connected {
		Connect()
	}
	lockRelation.Lock()
	defer lockRelation.Unlock()

	path := GetPathInfo(pid)
	if path == nil {
		return false
	}
	file := GetFileInfo(fid)
	if file == nil {
		return false
	}

	relation := Relation{
		Rid: GetInc(TableRelation) + 1,
		Pid: pid,
		Fid: fid,
	}

	if err := db.Table(TableRelation).Create(&relation).Error; err != nil {
		return false
	}

	UpdateInc(TableRelation, relation.Rid)

	UpdateAddNewFile(pid)

	return true
}

func DelRelation(pid, fid int64) bool {
	if !Connected {
		Connect()
	}
	lockRelation.Lock()
	defer lockRelation.Unlock()
	if db.Table(TableRelation).Where("pid=? AND fid=?", pid, fid).Delete(&Relation{}).Error != nil {
		return false
	}
	UpdateDelFile(pid)

	return true
}

func DelRelationByFile(fid int64) bool {
	if !Connected {
		Connect()
	}
	rel := GetRelationsByFile(fid)
	if rel == nil {
		return false
	}
	pid := rel.Pid()
	for _, v := range pid {
		UpdateDelFile(v)
	}
	rid := rel.Rid()

	lockRelation.Lock()
	defer lockRelation.Unlock()
	return db.Table(TableRelation).Where("rid IN (?)", rid).Delete(&Relation{}).Error == nil
}

func DelRelationByPath(pid int64) bool {
	if !Connected {
		Connect()
	}
	p := GetPathInfo(pid)
	if p == nil {
		return false
	}
	r := GetRelationsByPath(pid)
	if r == nil {
		return false
	}
	rid := r.Rid()
	lockRelation.Lock()
	defer lockRelation.Unlock()
	db.Table(TableRelation).Where("rid IN (?)", rid).Delete(&Relation{})

	UpdateDelFile(p.Ppid)
	return true
}

func GetRelationsByPath(pid int64) *Relations {
	if !Connected {
		Connect()
	}
	relations := make(Relations, 0)
	err := db.Table(TableRelation).Where("pid=?", pid).Find(&relations).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &relations
}

func GetRelationsByFile(fid int64) *Relations {
	if !Connected {
		Connect()
	}
	relations := make(Relations, 0)
	err := db.Table(TableRelation).Where("fid=?", fid).Find(&relations).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &relations
}

func GetRelations() *Relations {
	if !Connected {
		Connect()
	}
	relations := make(Relations, 0)
	err := db.Table(TableRelation).Find(&relations).Error
	if err != nil {
		log.Println(err)
		return nil
	}
	return &relations
}
