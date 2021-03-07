package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"via/db"
	"via/permission"
	"via/record"
)

// POST: Upload Path
func pathAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		ppid := Int64(r.FormValue("pp"))
		if ppid <= 0 {
			Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
			return
		}
		info := db.GetPathInfo(ppid)
		res := WritePermit(w, r, info)
		if res < 0 {
			return
		}
		title := r.FormValue("title")
		if len(title) == 0 {
			return
		}
		per := permission.ParseString(r.FormValue("permission"))
		if per == -1 {
			return
		}
		pwd := r.FormValue("password")
		comment := r.FormValue("comment")
		record.Printf("Add Path PPID:[%d] Title:[%s]\n", ppid, title)
		if db.AddPath(ppid, title, comment, pwd, per) {
			Fprint(w, TplRedirect(fmt.Sprintf("/?p=%d", ppid)))
		} else {
			Fprint(w, TplConfirmRedirect("Failed", "history", "history"))
		}
	}
}

func pathInfoEdit(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.ParseInt(r.FormValue("p"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}
	pathInfo := db.GetPathInfo(pid)
	if pathInfo == nil {
		log.Println(err)
		return
	}
	res := WritePermit(w, r, pathInfo)
	if res < 0 {
		return
	}
	pathInfo.Ppid = Int64(r.FormValue("ppid"))
	pathInfo.Title = r.FormValue("title")
	pathInfo.Comment = r.FormValue("comment")
	pathInfo.Size = Int64(r.FormValue("size"))
	pathInfo.Password = r.FormValue("password")
	pathInfo.Permission = permission.ParseString(r.FormValue("permission"))
	if db.UpdatePathInfo(pathInfo) {
		Fprint(w, TplRedirect("/"))
	} else {
		Fprint(w, TplConfirmRedirect("Failed", "/", "/"))
	}
	return
}

func pathInfoEditPage(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.ParseInt(r.FormValue("p"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}
	ph := db.GetPathInfo(pid)
	if ph == nil {
		log.Println(err)
		return
	}
	res := WritePermit(w, r, ph)
	if res < 0 {
		return
	}
	pwd := ""
	if len(r.FormValue("pwd")) != 0 {
		pwd = fmt.Sprintf(`<input type="hidden" name="pwd" value="%s">`, r.FormValue("pwd"))
	}
	tpl := `<!DOCTYPE html>
<title>Path Info Update</title>

<form action="/update/info/path" method="post">
	<input type="hidden" name="p" value="%d">
	%s
	<label>PPID:<input type="number" name="ppid" value="%d"></label><br/><br/>
	<label>Title:<input type="text" name="title" placeholder="Title" value="%s"></label><br/><br/>
	<label>Size:<input type="number" name="size" placeholder="Size" value="%d"></label><br/><br/>
	<label>Password:<input type="password" name="password" placeholder="Password" value="%s"></label><br/><br/>
	<label>Permission:<input type="text" name="permission" placeholder="Root Public Private Protected Login" value="%s"></label><br/><br/>
	<label><textarea name="comment" rows="10" cols="20">%s</textarea></label><br/><br/>
	<input type="submit" value="Submit">
</form>
`
	Fprintf(w, tpl, pid, pwd, ph.Ppid, ph.Title, ph.Size, ph.Password, permission.ToString(ph.Permission), ph.Comment)
	return

}

func pathDel(w http.ResponseWriter, r *http.Request) {
	pid := Int64(r.FormValue("p"))
	if pid <= 1 {
		Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
		return
	}
	p := db.GetPathInfo(pid)
	if p == nil {
		Fprint(w, TplConfirmRedirect("ERROR", "history", "history"))
		return
	}
	res := WritePermit(w, r, p)
	if res < 0 {
		return
	}
	record.Printf("Del Path PID:[%d]\n", pid)
	if db.DelPath(pid) {
		Fprint(w, TplConfirmRedirect("Successful", fmt.Sprintf("/?p=%d", p.Ppid), fmt.Sprintf("/?p=%d", p.Ppid)))
	} else {
		Fprint(w, TplConfirmRedirect("Failed", fmt.Sprintf("/?p=%d", p.Ppid), fmt.Sprintf("/?p=%d", p.Ppid)))
	}
}

func pathUpload(w http.ResponseWriter, r *http.Request) {
	pwd := ""
	if len(r.FormValue("pwd")) != 0 {
		pwd = `<label>Dir Password:<input type="password" name="pwd" placeholder="Dir Password"></label><br/><br/>`
	}
	tpl := `<!DOCTYPE html>
<title>MakeDir</title>
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

<form action="/path" method="post">
	<input type="hidden" name="pp" id="pp">
	%s
	<label>Title:<input type="text" name="title" placeholder="Title"></label><br/><br/>
	<label>Password:<input type="password" name="password" placeholder="Password"></label><br/><br/>
	<label>Permission:<input type="text" name="permission" placeholder="Root Public Private Protected Login" value="rw------rw"></label><br/><br/>
	<label><textarea name="comment" rows="10" cols="20"></textarea></label><br/><br/>
	<input type="submit" value="Submit">
</form>
<script type="text/javascript">
	document.getElementById('pp').value = getQueryVariable('p');
</script>
`
	Fprintf(w, tpl, pwd)

	return
}
