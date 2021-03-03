package session

import (
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
	"via/def"
)

func Update(w http.ResponseWriter, r *http.Request) {
	ses, _ := GetUser(r)
	ses.Values["username"] = def.Username
	ses.Values["password"] = def.Password
	ses.Options.MaxAge = 60 * 60 * 24
	err := ses.Save(r, w)
	if err != nil {
		_, _ = fmt.Fprintln(w, "ERROR session SetUserInfo")
		return
	}
}

func Verify(r *http.Request) bool {
	ses, err := Get(r, def.SessionName)
	if err != nil {
		return false
	}
	username, ok := ses.Values["username"].(string)
	if !ok {
		return false
	}
	return username == def.Username
}

func GetUser(r *http.Request) (*sessions.Session, error) {
	return Get(r, def.SessionName)
}
