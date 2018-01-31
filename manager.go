package session

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/alexedwards/scs/stores/cookiestore"
)

// Manager is a session manager.
type Manager struct {
	store    Store
	opts     *Options
	sessions map[string]*Session
	mu       sync.Mutex
}

// NewManager 返回session管理器
func NewManager(store Store, opts ...Option) *Manager {
	options := NewOptions(opts...)
	return &Manager{
		store:    store,
		opts:     &options,
		sessions: make(map[string]*Session),
	}
}

// Load 从r里加载session
func (m *Manager) Load(r *http.Request) *Session {
	return load(r, m.store, m.opts)
}

func NewCookieManager(key string) *Manager {
	store := cookiestore.New([]byte(key))
	return NewManager(store)
}

func (m *Manager) Multi(next http.Handler) http.Handler {
	return m.Use(next)
}

func (m *Manager) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := m.Load(r)

		if m.opts.idleTimeout > 0 {
			err := session.Touch(w)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		ctx := context.WithValue(r.Context(), sessionName(m.opts.name), session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type sessionName string

func load(r *http.Request, store Store, opts *Options) *Session {
	// Check to see if there is an already loaded session in the request context.
	val := r.Context().Value(sessionName(opts.name))
	if val != nil {
		s, ok := val.(*Session)
		if !ok {
			return &Session{loadErr: fmt.Errorf("scs: can not assert %T to *Session", val)}
		}
		return s
	}

	cookie, err := r.Cookie(opts.name)
	if err == http.ErrNoCookie {
		return newSession(store, opts)
	} else if err != nil {
		return &Session{loadErr: err}
	}

	if cookie.Value == "" {
		return newSession(store, opts)
	}
	token := cookie.Value

	j, found, err := store.Find(token)
	if err != nil {
		return &Session{loadErr: err}
	}
	if found == false {
		return newSession(store, opts)
	}

	data, deadline, err := decodeFromJSON(j)
	if err != nil {
		return &Session{loadErr: err}
	}

	s := &Session{
		token:    token,
		data:     data,
		deadline: deadline,
		store:    store,
		opts:     opts,
	}

	return s
}
