package db

import (
	"github.com/AkvicorEdwards/util"
	"log"
	"via/def"
)

func Verify(md5, sha256, clean bool, maxRoutineNum int) (walk *Walk) {
	if util.FileStat(def.Path) != 1 {
		log.Println("Missing Root Path")
		return nil
	}
	walk = NewWalk(md5, sha256, clean, maxRoutineNum)
	walk.Walk()

	return
}

func VerifyPath() *Paths {
	paths := GetPathInfos()
	size := GetInc(TablePath) + 1
	uf := NewUnionFind(size)
	for _, v := range *paths {
		uf.Union(v.Pid, v.Ppid)
	}
	parents := uf.CountSets()
	if len(parents) == 1 {
		return nil
	}
	return GetPathInfosByPid(parents)
}

func VerifyRelation() *Files {
	rel := GetRelations()
	files := GetFileInfosNotInFids(rel.Fid())
	if len(*files) == 0 {
		files = nil
	}
	return files
}
