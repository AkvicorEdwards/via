package db

import (
	"fmt"
	"github.com/AkvicorEdwards/util"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"via/def"
)

type MutexStrings struct {
	s []string
	sync.RWMutex
}

func (m *MutexStrings) Val() []string {
	m.RLock()
	defer m.RUnlock()
	cp := make([]string, len(m.s))
	copy(cp, m.s)
	return cp
}

func (m *MutexStrings) Add(s string) {
	m.Lock()
	defer m.Unlock()
	m.s = append(m.s, s)
}

func NewMutexStrings() *MutexStrings {
	return &MutexStrings{
		s:       make([]string, 0),
		RWMutex: sync.RWMutex{},
	}
}

type MutexString struct {
	s string
	sync.RWMutex
}

func (m *MutexString) Val() string {
	m.RLock()
	defer m.RUnlock()
	return m.s
}

func (m *MutexString) Set(s string) {
	m.Lock()
	defer m.Unlock()
	m.s = s
}

type MutexInt struct {
	i int
	sync.RWMutex
}

func (m *MutexInt) Val() int {
	m.RLock()
	defer m.RUnlock()
	return m.i
}

func (m *MutexInt) Set(v int) {
	m.Lock()
	defer m.Unlock()
	m.i = v
}

func (m *MutexInt) Inc() {
	m.Lock()
	defer m.Unlock()
	m.i += 1
}

func (m *MutexInt) Dec() {
	m.Lock()
	defer m.Unlock()
	m.i -= 1
}

func NewMutexInt(i int) *MutexInt {
	return &MutexInt{
		i:       i,
		RWMutex: sync.RWMutex{},
	}
}

type MutexBool struct {
	b bool
	sync.RWMutex
}

func (m *MutexBool) Val() bool {
	m.RLock()
	defer m.RUnlock()
	return m.b
}

func (m *MutexBool) Set(b bool) {
	m.Lock()
	defer m.Unlock()
	m.b = b
}

func NewMutexBool(b bool) *MutexBool {
	return &MutexBool{
		b:       b,
		RWMutex: sync.RWMutex{},
	}
}

type VerifyResult struct {
	PassedMD5    []int64
	PassedSHA256 []int64
	Damaged      []int64
	Missing      []int64
	Illegal      []string
	Error        []string
	Deleted      *MutexStrings
	DeleteFailed *MutexStrings
	sync.RWMutex
}

func (v *VerifyResult) PassedMD5Add(d int64) {
	v.Lock()
	defer v.Unlock()
	v.PassedMD5 = append(v.PassedMD5, d)
}

func (v *VerifyResult) PassedMD5Val() []int64 {
	v.RLock()
	defer v.RUnlock()
	cp := make([]int64, len(v.PassedMD5))
	copy(cp, v.PassedMD5)
	return cp
}

func (v *VerifyResult) PassedSHA256Add(d int64) {
	v.Lock()
	defer v.Unlock()
	v.PassedSHA256 = append(v.PassedSHA256, d)
}

func (v *VerifyResult) PassedSHA256Val() []int64 {
	v.RLock()
	defer v.RUnlock()
	cp := make([]int64, len(v.PassedSHA256))
	copy(cp, v.PassedSHA256)
	return cp
}

func (v *VerifyResult) DamagedAdd(d int64) {
	v.Lock()
	defer v.Unlock()
	v.Damaged = append(v.Damaged, d)
}

func (v *VerifyResult) DamagedVal() []int64 {
	v.RLock()
	defer v.RUnlock()
	cp := make([]int64, len(v.Damaged))
	copy(cp, v.Damaged)
	return cp
}

func (v *VerifyResult) MissingAdd(m int64) {
	v.Lock()
	defer v.Unlock()
	v.Missing = append(v.Missing, m)
}

func (v *VerifyResult) MissingVal() []int64 {
	v.RLock()
	defer v.RUnlock()
	cp := make([]int64, len(v.Missing))
	copy(cp, v.Missing)
	return cp
}

func (v *VerifyResult) IllegalAdd(i string) {
	v.Lock()
	defer v.Unlock()
	v.Illegal = append(v.Illegal, i)
}

func (v *VerifyResult) IllegalVal() []string {
	v.RLock()
	defer v.RUnlock()
	cp := make([]string, len(v.Illegal))
	copy(cp, v.Illegal)
	return cp
}

func (v *VerifyResult) ErrorAdd(e string) {
	v.Lock()
	defer v.Unlock()
	v.Error = append(v.Error, e)
}

func (v *VerifyResult) ErrorVal() []string {
	v.RLock()
	defer v.RUnlock()
	cp := make([]string, len(v.Error))
	copy(cp, v.Error)
	return cp
}

func NewVerifyResult() *VerifyResult {
	return &VerifyResult{
		PassedMD5:       make([]int64, 0),
		PassedSHA256:       make([]int64, 0),
		Damaged:      make([]int64, 0),
		Missing:      make([]int64, 0),
		Illegal:      make([]string, 0),
		Error:        make([]string, 0),
		Deleted:      NewMutexStrings(),
		DeleteFailed: NewMutexStrings(),
		RWMutex:      sync.RWMutex{},
	}
}

type Walk struct {
	File *VerifyResult
	Path *VerifyResult

	// Calculate MD5
	MD5 *MutexBool
	// Calculate SHA256
	SHA256 *MutexBool
	// Clean file
	Clean *MutexBool
	// MaxRoutineNum
	MaxRoutineNum chan bool
	sync.WaitGroup
}

func NewWalk(md5, sha256, clean bool, maxRoutineNum int) *Walk {
	if maxRoutineNum <= 0 {
		maxRoutineNum = 1
	}
	if maxRoutineNum > 17 {
		maxRoutineNum = 17
	}
	return &Walk{
		File:          NewVerifyResult(),
		Path:          NewVerifyResult(),
		MD5:           NewMutexBool(md5),
		SHA256:        NewMutexBool(sha256),
		Clean:         NewMutexBool(clean),
		MaxRoutineNum: make(chan bool, maxRoutineNum),
		WaitGroup: sync.WaitGroup{},
	}
}

func (w *Walk) Walk() {
	rootPath := def.Path
	rootDirs, err := ioutil.ReadDir(rootPath)
	if err != nil {
		log.Println("Cannot read dir")
		return
	}
	rootSize := GetInc(TableFilePath)
	rootValidPathSize := int64(0)
	paths := GetFilePaths()
	if paths == nil {
		log.Println("Cannot get FilePaths")
		return
	}
	pt := make(map[string]*FilePath)
	for k, v := range *paths {
		pt[fmt.Sprint(v.Pid)] = &(*paths)[k]
	}
	for _, rootFi := range rootDirs {
		if rootFi.IsDir() {
			childPath := path.Join(rootPath, rootFi.Name())
			child, ok := pt[rootFi.Name()]
			if !ok {
				w.HandleIllegalPath(childPath)
				continue
			}
			delete(pt, rootFi.Name())
			childValidPathSize := int64(0)
			rootValidPathSize++
			childDirs, err := ioutil.ReadDir(childPath)
			if err != nil {
				w.HandleErrorPath(childPath)
				continue
			}
			files := GetFileInfosByFilepath(child.Pid)
			fi := make(map[string]*File)
			for k, v := range *files {
				//log.Println(fmt.Sprintf("%d/%d %v", v.Filepath,v.Filename, &v))
				fi[fmt.Sprintf("%d/%d", v.Filepath,v.Filename)] = &(*files)[k]
			}
			for _, childFi := range childDirs {
				if childFi.IsDir() { // Illegal
					// The dirs should not appear in this directory
					w.HandleIllegalPath(path.Join(childPath, childFi.Name()))
					continue
				} else {
					childPathFile := path.Join(childPath, childFi.Name())
					f, ok := fi[fmt.Sprintf("%d/%s", child.Pid,childFi.Name())]
					if !ok {
						w.HandleIllegalFile(childPathFile)
						continue
					}
					//log.Printf("%d %s %v", child.Pid,childFi.Name(), f)
					childValidPathSize++
					delete(fi, fmt.Sprintf("%d/%s", child.Pid,childFi.Name()))
					// Verify Hash value
					w.Add(2)
					go w.VerifyMD5(childPathFile, f)
					go w.VerifySHA256(childPathFile, f)
				}
			}
			if childValidPathSize != child.Size {
				w.HandleDamagedPath(child.Pid)
			}
			for _, v := range fi {
				w.HandleMissingFile(v.Fid)
			}
		} else { // Illegal
			// The files should not appear in this directory
			w.HandleIllegalFile(path.Join(def.Path, rootFi.Name()))
			continue
		}
	}
	if rootSize != rootValidPathSize {
		w.HandleDamagedPath(0)
	}
	for _, v := range pt {
		w.HandleMissingPath(v.Pid)
	}
	w.Wait()
}

func (w *Walk) VerifyMD5(filename string, f *File) {
	defer w.Done()
	if w.MD5.Val() {
		w.MaxRoutineNum <- true
		defer func() {
			<-w.MaxRoutineNum
		}()
		log.Printf("Calculate MD5 [%s]\n", filename)
		md5, err := util.MD5File(filename)
		if err != nil {
			w.File.DamagedAdd(f.Fid)
			return
		}
		if fmt.Sprintf("%X", md5) != f.MD5 {
			w.File.DamagedAdd(f.Fid)
			log.Printf("MD5 Err  [%s] %d/%d[%s] [%s]\n", filename, f.Filepath, f.Filename, fmt.Sprintf("%X", md5), f.MD5)
			return
		}
		//log.Printf("MD5 Pass [%s]\n", filename)
		w.File.PassedMD5Add(f.Fid)
	}
}

func (w *Walk) VerifySHA256(filename string, f *File) {
	defer w.Done()
	if w.SHA256.Val() {
		w.MaxRoutineNum <- true
		defer func() {
			<-w.MaxRoutineNum
		}()
		log.Printf("Calculate SHA256 [%s]\n", filename)
		sha256, err := util.SHA256File(filename)
		if err != nil {
			w.File.DamagedAdd(f.Fid)
			return
		}
		if fmt.Sprintf("%X", sha256) != f.SHA256 {
			w.File.DamagedAdd(f.Fid)
			log.Printf("SHA256 Err  [%s] %d/%d[%s] [%s]\n", filename, f.Filepath, f.Filename, fmt.Sprintf("%X", sha256), f.SHA256)
			return
		}
		//log.Printf("SHA256 Pass [%s]\n", filename)
		w.File.PassedSHA256Add(f.Fid)
	}
}

func (w *Walk) HandleIllegalFile(filename string) {
	w.File.IllegalAdd(filename)
	if w.Clean.Val() {
		err := os.RemoveAll(filename)
		if err != nil {
			w.File.DeleteFailed.Add(filename)
		} else {
			w.File.Deleted.Add(filename)
		}
	}
}

func (w *Walk) HandleDamagedFile(filename int64) {
	w.File.DamagedAdd(filename)
}

func (w *Walk) HandleIllegalPath(filepath string) {
	w.Path.IllegalAdd(filepath)
	if w.Clean.Val() {
		err := os.RemoveAll(filepath)
		if err != nil {
			w.Path.DeleteFailed.Add(filepath)
		} else {
			w.Path.Deleted.Add(filepath)
		}
	}
}

func (w *Walk) HandleDamagedPath(filepath int64) {
	w.Path.DamagedAdd(filepath)
}

func (w *Walk) HandleErrorPath(filepath string) {
	w.Path.ErrorAdd(filepath)
}

func (w *Walk) HandleErrorFile(filepath string) {
	w.File.ErrorAdd(filepath)
}

func (w *Walk) HandleMissingPath(filepath int64) {
	w.Path.MissingAdd(filepath)
}

func (w *Walk) HandleMissingFile(filepath int64) {
	w.File.MissingAdd(filepath)
}
