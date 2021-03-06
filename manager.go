package session

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/ipiao/session/stores/cookiestore"
)

// Manager session控制器
type Manager struct {
	store    Store
	opts     Options
	sessions map[string]*Session // 只是为了更方便的查询session的数据，判别session之间的关系
	mu       sync.Mutex
}

// NewManager 返回session管理器
// 并且伴随生成一个gc任务
func NewManager(store Store, opts ...Option) *Manager {
	options := NewOptions(opts...)
	manager := &Manager{
		store:    store,
		opts:     options,
		sessions: make(map[string]*Session),
	}
	// 从store中加载sessions
	bs, err := store.Loads()
	if err != nil {
		log.Printf("can not load sessions from store, error occur:%v", err)
	}
	for _, b := range bs {
		id, data, expiry, err := decodeFromJSON(b)
		if err != nil {
			log.Printf("can decode from JSON:%v", err)
			continue
		}
		if expiry.After(time.Now()) {
			s := Session{
				id:       id,
				token:    id,
				data:     data,
				deadline: expiry,
				mu:       sync.Mutex{},
				opts:     options,
				store:    store,
			}
			manager.sessions[s.token] = &s
		}
	}
	go manager.RunGC()
	return manager
}

// NewCookieManager 返回cookie-session管理器
// 客户端存储
func NewCookieManager(key string, opts ...Option) *Manager {
	store := cookiestore.New([]byte(key))
	return NewManager(store)
}

// Option ...
func (m *Manager) Option(opts ...Option) {
	for _, o := range opts {
		o(&m.opts)
	}
}

// RunGC 运行gc,简单设定间隔
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

// FindSeesion 查找session
func (m *Manager) FindSeesion(fds ...Finder) []*Session {
	fd := MakeFinder(fds...)
	var ret = make([]*Session, 0)
	for _, s := range m.sessions {
		if fd(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

// FindHandleSeesion 查找并处理session
func (m *Manager) FindHandleSeesion(fd Finder, hd Handle) []*Session {
	var ret = make([]*Session, 0)
	for _, s := range m.sessions {
		if fd(s) {
			hd(s)
			ret = append(ret, s)
		}
	}
	return ret
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

// Stat 状态
func (m *Manager) Stat() {
	// TODO 返回 manager状态
}

// Close 关闭
func (m *Manager) Close() error {
	return m.store.Dumps()
}

//-------------------------
//-- handle http request
//-------------------------

type sessionName string

// Load 选择manager中存在时候，从中获取
func (m *Manager) Load(r *http.Request) (*Session, error) {
	return m.load(r, true)
}

// LoadIM 不从manager中查找获取，从中获取
// Load ingore manager
func (m *Manager) LoadIM(r *http.Request) (*Session, error) {
	return m.load(r, false)
}

// load 从r里加载session
// 1.从request上下文中直接获取
// 2.从cookie中获取token，根据token获取Session
// 3.从manager中获取
// 4.如果没有则生成一个session
func (m *Manager) load(r *http.Request, queryManager bool) (*Session, error) {
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
		log.Println("newsession:", err)
		return m.NewSession()
	} else if err != nil {
		return nil, err
	}
	if cookie.Value == "" {
		log.Println("newsession cookie.Value is empty ")
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
	id, data, deadline, err := decodeFromJSON(j)
	if err != nil {
		return nil, err
	}
	if queryManager {
		ss := m.FindSeesion(FindByID(id), FindTimeIn())
		if len(ss) == 1 {
			return ss[0], nil
		}
	}
	s := &Session{
		id:       id,
		token:    token,
		data:     data,
		deadline: deadline,
		store:    m.store,
		opts:     m.opts,
	}
	return s, nil
}

// Write 写入数据
func (m *Manager) Write(session *Session, w http.ResponseWriter) error {
	return session.WriteToResponseWriter(w)
}

// Use 用作中间件，作为示例，具体使用根据业务场景而定
func (m *Manager) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 加载一个session
		session, err := m.Load(r)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if session.MayTouch() {
			err = session.WriteToResponseWriter(w)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		ctx := context.WithValue(r.Context(), sessionName(session.opts.name), session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
