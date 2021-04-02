package handler

import (
	"fmt"
	"github.com/AkvicorEdwards/util"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"via/db"
	"via/def"
	"via/permission"
	"via/record"
)

//	GET: Get file content
//	POST: Upload file
func fileGet(w http.ResponseWriter, r *http.Request) {
	fid, err := strconv.ParseInt(r.FormValue("f"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}
	fileInfo := db.GetFileInfo(fid)
	if fileInfo == nil {
		log.Println(err)
		return
	}
	res := ReadPermit(w, r, fileInfo)
	if res < 0 {
		return
	}
	filename := filepath.Join(def.Path, fmt.Sprint(fileInfo.Filepath), fmt.Sprint(fileInfo.Filename))
	func() {
		ac := r.Header.Values("Accept")
		for _, v := range ac {
			if strings.Contains(v, "text/html") {
				go db.FileAccessed(fid)
				record.Printf("GET File: [%s]\n", filename)
			}
		}
	}()

	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()

	w.Header().Add("Content-Type", FileType(filename))
	//w.Header().Add("Content-Disposition", "attachment; filename=\""+fileInfo.Title+"\"")
	w.Header().Add("Content-Disposition", "filename=\""+fileInfo.Title+"\"")
	http.ServeContent(w, r, fileInfo.Title, time.Now(), file)
	return
}

func fileInfoEdit(w http.ResponseWriter, r *http.Request) {
	pid := Int64(r.FormValue("p"))
	ul := fmt.Sprintf("/?p=%d", pid)
	fid := Int64(r.FormValue("f"))

	fileInfo := db.GetFileInfo(fid)
	if fileInfo == nil {
		return
	}
	res := WritePermit(w, r, fileInfo)
	if res < 0 {
		return
	}
	fileInfo.Title = r.FormValue("title")
	fileInfo.Comment = r.FormValue("comment")
	fileInfo.MD5 = r.FormValue("md5")
	fileInfo.SHA256 = r.FormValue("sha256")
	fileInfo.Size = Int64(r.FormValue("size"))
	fileInfo.Password = r.FormValue("password")
	fileInfo.Permission = permission.ParseString(r.FormValue("permission"))
	fileInfo.Priority = Int64(r.FormValue("priority"))
	if db.UpdateFileInfo(fileInfo) {
		Fprint(w, TplRedirect(ul))
	} else {
		Fprint(w, TplConfirmRedirect("Failed", "history", ul))
	}
	return
}

func fileInfoEditPage(w http.ResponseWriter, r *http.Request) {
	fid := Int64(r.FormValue("f"))
	pid := Int64(r.FormValue("p"))
	fi := db.GetFileInfo(fid)
	if fi == nil {
		return
	}
	res := WritePermit(w, r, fi)
	if res < 0 {
		return
	}
	pwd := ""
	if len(r.FormValue("pwd")) != 0 {
		pwd = fmt.Sprintf(`<input type="hidden" name="pwd" value="%s">`, r.FormValue("pwd"))
	}
	tpl := `<!DOCTYPE html>
<title>File Info Update</title>

<form action="/update/info/file" method="post">
	<input type="hidden" name="f" value="%d">
	<input type="hidden" name="p" value="%d">
	%s
	<label>Title:<input type="text" name="title" placeholder="Title" value="%s"></label><br/><br/>
	<label>MD5:<input type="text" name="md5" placeholder="MD5" value="%s"></label><br/><br/>
	<label>SHA256:<input type="text" name="sha256" placeholder="SHA256" value="%s"></label><br/><br/>
	<label>Size:<input type="number" name="size" placeholder="Size" value="%d"></label><br/><br/>
	<label>Password:<input type="password" name="password" placeholder="Password" value="%s"></label><br/><br/>
	<label>Permission:<input type="text" name="permission" placeholder="Root Public Private Protected Login" value="%s"></label><br/><br/>
	<label>Priority:<input type="number" name="priority" value="%d"></label><br/><br/>
	<label><textarea name="comment" rows="10" cols="20">%s</textarea></label><br/><br/>
	<input type="submit" value="Submit">
</form>
`
	Fprintf(w, tpl, fid, pid, pwd, fi.Title, fi.MD5, fi.SHA256, fi.Size, fi.Password, permission.ToString(fi.Permission), fi.Priority, fi.Comment)
	return

}

func fileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			log.Println(err)
			return
		}
		pid := Int64(r.FormValue("p"))
		if pid <= 0 {
			Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
			return
		}
		url := fmt.Sprintf("/?p=%d", pid)
		info := db.GetPathInfo(pid)
		if info == nil {
			log.Println(err)
			return
		}
		res := ReadPermit(w, r, info)
		if res < 0 {
			Fprint(w, TplConfirmRedirect("ERROR: Need Dir Password", "history", "history"))
			return
		}
		title := r.FormValue("title")
		comment := r.FormValue("comment")
		password := r.FormValue("password")
		permissions := permission.ParseString(strings.TrimSpace(r.FormValue("permission")))
		if permissions == -1 {
			Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
			return
		}
		priority := Int64(r.FormValue("priority"))
		m5 := ""
		s6 := ""
		filePath := int64(0)
		filename := int64(0)
		size := int64(0)
		if !func() bool {
			file, handler, err := r.FormFile("file")
			if err != nil {
				log.Println(err)
				return false
			}
			defer func() { _ = file.Close() }()
			size = handler.Size

			fp := db.CalculateFilePath()
			if fp == nil {
				return false
			}
			de := true
			defer func() {
				if de {
					_ = db.DecreaseFilePathSize(fp.Pid)
				}
			}()
			S := func(str interface{}) string { return fmt.Sprint(str) }
			filePath = fp.Pid
			filename = fp.Filename
			fullPath := path.Join(def.Path, S(filePath))
			if stat := util.FileStat(fullPath); stat != 1 {
				if stat == 2 {
					err = os.RemoveAll(fullPath)
					if err != nil {
						return false
					}
				}
				err := os.MkdirAll(fullPath, 0700)
				if err != nil {
					return false
				}
			}
			if !func() bool {
				deleted := false
				defer func() {
					if deleted {
						_ = os.Remove(path.Join(def.Path, S(filePath), S(filename)))
						record.Printf("Delete Upload File: [%s]\n", path.Join(S(filePath), S(filename)))
					}
				}()
				dst, err := os.Create(path.Join(def.Path, S(filePath), S(filename)))
				if err != nil {
					Fprint(w, TplConfirmRedirect("ERROR", url, "history"))
					return false
				}
				defer func() { _ = dst.Close() }()
				n, err := io.Copy(dst, file)
				if err != nil {
					deleted = true
					Fprint(w, TplConfirmRedirect("ERROR", url, "history"))
					return false
				}
				record.Printf("Upload File: [%s]\n", path.Join(S(filePath), S(filename)))
				if size != n {
					deleted = true
					Fprint(w, TplConfirmRedirect("ERROR", url, "history"))
					return false
				}
				wg := sync.WaitGroup{}
				wg.Add(2)
				md5 := [16]byte{}
				sha256 := [32]byte{}
				go func(p string) {
					pt := path.Join(S(filePath), S(filename))
					record.Printf("Calculate MD5 [%s]\n", pt)
					md5, _ = util.MD5File(p)
					record.Printf("MD5 [%s] [%x]\n", pt, md5)
					wg.Done()
				}(path.Join(def.Path, S(filePath), S(filename)))

				time.Sleep(time.Microsecond)

				go func(p string) {
					pt := path.Join(S(filePath), S(filename))
					record.Printf("Calculate SHA256 [%s]\n", pt)
					sha256, _ = util.SHA256File(p)
					record.Printf("SHA256 [%s] [%x]\n", pt, sha256)
					wg.Done()
				}(path.Join(def.Path, S(filePath), S(filename)))

				wg.Wait()
				m5 = fmt.Sprintf("%X", md5)
				s6 = fmt.Sprintf("%X", sha256)
				if fe := db.FileExist(m5, s6); fe != nil {
					lst := ""
					for _, v := range *fe {
						lst += fmt.Sprintf("[%d][%s]", v.Fid, v.Title)
					}
					Fprint(w, TplConfirmRedirect(lst, "history", "history"))
					deleted = true
					return false
				}
				return true
			}() {
				return false
			}
			de = false
			return true
		}() {
			return
		}
		if fid := db.AddFile(title, comment, m5, s6, password,
			filename, filePath, size, permissions, priority); fid != -1 {
			record.Printf("Add Relation PID:[%d] FID:[%d]\n", pid, fid)
			if db.AddRelation(pid, fid) {
				Fprint(w, TplRedirect(url))
			} else {
				Fprint(w, TplConfirmRedirect("Unable to establish file-folder relationship", url, "history"))
			}
		} else {
			Fprint(w, TplConfirmRedirect("Failed", url, "history"))
		}
	}

}

func fileUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			log.Println(err)
			return
		}
		pid := Int64(r.FormValue("p"))
		if pid <= 0 {
			Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
			return
		}
		fid := Int64(r.FormValue("f"))
		if fid <= 0 {
			Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
			return
		}
		// TODO

		url := fmt.Sprintf("/?p=%d", pid)
		info := db.GetPathInfo(pid)
		if info == nil {
			log.Println(err)
			return
		}
		res := ReadPermit(w, r, info)
		if res < 0 {
			Fprint(w, TplConfirmRedirect("ERROR: Need Dir Password", "history", "history"))
			return
		}
		title := r.FormValue("title")
		comment := r.FormValue("comment")
		password := r.FormValue("password")
		permissions := permission.ParseString(strings.TrimSpace(r.FormValue("permission")))
		if permissions == -1 {
			Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
			return
		}
		priority := Int64(r.FormValue("priority"))
		m5 := ""
		s6 := ""
		filePath := int64(0)
		filename := int64(0)
		size := int64(0)
		if !func() bool {
			file, handler, err := r.FormFile("file")
			if err != nil {
				log.Println(err)
				return false
			}
			defer func() { _ = file.Close() }()
			size = handler.Size

			fp := db.CalculateFilePath()
			if fp == nil {
				return false
			}
			de := true
			defer func() {
				if de {
					_ = db.DecreaseFilePathSize(fp.Pid)
				}
			}()
			S := func(str interface{}) string { return fmt.Sprint(str) }
			filePath = fp.Pid
			filename = fp.Filename
			fullPath := path.Join(def.Path, S(filePath))
			if stat := util.FileStat(fullPath); stat != 1 {
				if stat == 2 {
					err = os.RemoveAll(fullPath)
					if err != nil {
						return false
					}
				}
				err := os.MkdirAll(fullPath, 0700)
				if err != nil {
					return false
				}
			}
			if !func() bool {
				deleted := false
				defer func() {
					if deleted {
						_ = os.Remove(path.Join(def.Path, S(filePath), S(filename)))
						record.Printf("Delete Upload File: [%s]\n", path.Join(S(filePath), S(filename)))
					}
				}()
				dst, err := os.Create(path.Join(def.Path, S(filePath), S(filename)))
				if err != nil {
					return false
				}
				defer func() { _ = dst.Close() }()
				n, err := io.Copy(dst, file)
				if err != nil {
					deleted = true
					return false
				}
				record.Printf("Upload File: [%s]\n", path.Join(S(filePath), S(filename)))
				if size != n {
					deleted = true
					return false
				}
				return true
			}() {
				return false
			}
			wg := sync.WaitGroup{}
			wg.Add(2)
			md5 := [16]byte{}
			sha256 := [32]byte{}
			go func(p string) {
				pt := path.Join(S(filePath), S(filename))
				record.Printf("Calculate MD5 [%s]\n", pt)
				md5, _ = util.MD5File(p)
				record.Printf("MD5 [%s] [%x]\n", pt, md5)
				wg.Done()
			}(path.Join(def.Path, S(filePath), S(filename)))

			time.Sleep(time.Microsecond)

			go func(p string) {
				pt := path.Join(S(filePath), S(filename))
				record.Printf("Calculate SHA256 [%s]\n", pt)
				sha256, _ = util.SHA256File(p)
				record.Printf("SHA256 [%s] [%x]\n", pt, sha256)
				wg.Done()
			}(path.Join(def.Path, S(filePath), S(filename)))

			wg.Wait()
			m5 = fmt.Sprintf("%X", md5)
			s6 = fmt.Sprintf("%X", sha256)
			de = false
			return true
		}() {
			Fprint(w, TplConfirmRedirect("ERROR", url, "history"))
			return
		}
		if fid := db.AddFile(title, comment, m5, s6, password,
			filename, filePath, size, permissions, priority); fid != -1 {
			record.Printf("Add Relation PID:[%d] FID:[%d]\n", pid, fid)
			if db.AddRelation(pid, fid) {
				Fprint(w, TplRedirect(url))
			} else {
				Fprint(w, TplConfirmRedirect("Unable to establish file-folder relationship", url, "history"))
			}
		} else {
			Fprint(w, TplConfirmRedirect("Failed", url, "history"))
		}
	}

}

func fileIndex(w http.ResponseWriter, r *http.Request) {
	pid := Int64(r.FormValue("p"))
	pid = Max(pid, 1)
	curPath := db.GetPathInfo(pid)
	res := ReadPermit(w, r, curPath)
	if res < 0 {
		return
	}

	perWrite := false
	res = WritePermit(w, r, curPath)
	if res > 0 {
		perWrite = true
	}

	fids := db.GetFileInfosByPath(pid)
	dirs := db.GetPathInfosByParent(pid)

	by, _ := strconv.Atoi(r.FormValue("by"))
	order, _ := strconv.Atoi(r.FormValue("order"))

	fids.Sort(by, order == 1)
	dirs.Sort(by, order == 1)

	tplHead := `<!DOCTYPE html>
<title>File Index</title>
<script type="text/javascript">
function mbar(sobj) {
	var docurl = sobj.options[sobj.selectedIndex].value;
	if (docurl != "") {
		window.location.replace(docurl);
	}
}

function delete_confirm(url, msg) {
	let res=confirm(msg);
	if (res === true) {
		window.location.href=url;
	}
}
</script>
`
	if perWrite {
		pwd := ""
		if curPath.Permit(permission.WriteProtectedPwd) {
			pwd = "&pwd=y"
		}
		tplHead += fmt.Sprintf(`<h4><a href="/page/upload/file?p=%d%s" style="color:black;">Upload</a>`+
			`&nbsp;&nbsp;&nbsp;<a href="/page/upload/path?p=%d%s" style="color:black;">MakeDir</a></h4>`,
			pid, pwd, pid, pwd)
	}
	tplHead += `<h4><input type="text" id="search_val">&nbsp;&nbsp;&nbsp;<select name="search" id="search">` +
		`<option value="title">Title</option><option value="id">ID</option><option value="sha256">SHA256</option>` +
		`<option value="md5">MD5</option></select>&nbsp;&nbsp;&nbsp;<a style="color:black;" ` +
		`onclick="window.location.href='/search/file?search='+document.getElementById('search').value+` +
		`'&val='+document.getElementById('search_val').value;">Search</a></h4>`
	tplHead += fmt.Sprintf("<h3>%s</h3>", func() string {
		id := pid
		res := ""
		deadLoop := false
		for {
			p := db.GetPathInfo(id)
			if p == nil {
				break
			}
			res = fmt.Sprintf(`<a href="/?p=%d" style="color:black;">%s</a>/%s`, p.Pid, p.Title, res)
			id = p.Ppid
			// parent is self
			if p.Pid == p.Ppid {
				break
			}
			// dead loop
			if deadLoop && p.Pid == curPath.Pid {
				break
			}
			deadLoop = true
		}
		return res
	}())
	Fprint(w, tplHead)
	Fprint(w, strings.ReplaceAll(`
<div>
	<select onchange=mbar(this) name="select">
		<option selected>defaults</option>
		<option value="?p=?&by=1&order=1">ID DESC</option>
		<option value="?p=?&by=1">ID ASC</option>
		<option value="?p=?&by=2&order=1">Size DESC</option>
		<option value="?p=?&by=2">Size ASC</option>
		<option value="?p=?&by=3">Created DESC</option>
		<option value="?p=?&by=3&order=1">Created ASC</option>
		<option value="?p=?&by=4&order=1">Modified DESC</option>
		<option value="?p=?&by=4">Modified ASC</option>
		<option value="?p=?&by=5&order=1">Accessed DESC</option>
		<option value="?p=?&by=5">Accessed ASC</option>
		<option value="?p=?&by=6">Views DESC</option>
		<option value="?p=?&by=6&order=1">Views ASC</option>
		<option value="?p=?&by=7&order=1">Priority DESC</option>
		<option value="?p=?&by=7">Priority ASC</option>
	</select>
</div>
<br/>`, "p=?", fmt.Sprintf("p=%d", pid)))

	for _, v := range *dirs {
		Fprint(w, `<div style="margin: 4px; border-top:3px dashed grey">`)
		Fprintf(w, `<a>[%s] ID:[%d] Created:[%s] Modified:[%s] Accessed:[%s] Views:[%d] Size:[%d]</a>`,
			permission.ToString(v.Permission), v.Pid, TimeUnixFormat(v.Created), TimeUnixFormat(v.Modified),
			TimeUnixFormat(v.Accessed), v.Views, v.Size)
		if perWrite {
			pwd := ""
			if v.Permit(permission.WriteProtectedPwd) {
				pwd = "&pwd=y"
			}
			if def.DeleteButton {
				Fprintf(w, `&nbsp;<a onclick="delete_confirm('/del/path?p=%d%s', '%d %s')" style="color:red;">DELETE</a>`, v.Pid, pwd, v.Pid, v.Title)
			}
			if def.EditButton {
				Fprintf(w, `&nbsp;<a href="/page/update/info/path?p=%d%s" style="color:orange;">EDIT</a>`, v.Pid, pwd)
			}
		}
		Fprint(w, fmt.Sprintf(`<br /><strong><a href="/?p=%d" style="color:blue;">%s</a></strong>`, v.Pid, v.Title))
		Fprint(w, `</div>`)
	}
	for _, v := range *fids {
		Fprint(w, `<div style="margin: 4px; border-top:3px dashed grey">`)
		Fprintf(w, `<a>[%s] ID:[%d] Created:[%s] Modified:[%s] Accessed:[%s] Views:[%d] Priority:[%d] Size:[%s/%d]</a>`,
			permission.ToString(v.Permission), v.Fid, TimeUnixFormat(v.Created), TimeUnixFormat(v.Modified),
			TimeUnixFormat(v.Accessed), v.Views, v.Priority, fmt.Sprintf("%.2f", float32(v.Size)/1024.0/1024.0), v.Size)
		if perWrite {
			pwd := ""
			if v.Permit(permission.WriteProtectedPwd) {
				pwd = "&pwd=y"
			}
			if def.DeleteButton {
				Fprintf(w, `&nbsp;<a onclick="delete_confirm('/del/file?f=%d&p=%d%s', '%d %s')" style="color:red;">DELETE</a>`, v.Fid, pid, pwd, v.Fid, v.Title)
			}
			if def.EditButton {
				Fprintf(w, `&nbsp;<a href="/page/update/info/file?f=%d&p=%d%s" style="color:orange;">EDIT</a>`, v.Fid, pid, pwd)
			}
			if def.UpdateButton {
				Fprintf(w, `&nbsp;<a href="/page/update/file?f=%d&p=%d%s" style="color:purple;">UPDATE</a>`, v.Fid, pid, pwd)
			}
		}
		Fprintf(w, `<br /><a>MD5: [%s]<br />SHA256: [%s]</a>`, v.MD5, v.SHA256)
		Fprint(w, fmt.Sprintf(`<br /><strong><a href="/file?f=%d" style="color:green;">%s</a></strong>`, v.Fid, v.Title))
		Fprint(w, `</div>`)
	}
	go db.PathAccessed(pid)
	return
}

func fileUploadPage(w http.ResponseWriter, r *http.Request) {
	pwd := ""
	if len(r.FormValue("pwd")) != 0 {
		pwd = `<label>Dir Password:<input type="password" name="pwd" placeholder="Dir Password"></label><br/><br/>`
	}
	tpl := `<!DOCTYPE html>
<title>File Upload</title>
<script type="text/javascript">
function getQueryVariable(variable) {
	var query = window.location.search.substring(1);
	var vars = query.split("&");
	for (var i=0;i<vars.length;i++) {
		var pair = vars[i].split("=");
		if(pair[0] == variable){return pair[1];}
	}
	return(false);
}
</script>
<h4>
	<input type="text" id="search_val">&nbsp;&nbsp;&nbsp;<select name="search" id="search">
	<option value="title">Title</option><option value="id">ID</option><option value="sha256">SHA256</option>
	<option value="md5">MD5</option></select>&nbsp;&nbsp;&nbsp;
	<a style="color:black;" onclick="window.location.href='/search/file?` +
		`search='+document.getElementById('search').value+'&val='+document.getElementById('search_val').value;">Search</a>
</h4>
<form action="/upload/file" method="post" enctype="multipart/form-data">
	<input type="hidden" name="p" id="path">
	%s
	<label>Title:<input type="text" name="title" id="title" placeholder="Title"></label><br/><br/>
	<label>Password:<input type="password" name="password" placeholder="Password"></label><br/><br/>
	<label>Permission:<input type="text" name="permission" placeholder="Root Public Private Protected Login" value="rw------rw"></label><br/><br/>
	<label>Priority:<input type="number" name="priority" value="1000"></label><br/><br/>
	<label><textarea name="comment" rows="10" cols="20"></textarea></label><br/><br/>
	<label>File：<input type="file" name="file" placeholder="File" onChange="if(this.value)getFilename(this.value)"></label><br/><br/>
	<input type="submit" value="Submit">
</form>
<script type="text/javascript">
	document.getElementById('path').value = getQueryVariable('p');
	
	function getFilename(name) {
		let f1 = name.lastIndexOf("\\");
		if (f1 >= 0 && f1 < name.length) {
			document.getElementById("title").value = name.substring(f1+1, name.length);
		}
	}
</script>
`
	Fprintf(w, tpl, pwd)
	return
}

func fileUpdatePage(w http.ResponseWriter, r *http.Request) {
	fid := Int64(r.FormValue("f"))
	info := db.GetFileInfo(fid)
	if info == nil {
		return
	}
	pid := r.FormValue("p")
	pwd := ""
	if len(r.FormValue("pwd")) != 0 {
		pwd = `<label>Dir Password:<input type="password" name="pwd" placeholder="Dir Password"></label><br/><br/>`
	}
	tpl := `<!DOCTYPE html>
<title>File Update</title>

<form action="/update/file" method="post" enctype="multipart/form-data">
	<input type="hidden" type="number" name="f" value="%d">
	<input type="hidden" type="number" name="p" value="%s">
	%s
	<label>Title:<input type="text" name="title" id="title" placeholder="Title" value="%s"></label><br/><br/>
	<label>Password:<input type="password" name="password" placeholder="Password" value="%s"></label><br/><br/>
	<label>Permission:<input type="text" name="permission" placeholder="Root Public Private Protected Login" value="rw------rw" value="%s"></label><br/><br/>
	<label>Priority:<input type="number" name="priority" value="1000" value="%d"></label><br/><br/>
	<label><textarea name="comment" rows="10" cols="20">%s</textarea></label><br/><br/>
	<label>File：<input type="file" name="file" placeholder="File" onChange="if(this.value)getFilename(this.value)"></label><br/><br/>
	<input type="submit" value="Submit">
</form>

<script type="text/javascript">
	function getFilename(name) {
		let f1 = name.lastIndexOf("\\");
		if (f1 >= 0 && f1 < name.length) {
			document.getElementById("title").value = name.substring(f1+1, name.length);
		}
	}
</script>
`
	Fprintf(w, tpl, fid, pid, pwd, info.Title, info.Password, permission.ToString(info.Permission), info.Priority, info.Comment)
	return
}

func fileDel(w http.ResponseWriter, r *http.Request) {
	fid := Int64(r.FormValue("f"))
	if fid <= 0 {
		Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
		return
	}
	pid := Int64(r.FormValue("p"))
	if pid <= 0 {
		Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
		return
	}

	res := WritePermit(w, r, db.GetFileInfo(fid))
	if res < 0 {
		return
	}
	record.Printf("Del Relation FID:[%d] PID:[%d]\n", fid, pid)
	if db.DelRelation(pid, fid) {
		Fprint(w, TplConfirmRedirect("Successful", fmt.Sprintf("/?p=%d", pid), fmt.Sprintf("/?p=%d", pid)))
	} else {
		Fprint(w, TplConfirmRedirect("Failed", fmt.Sprintf("/?p=%d", pid), fmt.Sprintf("/?p=%d", pid)))
	}
}

func fileSearch(w http.ResponseWriter, r *http.Request) {
	var files *db.Files
	var paths *db.Paths
	switch r.FormValue("search") {
	case "id":
		file := db.GetFileInfo(Int64(r.FormValue("val")))
		if file != nil {
			files = &db.Files{*file}
		}
	case "title":
		files = db.FileSearchByTitle(r.FormValue("val"))
		paths = db.PathSearchByTitle(r.FormValue("val"))
	case "sha256":
		files = db.FileSearchBySHA256(strings.ToUpper(r.FormValue("val")))
	case "md5":
		files = db.FileSearchByMD5(strings.ToUpper(r.FormValue("val")))
	}
	Fprint(w, `<!DOCTYPE html>
<title>File Search</title>
<script type="text/javascript">
function mbar(sobj) {
	var docurl = sobj.options[sobj.selectedIndex].value;
	if (docurl != "") {
		window.location.replace(docurl);
	}
}

function delete_confirm(url, msg) {
	let res=confirm(msg);
	if (res === true) {
		window.location.href=url;
	}
}
</script>
<h4><a href="/" style="color:black;">ROOT</a></h4>
`)
	Fprint(w, strings.ReplaceAll(`
<div>
	<select onchange=mbar(this) name="select">
		<option selected>defaults</option>
		<option value="/search/file?search=?&val=?&by=1&order=1">ID DESC</option>
		<option value="/search/file?search=?&val=?&by=1">ID ASC</option>
		<option value="/search/file?search=?&val=?&by=2&order=1">Size DESC</option>
		<option value="/search/file?search=?&val=?&by=2">Size ASC</option>
		<option value="/search/file?search=?&val=?&by=3">Created DESC</option>
		<option value="/search/file?search=?&val=?&by=3&order=1">Created ASC</option>
		<option value="/search/file?search=?&val=?&by=4&order=1">Modified DESC</option>
		<option value="/search/file?search=?&val=?&by=4">Modified ASC</option>
		<option value="/search/file?search=?&val=?&by=5&order=1">Accessed DESC</option>
		<option value="/search/file?search=?&val=?&by=5">Accessed ASC</option>
		<option value="/search/file?search=?&val=?&by=6">Views DESC</option>
		<option value="/search/file?search=?&val=?&by=6&order=1">Views ASC</option>
		<option value="/search/file?search=?&val=?&by=7&order=1">Priority DESC</option>
		<option value="/search/file?search=?&val=?&by=7">Priority ASC</option>
	</select>
</div>
<br/>`, "/search/file?search=?&val=?", fmt.Sprintf("/search/file?search=%s&val=%s",
	r.FormValue("search"), r.FormValue("val"))))

	by, _ := strconv.Atoi(r.FormValue("by"))
	order, _ := strconv.Atoi(r.FormValue("order"))
	if paths != nil {
		paths.Sort(by, order == 1)
		for _, v := range *paths {
			Fprint(w, `<div style="margin: 4px; border-top:3px dashed grey">`)
			Fprintf(w, `<a>[%s] ID:[%d] Created:[%s] Modified:[%s] Accessed:[%s] Views:[%d] Size:[%d]</a>`,
				permission.ToString(v.Permission), v.Pid, TimeUnixFormat(v.Created), TimeUnixFormat(v.Modified),
				TimeUnixFormat(v.Accessed), v.Views, v.Size)

			pwd := ""
			if v.Permit(permission.WriteProtectedPwd) {
				pwd = "&pwd=y"
			}
			if def.DeleteButton {
				Fprintf(w, `&nbsp;<a onclick="delete_confirm('/del/path?p=%d%s', '%d %s')" style="color:red;">DELETE</a>`, v.Pid, pwd, v.Pid, v.Title)
			}
			if def.EditButton {
				Fprintf(w, `&nbsp;<a href="/page/update/info/path?p=%d%s" style="color:orange;">EDIT</a>`, v.Pid, pwd)
			}

			Fprint(w, fmt.Sprintf(`<br /><strong><a href="/?p=%d" style="color:blue;">%s</a></strong>`, v.Pid, v.Title))
			Fprint(w, `</div>`)
		}
	}
	if files != nil {
		files.Sort(by, order == 1)
		for _, v := range *files {
			Fprint(w, `<div style="margin: 4px; border-top:3px dashed grey">`)
			Fprintf(w, `<a>[%s] ID:[%d] Created:[%s] Modified:[%s] Accessed:[%s] Views:[%d] Priority:[%d] Size:[%s/%d]</a>`,
				permission.ToString(v.Permission), v.Fid, TimeUnixFormat(v.Created), TimeUnixFormat(v.Modified),
				TimeUnixFormat(v.Accessed), v.Views, v.Priority, fmt.Sprintf("%.2f", float32(v.Size)/1024.0/1024.0), v.Size)
			pwd := ""
			if v.Permit(permission.WriteProtectedPwd) {
				pwd = "&pwd=y"
			}
			if def.EditButton {
				Fprintf(w, `&nbsp;<a href="/page/update/info/file?f=%d&p=%d%s" style="color:orange;">EDIT</a>`, v.Fid, 0, pwd)
			}
			if def.UpdateButton {
				Fprintf(w, `&nbsp;<a href="/page/update/file?f=%d&p=%d%s" style="color:purple;">UPDATE</a>`, v.Fid, 0, pwd)
			}

			Fprintf(w, `<br /><a>MD5: [%s]<br />SHA256: [%s]</a>`, v.MD5, v.SHA256)
			Fprint(w, fmt.Sprintf(`<br /><strong><a href="/file?f=%d" style="color:green;">%s</a></strong>`, v.Fid, v.Title))
			Fprint(w, `</div>`)
		}
	}
}
