package session

import (
	"context"
	"fmt"
	"github.com/alexedwards/scs/stores/cookiestore"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Manager session控制器
type Manager struct {
	store    Store
	opts     *Options
	sessions map[string]*Session // 只是为了更方便的查询session的数据，判别session之间的关系
	mu       sync.Mutex
}

// NewManager 返回session管理器
// 并且伴随生成一个gc任务
func NewManager(store Store, opts ...Option) *Manager {
	options := NewOptions(opts...)
	manager := &Manager{
		store:    store,
		opts:     &options,
		sessions: make(map[string]*Session),
	}
	go manager.RunGC()
	return manager
}

func (m *Manager) RunGC() {
	d := time.Minute * 15
	if m.opts.idleTimeout > 0 {
		d = time.Duration(math.Ceil(m.opts.idleTimeout.Minutes()/2)) * time.Minute
	}
	time.AfterFunc(d, func() {
		m.gc()
		m.RunGC()
	})
}

func (m *Manager) gc() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.sessions {
		if v.TimeOut() {
			// 这里要求所有的存储器自带GC
			//if !m.store.AutoGC(){
			//	v.Destroy()
			//}
			delete(m.sessions, k)
		}
	}
}

// NewSession 创建并且返回一个Session
func (m *Manager) NewSession() (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, err := newSession(m.store, m.opts)
	if err != nil {
		return nil, err
	}
	m.sessions[s.token] = s
	return s, nil
}

//-------------------------
//-- handle http request
//-------------------------

type sessionName string

// Load 从r里加载session
// 1.从request上下文中直接获取
// 2.从cookie中获取token，根据token获取Session
// 3.如果没有则生成一个session
func (m *Manager) Load(r *http.Request) (*Session, error) {
	// 检查上下文中是否存在session信息
	val := r.Context().Value(sessionName(m.opts.name))
	if val != nil {
		s, ok := val.(*Session)
		if !ok {
			return nil, fmt.Errorf("scs: can not assert %T to *Session", val)
		}
		return s, nil
	}
	// 如果上下文中没有，从cokie中获取token,如果获取不到，直接生成
	cookie, err := r.Cookie(m.opts.name)
	if err == http.ErrNoCookie {
		return m.NewSession()
	} else if err != nil {
		return nil, err
	}
	if cookie.Value == "" {
		return m.NewSession()
	}
	token := cookie.Value
	// 根据token从Store中获取数据，如果store里没有，生成一个
	j, found, err := m.store.Find(token)
	if err != nil {
		return nil, err
	}
	if found == false {
		return m.NewSession()
	}
	// 根据数据生成一个session
	data, deadline, err := decodeFromJSON(j)
	if err != nil {
		return nil, err
	}
	s := &Session{
		token:    token,
		data:     data,
		deadline: deadline,
		store:    m.store,
		opts:     m.opts,
	}
	return s, nil
}

func (m *Manager) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 加载一个session
		session, err := m.Load(r)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		expiry := session.GetExpiry()
		// 如果设置了闲置时间
		j, err := encodeToJSON(session.data, session.deadline)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		err = session.Write(j)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if ce, ok := session.store.(clientStore); ok {
			session.token, err = ce.MakeToken(j, expiry)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		// 设置cookie
		cookie := &http.Cookie{
			Name:     session.opts.name,
			Value:    session.token,
			Path:     session.opts.path,
			Domain:   session.opts.domain,
			Secure:   session.opts.secure,
			HttpOnly: session.opts.httpOnly,
		}
		if session.opts.persist == true {
			// Round up expiry time to the nearest second.
			cookie.Expires = time.Unix(expiry.Unix()+1, 0)
			cookie.MaxAge = int(expiry.Sub(time.Now()).Seconds() + 1)
		}

		// 重写存在的cookie
		var set bool
		for i, h := range w.Header()["Set-Cookie"] {
			if strings.HasPrefix(h, fmt.Sprintf("%s=", session.opts.name)) {
				w.Header()["Set-Cookie"][i] = cookie.String()
				set = true
				break
			}
		}
		// 如果不存在，则新生成一个
		if !set {
			http.SetCookie(w, cookie)
		}
		ctx := context.WithValue(r.Context(), sessionName(m.opts.name), session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
