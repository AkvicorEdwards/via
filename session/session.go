package session

import (
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"via/def"
)

// Note: Don't store your key in your source code. Pass it via an
// environmental variable, or flag (or both), and don't accidentally commit it
// alongside your code. Ensure your key is sufficiently random - i.e. use Go's
// crypto/rand or securecookie.GenerateRandomKey(32) and persist the result.
// []byte(os.Getenv("SESSION_KEY"))
var store = func() (store *sessions.CookieStore) {
	store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	store.Options.Domain = def.SessionDomain
	store.Options.Path = def.SessionPath
	return store
}()

func Get(r *http.Request, name string) (*sessions.Session, error) {
	return store.Get(r, name)
}

func New(r *http.Request, name string) (*sessions.Session, error) {
	return store.New(r, name)
}

func Save(w http.ResponseWriter, r *http.Request, session *sessions.Session) error {
	return store.Save(r, w, session)
}

func MaxAge(age int) {
	store.MaxAge(age)
}
