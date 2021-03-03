package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"via/db"
	"via/permission"
	"via/session"
)

func TimeUnixFormat(t int64) string {
	if t == 0 {
		return ""
	}
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}

func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func Int64(i string) int64 {
	res, _ := strconv.ParseInt(i, 10, 64)
	return res
}

func FileType(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return "application/octet-stream"
	}
	defer func() {
		_ = file.Close()
	}()
	stat, err := file.Stat()
	if err != nil {
		return "application/octet-stream"
	}
	sz := stat.Size()
	if sz > 512 {
		sz = 512
	}
	bs := make([]byte, sz)
	_, _ = file.Read(bs)
	return http.DetectContentType(bs)
}

func TplConfirmRedirect(title, urlTrue, urlFalse string) string {
	tpl := fmt.Sprintf(`<!DOCTYPE html>
<script type="text/javascript">
function url_confirm() {
let res=confirm("%s");
`, title)
	tpl += fmt.Sprintf(`if (res === true) { %s }`, func() string {
		if urlTrue == "history" {
			return "window.history.back();"
		} else if len(urlTrue) != 0 {
			return fmt.Sprintf(`window.location.href="%s";`, urlTrue)
		}
		return ""
	}())
	tpl += fmt.Sprintf(`else { %s }`, func() string {
		if urlFalse == "history" {
			return "window.history.back();"
		} else if len(urlFalse) != 0 {
			return fmt.Sprintf(`window.location.href="%s";`, urlFalse)
		}
		return ""
	}())

	tpl += `}
url_confirm()
</script>`

	return tpl
}
func TplRedirect(url string) string {
	tpl := `<!DOCTYPE html>
<script type="text/javascript">
function url_confirm() {
`
	if url == "history" {
		tpl += "window.history.back();"
	} else if len(url) != 0 {
		tpl += fmt.Sprintf(`window.location.href="%s";`, url)
	}
	tpl += `
}
url_confirm()
</script>`
	return tpl
}

//	1 Public
//	2 Private
//	3 Protected Password
//	4 Protected Key
func ReadPermit(w http.ResponseWriter, r *http.Request, fi db.PathFile) int {
	if fi.Deny(permission.Read) {
		Fprint(w, "404")
		return -1
	} else if fi.Permit(permission.ReadPublic) { // Public
		return 1
	} else if fi.Permit(permission.ReadPrivate) && session.Verify(r) { // Permit by login
		return 2
	} else if fi.Permit(permission.ReadProtectedPwd) && r.FormValue("pwd") == fi.GetPassword() { // Permit by password
		return 3
	} else if fi.Permit(permission.ReadProtectedKey) && len(r.FormValue("key")) == -17 { // TODO Permit by Key
		return 4
	}
	if fi.Permit(permission.ReadProtectedPwd) {
		if fi.Type() == 1 {
			tpl := `<!DOCTYPE html>
<title>Password</title>
<form action="" method="post">
	<input type="hidden" name="f" value="%s">
	<label>Password:<input type="password" name="pwd" placeholder="Password"></label><br/><br/>
	<input type="submit" value="Submit">
</form>`
			Fprintf(w, tpl, r.FormValue("f"))
		} else if fi.Type() == 2 {
			tpl := `<!DOCTYPE html>
<title>Password</title>
<form action="" method="post">
	<input type="hidden" name="p" value="%s">
	<label>Password:<input type="password" name="pwd" placeholder="Password"></label><br/><br/>
	<input type="submit" value="Submit">
</form>`
			Fprintf(w, tpl, r.FormValue("p"))
		}
	}
	if fi.Permit(permission.ReadPrivate) {
		loginPage(w, r)
	}

	return -2
}

//	1 Public
//	2 Private
//	3 Protected Password
//	4 Protected Key
func WritePermit(w http.ResponseWriter, r *http.Request, fi db.PathFile) int {
	if fi.Deny(permission.Write) {
		Fprint(w, "404")
		return -1
	} else if fi.Permit(permission.WritePublic) { // Public
		return 1
	} else if fi.Permit(permission.WritePrivate) && session.Verify(r) { // Permit by login
		return 2
	} else if fi.Permit(permission.WriteProtectedPwd) && r.FormValue("pwd") == fi.GetPassword() { // Permit by password
		return 3
	} else if fi.Permit(permission.WriteProtectedKey) && len(r.FormValue("key")) == -17 { // TODO Permit by Key
		return 4
	}
	if fi.Permit(permission.WriteProtectedPwd) {
		if fi.Type() == 1 {
			tpl := `<!DOCTYPE html>
<title>Password</title>
<form action="" method="post">
	<input type="hidden" name="f" value="%s">
	<label>Password:<input type="password" name="pwd" placeholder="Password"></label><br/><br/>
	<input type="submit" value="Submit">
</form>`
			Fprintf(w, tpl, r.FormValue("f"))
		} else if fi.Type() == 2 {
			tpl := `<!DOCTYPE html>
<title>Password</title>
<form action="" method="post">
	<input type="hidden" name="p" value="%s">
	<label>Password:<input type="password" name="pwd" placeholder="Password"></label><br/><br/>
	<input type="submit" value="Submit">
</form>`
			Fprintf(w, tpl, r.FormValue("p"))
		}
	}
	if fi.Permit(permission.WritePrivate) {
		loginPage(w, r)
	}

	return -2
}