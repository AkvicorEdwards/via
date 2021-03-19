package handler

import (
	"fmt"
	"net/http"
	"via/def"
	"via/record"
	"via/session"
)

// Public
var public = make(map[string]func(http.ResponseWriter, *http.Request))
// Protected by login/password/key
var protected = make(map[string]func(http.ResponseWriter, *http.Request))
// Must login
var private = make(map[string]func(http.ResponseWriter, *http.Request))

func ParsePrefix() {
	// API
	public["/login"] = login
	protected["/file"] = fileGet
	protected["/upload/file"] = fileUpload
	//protected["/update/file"] = fileUpdate
	protected["/update/info/file"] = fileInfoEdit
	protected["/update/info/path"] = pathInfoEdit
	protected["/path"] = pathAdd
	protected["/del/path"] = pathDel
	protected["/del/file"] = fileDel
	private["/search/file"] = fileSearch

	// Page
	public["/page/login"] = loginPage
	protected["/"] = fileIndex
	protected["/page/upload/file"] = fileUploadPage
	protected["/page/update/file"] = fileUpdatePage
	protected["/page/upload/path"] = pathUpload
	protected["/page/update/info/file"] = fileInfoEditPage
	protected["/page/update/info/path"] = pathInfoEditPage
}

type MyHandler struct{}

func (*MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := public[r.URL.Path]; ok {
		h(w, r)
		return
	}
	if h, ok := protected[r.URL.Path]; ok {
		h(w, r)
		return
	}
	if !session.Verify(r) {
		public["/page/login"](w, r)
		return
	}

	if h, ok := private[r.URL.Path]; ok {
		h(w, r)
		return
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	url := r.FormValue("url")
	if len(url) == 0 || url=="/?" || url=="?" {
		url = "/"
	}
	if username == def.Username && password == def.Password {
		session.Update(w, r)
		Fprint(w, TplRedirect(url))
		record.Printf("Login Successful U:[%s]\n", username)
	} else {
		Fprint(w, TplRedirect("history"))
		record.Printf("Login Failed U:[%s] P:[%s]\n", username, password)
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	tpl := `<!DOCTYPE html>
<title>Login</title>
<form action="/login" method="post">
	<input type="hidden" name="url" id="url" value="%s">
	<label>Username:<input type="text" name="username"></label><br/><br/>
	<label>Password:<input type="password" name="password"></label><br/><br/>
	<input type="submit" value="Login">
</form>`

	Fprintf(w, tpl, fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery))
	return
}

func Fprint(w http.ResponseWriter, a ...interface{}) {
	_, _ = fmt.Fprint(w, a...)
}

func Fprintf(w http.ResponseWriter, format string, a ...interface{}) {
	_, _ = fmt.Fprintf(w, format, a...)
}

func Fprintln(w http.ResponseWriter, a ...interface{}) {
	_, _ = fmt.Fprintln(w, a...)
}
