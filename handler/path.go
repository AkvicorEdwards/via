package handler

import (
	"fmt"
	"net/http"
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
