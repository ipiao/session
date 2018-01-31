package session

// import (
// 	"bytes"
// 	"crypto/rand"
// 	"encoding/base64"
// 	"encoding/gob"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"sort"
// 	"strconv"
// 	"strings"
// 	"sync"
// 	"time"
// )

// // ErrTypeAssertionFailed 断言错误
// var ErrTypeAssertionFailed = errors.New("type assertion failed")

// // CookieSession 一个会话状态
// type CookieSession struct {
// 	token    string                 // session的Token值，其实也就是sessionID
// 	data     map[string]interface{} // session储存数据
// 	deadline time.Time              // session过期时间
// 	mu       sync.Mutex
// 	opts     *Options
// 	loadErr  error
// 	store    Store
// }

// func newSession(store Store, opts *Options) *CookieSession {
// 	return &CookieSession{
// 		data:     make(map[string]interface{}),
// 		deadline: time.Now().Add(opts.lifetime),
// 		store:    store,
// 		opts:     opts,
// 	}
// }

// // GetString 获取String
// func (s *CookieSession) GetString(key string) (string, error) {
// 	v, exists, err := s.Get(key)
// 	if err != nil {
// 		return "", err
// 	}
// 	if exists == false {
// 		return "", nil
// 	}

// 	str, ok := v.(string)
// 	if ok == false {
// 		return "", ErrTypeAssertionFailed
// 	}
// 	return str, nil
// }

// // PutString 存储string
// func (s *CookieSession) PutString(w http.ResponseWriter, key string, val string) error {
// 	return s.Put(w, key, val)
// }

// // PopString 移除并返回
// func (s *CookieSession) PopString(w http.ResponseWriter, key string) (string, error) {
// 	v, exists, err := s.Pop(w, key)
// 	if err != nil {
// 		return "", err
// 	}
// 	if exists == false {
// 		return "", nil
// 	}

// 	str, ok := v.(string)
// 	if ok == false {
// 		return "", ErrTypeAssertionFailed
// 	}
// 	return str, nil
// }

// // GetBool 获取Bool
// func (s *CookieSession) GetBool(key string) (bool, error) {
// 	v, exists, err := s.Get(key)
// 	if err != nil {
// 		return false, err
// 	}
// 	if exists == false {
// 		return false, nil
// 	}

// 	b, ok := v.(bool)
// 	if ok == false {
// 		return false, ErrTypeAssertionFailed
// 	}
// 	return b, nil
// }

// // PutBool 存入，存在则替换
// func (s *CookieSession) PutBool(w http.ResponseWriter, key string, val bool) error {
// 	return s.Put(w, key, val)
// }

// // PopBool 移除并返回
// func (s *CookieSession) PopBool(w http.ResponseWriter, key string) (bool, error) {
// 	v, exists, err := s.Pop(w, key)
// 	if err != nil {
// 		return false, err
// 	}
// 	if exists == false {
// 		return false, nil
// 	}

// 	b, ok := v.(bool)
// 	if ok == false {
// 		return false, ErrTypeAssertionFailed
// 	}
// 	return b, nil
// }

// // GetInt 获取
// func (s *CookieSession) GetInt(key string) (int, error) {
// 	v, exists, err := s.Get(key)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if exists == false {
// 		return 0, nil
// 	}

// 	switch v := v.(type) {
// 	case int:
// 		return v, nil
// 	case json.Number:
// 		return strconv.Atoi(v.String())
// 	}
// 	return 0, ErrTypeAssertionFailed
// }

// // PutInt 存入，存在则替换
// func (s *CookieSession) PutInt(w http.ResponseWriter, key string, val int) error {
// 	return s.Put(w, key, val)
// }

// // PopInt 移除并返回
// func (s *CookieSession) PopInt(w http.ResponseWriter, key string) (int, error) {
// 	v, exists, err := s.Pop(w, key)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if exists == false {
// 		return 0, nil
// 	}

// 	switch v := v.(type) {
// 	case int:
// 		return v, nil
// 	case json.Number:
// 		return strconv.Atoi(v.String())
// 	}
// 	return 0, ErrTypeAssertionFailed
// }

// // GetInt64 获取
// func (s *CookieSession) GetInt64(key string) (int64, error) {
// 	v, exists, err := s.Get(key)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if exists == false {
// 		return 0, nil
// 	}

// 	switch v := v.(type) {
// 	case int64:
// 		return v, nil
// 	case json.Number:
// 		return v.Int64()
// 	}
// 	return 0, ErrTypeAssertionFailed
// }

// // PutInt64 存入，存在则替换
// func (s *CookieSession) PutInt64(w http.ResponseWriter, key string, val int64) error {
// 	return s.Put(w, key, val)
// }

// // PopInt64 移除并返回
// func (s *CookieSession) PopInt64(w http.ResponseWriter, key string) (int64, error) {
// 	v, exists, err := s.Pop(w, key)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if exists == false {
// 		return 0, nil
// 	}

// 	switch v := v.(type) {
// 	case int64:
// 		return v, nil
// 	case json.Number:
// 		return v.Int64()
// 	}
// 	return 0, ErrTypeAssertionFailed
// }

// // GetFloat64 获取
// func (s *CookieSession) GetFloat64(key string) (float64, error) {
// 	v, exists, err := s.Get(key)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if exists == false {
// 		return 0, nil
// 	}

// 	switch v := v.(type) {
// 	case float64:
// 		return v, nil
// 	case json.Number:
// 		return v.Float64()
// 	}
// 	return 0, ErrTypeAssertionFailed
// }

// // PutFloat64 存入，存在则替换
// func (s *CookieSession) PutFloat64(w http.ResponseWriter, key string, val float64) error {
// 	return s.Put(w, key, val)
// }

// // PopFloat64 移除并返回
// func (s *CookieSession) PopFloat64(w http.ResponseWriter, key string) (float64, error) {
// 	v, exists, err := s.Pop(w, key)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if exists == false {
// 		return 0, nil
// 	}

// 	switch v := v.(type) {
// 	case float64:
// 		return v, nil
// 	case json.Number:
// 		return v.Float64()
// 	}
// 	return 0, ErrTypeAssertionFailed
// }

// // GetTime 获取
// func (s *CookieSession) GetTime(key string) (time.Time, error) {
// 	v, exists, err := s.Get(key)
// 	if err != nil {
// 		return time.Time{}, err
// 	}
// 	if exists == false {
// 		return time.Time{}, nil
// 	}

// 	switch v := v.(type) {
// 	case time.Time:
// 		return v, nil
// 	case string:
// 		return time.Parse(time.RFC3339, v)
// 	}
// 	return time.Time{}, ErrTypeAssertionFailed
// }

// // PutTime 存入，存在则替换
// func (s *CookieSession) PutTime(w http.ResponseWriter, key string, val time.Time) error {
// 	return s.Put(w, key, val)
// }

// // PopTime 移除并返回
// func (s *CookieSession) PopTime(w http.ResponseWriter, key string) (time.Time, error) {
// 	v, exists, err := s.Pop(w, key)
// 	if err != nil {
// 		return time.Time{}, err
// 	}
// 	if exists == false {
// 		return time.Time{}, nil
// 	}

// 	switch v := v.(type) {
// 	case time.Time:
// 		return v, nil
// 	case string:
// 		return time.Parse(time.RFC3339, v)
// 	}
// 	return time.Time{}, ErrTypeAssertionFailed
// }

// // GetBytes 获取
// func (s *CookieSession) GetBytes(key string) ([]byte, error) {
// 	v, exists, err := s.Get(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if exists == false {
// 		return nil, nil
// 	}

// 	switch v := v.(type) {
// 	case []byte:
// 		return v, nil
// 	case string:
// 		return base64.StdEncoding.DecodeString(v)
// 	}
// 	return nil, ErrTypeAssertionFailed
// }

// // PutBytes 存入，存在则替换
// func (s *CookieSession) PutBytes(w http.ResponseWriter, key string, val []byte) error {
// 	if val == nil {
// 		return errors.New("value must not be nil")
// 	}

// 	return s.Put(w, key, val)
// }

// // PopBytes 移除并返回
// func (s *CookieSession) PopBytes(w http.ResponseWriter, key string) ([]byte, error) {
// 	v, exists, err := s.Pop(w, key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if exists == false {
// 		return nil, nil
// 	}

// 	switch v := v.(type) {
// 	case []byte:
// 		return v, nil
// 	case string:
// 		return base64.StdEncoding.DecodeString(v)
// 	}
// 	return nil, ErrTypeAssertionFailed
// }

// // GetObject 获取
// func (s *CookieSession) GetObject(key string, dst interface{}) error {
// 	b, err := s.GetBytes(key)
// 	if err != nil {
// 		return err
// 	}
// 	if b == nil {
// 		return nil
// 	}

// 	return gobDecode(b, dst)
// }

// // PutObject 存入，存在则替换
// func (s *CookieSession) PutObject(w http.ResponseWriter, key string, val interface{}) error {
// 	if val == nil {
// 		return errors.New("value must not be nil")
// 	}

// 	b, err := gobEncode(val)
// 	if err != nil {
// 		return err
// 	}

// 	return s.PutBytes(w, key, b)
// }

// // PopObject 移除并返回
// func (s *CookieSession) PopObject(w http.ResponseWriter, key string, dst interface{}) error {
// 	b, err := s.PopBytes(w, key)
// 	if err != nil {
// 		return err
// 	}
// 	if b == nil {
// 		return nil
// 	}

// 	return gobDecode(b, dst)
// }

// // Keys 返回所有的键
// func (s *CookieSession) Keys() ([]string, error) {
// 	if s.loadErr != nil {
// 		return nil, s.loadErr
// 	}

// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	keys := make([]string, len(s.data))
// 	i := 0
// 	for k := range s.data {
// 		keys[i] = k
// 		i++
// 	}

// 	sort.Strings(keys)
// 	return keys, nil
// }

// // Exists 是否存在给定键的数据
// func (s *CookieSession) Exists(key string) (bool, error) {
// 	if s.loadErr != nil {
// 		return false, s.loadErr
// 	}

// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	_, exists := s.data[key]
// 	return exists, nil
// }

// // Remove 移除给定键数据
// func (s *CookieSession) Remove(w http.ResponseWriter, key string) error {
// 	if s.loadErr != nil {
// 		return s.loadErr
// 	}

// 	s.mu.Lock()

// 	_, exists := s.data[key]
// 	if exists == false {
// 		s.mu.Unlock()
// 		return nil
// 	}

// 	delete(s.data, key)
// 	s.mu.Unlock()

// 	return s.write(w)
// }

// // Clear 清楚所有的数据
// func (s *CookieSession) Clear(w http.ResponseWriter) error {
// 	if s.loadErr != nil {
// 		return s.loadErr
// 	}

// 	s.mu.Lock()

// 	if len(s.data) == 0 {
// 		s.mu.Unlock()
// 		return nil
// 	}

// 	for key := range s.data {
// 		delete(s.data, key)
// 	}
// 	s.mu.Unlock()

// 	return s.write(w)
// }

// // RenewToken 重新创建token
// func (s *CookieSession) RenewToken(w http.ResponseWriter) error {
// 	if s.loadErr != nil {
// 		return s.loadErr
// 	}

// 	s.mu.Lock()

// 	err := s.store.Delete(s.token)
// 	if err != nil {
// 		s.mu.Unlock()
// 		return err
// 	}

// 	token, err := generateToken()
// 	if err != nil {
// 		s.mu.Unlock()
// 		return err
// 	}

// 	s.token = token
// 	s.deadline = time.Now().Add(s.opts.lifetime)
// 	s.mu.Unlock()

// 	return s.write(w)
// }

// // Destroy 摧毁session
// func (s *CookieSession) Destroy(w http.ResponseWriter) error {
// 	if s.loadErr != nil {
// 		return s.loadErr
// 	}

// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	err := s.store.Delete(s.token)
// 	if err != nil {
// 		return err
// 	}

// 	s.token = ""
// 	for key := range s.data {
// 		delete(s.data, key)
// 	}

// 	cookie := &http.Cookie{
// 		Name:     s.opts.name,
// 		Value:    "",
// 		Path:     s.opts.path,
// 		Domain:   s.opts.domain,
// 		Secure:   s.opts.secure,
// 		HttpOnly: s.opts.httpOnly,
// 		Expires:  time.Unix(1, 0),
// 		MaxAge:   -1,
// 	}
// 	http.SetCookie(w, cookie)

// 	return nil
// }

// // Touch 相当于刷新一下时间
// func (s *CookieSession) Touch(w http.ResponseWriter) error {
// 	if s.loadErr != nil {
// 		return s.loadErr
// 	}

// 	return s.write(w)
// }

// //-------------------------

// // Get 获取key对应的值
// func (s *CookieSession) Get(key string) (interface{}, bool, error) {
// 	if s.loadErr != nil {
// 		return nil, false, s.loadErr
// 	}

// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	v, exists := s.data[key]
// 	return v, exists, nil
// }

// // Put 存入，存在则替换
// func (s *CookieSession) Put(w http.ResponseWriter, key string, val interface{}) error {
// 	if s.loadErr != nil {
// 		return s.loadErr
// 	}

// 	s.mu.Lock()
// 	s.data[key] = val
// 	s.mu.Unlock()

// 	return s.write(w)
// }

// // Pop 移除并返回
// func (s *CookieSession) Pop(w http.ResponseWriter, key string) (interface{}, bool, error) {
// 	if s.loadErr != nil {
// 		return nil, false, s.loadErr
// 	}
// 	s.mu.Lock()

// 	v, exists := s.data[key]
// 	if exists == false {
// 		s.mu.Unlock()
// 		return nil, false, nil
// 	}
// 	delete(s.data, key)
// 	s.mu.Unlock()

// 	err := s.write(w)
// 	if err != nil {
// 		return nil, false, err
// 	}
// 	return v, true, nil
// }

// // 写入并更改相应的数据
// func (s *CookieSession) write(w http.ResponseWriter) error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	j, err := encodeToJSON(s.data, s.deadline)
// 	if err != nil {
// 		return err
// 	}

// 	expiry := s.deadline
// 	if s.opts.idleTimeout > 0 {
// 		ie := time.Now().Add(s.opts.idleTimeout)
// 		if ie.Before(expiry) {
// 			expiry = ie
// 		}
// 	}

// 	if ce, ok := s.store.(cookieStore); ok {
// 		s.token, err = ce.MakeToken(j, expiry)
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		if s.token == "" {
// 			s.token, err = generateToken()
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		err = s.store.Save(s.token, j, expiry)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	cookie := &http.Cookie{
// 		Name:     s.opts.name,
// 		Value:    s.token,
// 		Path:     s.opts.path,
// 		Domain:   s.opts.domain,
// 		Secure:   s.opts.secure,
// 		HttpOnly: s.opts.httpOnly,
// 	}
// 	if s.opts.persist == true {
// 		// Round up expiry time to the nearest second.
// 		cookie.Expires = time.Unix(expiry.Unix()+1, 0)
// 		cookie.MaxAge = int(expiry.Sub(time.Now()).Seconds() + 1)
// 	}

// 	// Overwrite any existing cookie header for the session...
// 	var set bool
// 	for i, h := range w.Header()["Set-Cookie"] {
// 		if strings.HasPrefix(h, fmt.Sprintf("%s=", s.opts.name)) {
// 			w.Header()["Set-Cookie"][i] = cookie.String()
// 			set = true
// 			break
// 		}
// 	}
// 	// Or set a new one if necessary.
// 	if !set {
// 		http.SetCookie(w, cookie)
// 	}

// 	return nil
// }

// // 生成token
// func generateToken() (string, error) {
// 	b := make([]byte, 32)
// 	_, err := rand.Read(b)
// 	if err != nil {
// 		return "", err
// 	}
// 	return base64.RawURLEncoding.EncodeToString(b), nil
// }

// func gobEncode(v interface{}) ([]byte, error) {
// 	buf := new(bytes.Buffer)
// 	err := gob.NewEncoder(buf).Encode(v)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

// func gobDecode(b []byte, dst interface{}) error {
// 	buf := bytes.NewBuffer(b)
// 	return gob.NewDecoder(buf).Decode(dst)
// }

// func encodeToJSON(data map[string]interface{}, deadline time.Time) ([]byte, error) {
// 	return json.Marshal(&struct {
// 		Data     map[string]interface{} `json:"data"`
// 		Deadline int64                  `json:"deadline"`
// 	}{
// 		Data:     data,
// 		Deadline: deadline.UnixNano(),
// 	})
// }

// func decodeFromJSON(j []byte) (map[string]interface{}, time.Time, error) {
// 	aux := struct {
// 		Data     map[string]interface{} `json:"data"`
// 		Deadline int64                  `json:"deadline"`
// 	}{}

// 	dec := json.NewDecoder(bytes.NewReader(j))
// 	dec.UseNumber()
// 	err := dec.Decode(&aux)
// 	if err != nil {
// 		return nil, time.Time{}, err
// 	}
// 	return aux.Data, time.Unix(0, aux.Deadline), nil
// }

// type sessionName string

// func load(r *http.Request, store Store, opts *Options) *CookieSession {
// 	// Check to see if there is an already loaded session in the request context.
// 	val := r.Context().Value(sessionName(opts.name))
// 	if val != nil {
// 		s, ok := val.(*CookieSession)
// 		if !ok {
// 			return &CookieSession{loadErr: fmt.Errorf("scs: can not assert %T to *CookieSession", val)}
// 		}
// 		return s
// 	}

// 	cookie, err := r.Cookie(opts.name)
// 	if err == http.ErrNoCookie {
// 		return newSession(store, opts)
// 	} else if err != nil {
// 		return &CookieSession{loadErr: err}
// 	}

// 	if cookie.Value == "" {
// 		return newSession(store, opts)
// 	}
// 	token := cookie.Value

// 	j, found, err := store.Find(token)
// 	if err != nil {
// 		return &CookieSession{loadErr: err}
// 	}
// 	if found == false {
// 		return newSession(store, opts)
// 	}

// 	data, deadline, err := decodeFromJSON(j)
// 	if err != nil {
// 		return &CookieSession{loadErr: err}
// 	}

// 	s := &CookieSession{
// 		token:    token,
// 		data:     data,
// 		deadline: deadline,
// 		store:    store,
// 		opts:     opts,
// 	}

// 	return s
// }
