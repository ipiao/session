package session

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"
)

// ErrTypeAssertionFailed 断言错误
var ErrTypeAssertionFailed = errors.New("type assertion failed")

// Session 一个会话状态
type Session struct {
	id             string                 // session的id值，一般情况下就是token
	token          string                 // session的Token值，其实也就是sessionID
	data           map[string]interface{} // session储存数据
	deadline       time.Time              // session过期时间
	lastAccessTime time.Time
	mu             sync.Mutex
	opts           Options
	store          Store
}

// newSession 返回一个默认的Session
func newSession(store Store, opts Options) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	s := &Session{
		id:       token,
		data:     make(map[string]interface{}),
		deadline: time.Now().Add(opts.lifetime),
		store:    store,
		opts:     opts,
		token:    token,
	}
	return s, nil
}

// GetID 获取sessionID
func (s *Session) GetID() string {
	return s.id
}

// GetToken 获取sessionToken
func (s *Session) GetToken() string {
	return s.token
}

// GetData 获取session Data
func (s *Session) GetData() map[string]interface{} {
	return s.data
}

// GetExpiry 获取过期时间点
func (s *Session) GetExpiry() time.Time {
	expiry := s.deadline
	if s.opts.idleTimeout > 0 {
		ie := time.Now().Add(s.opts.idleTimeout)
		if ie.Before(expiry) {
			expiry = ie
		}
	}
	return expiry
}

// LastAccessTime 获取上次时间
func (s *Session) LastAccessTime() time.Time {
	return s.lastAccessTime
}

// TimeOut 判断session是否过期
// 1.验证过期时间
// 2.如果未过期，到数据库中查找，如果存不存在
func (s *Session) TimeOut() bool {
	if time.Now().Before(s.deadline) {
		//_, found, _ := s.store.Find(s.token)
		//if !found {
		//	return true
		//}
		return false
	}
	return true
}

// MayTouch 建议刷新
func (s *Session) MayTouch() bool {
	return s.opts.idleTimeout > 0 && s.lastAccessTime.Add(s.opts.touchInterval).Before(time.Now()) || s.lastAccessTime.IsZero()
}

// GetString 获取String
func (s *Session) GetString(key string) (string, error) {
	v, exists, err := s.Get(key)
	if err != nil {
		return "", err
	}
	if exists == false {
		return "", nil
	}

	str, ok := v.(string)
	if ok == false {
		return "", ErrTypeAssertionFailed
	}
	return str, nil
}

// PutString 存储string
func (s *Session) PutString(key string, val string) error {
	return s.Put(key, val)
}

// PopString 移除并返回
func (s *Session) PopString(key string) (string, error) {
	v, exists, err := s.Pop(key)
	if err != nil {
		return "", err
	}
	if exists == false {
		return "", nil
	}

	str, ok := v.(string)
	if ok == false {
		return "", ErrTypeAssertionFailed
	}
	return str, nil
}

// GetBool 获取Bool
func (s *Session) GetBool(key string) (bool, error) {
	v, exists, err := s.Get(key)
	if err != nil {
		return false, err
	}
	if exists == false {
		return false, nil
	}

	b, ok := v.(bool)
	if ok == false {
		return false, ErrTypeAssertionFailed
	}
	return b, nil
}

// PutBool 存入，存在则替换
func (s *Session) PutBool(key string, val bool) error {
	return s.Put(key, val)
}

// PopBool 移除并返回
func (s *Session) PopBool(key string) (bool, error) {
	v, exists, err := s.Pop(key)
	if err != nil {
		return false, err
	}
	if exists == false {
		return false, nil
	}

	b, ok := v.(bool)
	if ok == false {
		return false, ErrTypeAssertionFailed
	}
	return b, nil
}

// GetInt 获取
func (s *Session) GetInt(key string) (int, error) {
	v, exists, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case int:
		return v, nil
	case json.Number:
		return strconv.Atoi(v.String())
	}
	return 0, ErrTypeAssertionFailed
}

// PutInt 存入，存在则替换
func (s *Session) PutInt(key string, val int) error {
	return s.Put(key, val)
}

// PopInt 移除并返回
func (s *Session) PopInt(key string) (int, error) {
	v, exists, err := s.Pop(key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case int:
		return v, nil
	case json.Number:
		return strconv.Atoi(v.String())
	}
	return 0, ErrTypeAssertionFailed
}

// GetInt64 获取
func (s *Session) GetInt64(key string) (int64, error) {
	v, exists, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case int64:
		return v, nil
	case json.Number:
		return v.Int64()
	}
	return 0, ErrTypeAssertionFailed
}

// PutInt64 存入，存在则替换
func (s *Session) PutInt64(key string, val int64) error {
	return s.Put(key, val)
}

// PopInt64 移除并返回
func (s *Session) PopInt64(key string) (int64, error) {
	v, exists, err := s.Pop(key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case int64:
		return v, nil
	case json.Number:
		return v.Int64()
	}
	return 0, ErrTypeAssertionFailed
}

// GetFloat64 获取
func (s *Session) GetFloat64(key string) (float64, error) {
	v, exists, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case float64:
		return v, nil
	case json.Number:
		return v.Float64()
	}
	return 0, ErrTypeAssertionFailed
}

// PutFloat64 存入，存在则替换
func (s *Session) PutFloat64(key string, val float64) error {
	return s.Put(key, val)
}

// PopFloat64 移除并返回
func (s *Session) PopFloat64(key string) (float64, error) {
	v, exists, err := s.Pop(key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case float64:
		return v, nil
	case json.Number:
		return v.Float64()
	}
	return 0, ErrTypeAssertionFailed
}

// GetTime 获取
func (s *Session) GetTime(key string) (time.Time, error) {
	v, exists, err := s.Get(key)
	if err != nil {
		return time.Time{}, err
	}
	if exists == false {
		return time.Time{}, nil
	}

	switch v := v.(type) {
	case time.Time:
		return v, nil
	case string:
		return time.Parse(time.RFC3339, v)
	}
	return time.Time{}, ErrTypeAssertionFailed
}

// PutTime 存入，存在则替换
func (s *Session) PutTime(key string, val time.Time) error {
	return s.Put(key, val)
}

// PopTime 移除并返回
func (s *Session) PopTime(key string) (time.Time, error) {
	v, exists, err := s.Pop(key)
	if err != nil {
		return time.Time{}, err
	}
	if exists == false {
		return time.Time{}, nil
	}

	switch v := v.(type) {
	case time.Time:
		return v, nil
	case string:
		return time.Parse(time.RFC3339, v)
	}
	return time.Time{}, ErrTypeAssertionFailed
}

// GetBytes 获取
func (s *Session) GetBytes(key string) ([]byte, error) {
	v, exists, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	if exists == false {
		return nil, nil
	}

	switch v := v.(type) {
	case []byte:
		return v, nil
	case string:
		return base64.StdEncoding.DecodeString(v)
	}
	return nil, ErrTypeAssertionFailed
}

// PutBytes 存入，存在则替换
func (s *Session) PutBytes(key string, val []byte) error {
	if val == nil {
		return errors.New("value must not be nil")
	}
	return s.Put(key, val)
}

// PopBytes 移除并返回
func (s *Session) PopBytes(key string) ([]byte, error) {
	v, exists, err := s.Pop(key)
	if err != nil {
		return nil, err
	}
	if exists == false {
		return nil, nil
	}

	switch v := v.(type) {
	case []byte:
		return v, nil
	case string:
		return base64.StdEncoding.DecodeString(v)
	}
	return nil, ErrTypeAssertionFailed
}

// GetObject 获取
func (s *Session) GetObject(key string, dst interface{}) error {
	b, err := s.GetBytes(key)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}

	return gobDecode(b, dst)
}

// PutObject 存入，存在则替换
func (s *Session) PutObject(key string, val interface{}) error {
	if val == nil {
		return errors.New("value must not be nil")
	}
	b, err := gobEncode(val)
	if err != nil {
		return err
	}
	return s.PutBytes(key, b)
}

// PopObject 移除并返回
func (s *Session) PopObject(key string, dst interface{}) error {
	b, err := s.PopBytes(key)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}
	return gobDecode(b, dst)
}

// Keys 返回所有的键
func (s *Session) Keys() ([]string, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	keys := make([]string, len(s.data))
	i := 0
	for k := range s.data {
		keys[i] = k
		i++
	}

	sort.Strings(keys)
	return keys, nil
}

// Exists 是否存在给定键的数据
func (s *Session) Exists(key string) (bool, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.data[key]
	return exists, nil
}

// Remove 移除给定键数据
func (s *Session) Remove(key string) error {

	s.mu.Lock()

	_, exists := s.data[key]
	if exists == false {
		s.mu.Unlock()
		return nil
	}

	delete(s.data, key)
	s.mu.Unlock()

	return s.write()
}

// Clear 清楚所有的数据
func (s *Session) Clear() error {

	s.mu.Lock()

	if len(s.data) == 0 {
		s.mu.Unlock()
		return nil
	}

	for key := range s.data {
		delete(s.data, key)
	}
	s.mu.Unlock()

	return s.write()
}

// // RenewToken 重建token
// func (s *Session) RenewToken() error {
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

// 	return s.write()
// }

// Destroy 摧毁session
func (s *Session) Destroy() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.store.Delete(s.token)
	if err != nil {
		return err
	}
	s.token = ""
	s.id = ""
	for key := range s.data {
		delete(s.data, key)
	}
	return nil
}

// Write 相当于刷新一下时间
func (s *Session) Write(bs []byte) error {
	return s.write(bs)
}

//-------------------------

// Get 获取key对应的值
// err:如果将从store中获取值，将会有错误返回
func (s *Session) Get(key string) (interface{}, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, exists := s.data[key]
	return v, exists, nil
}

// Put 存入，存在则替换
func (s *Session) Put(key string, val interface{}) error {
	s.mu.Lock()
	s.data[key] = val
	s.mu.Unlock()
	return s.write()
}

// Pop 移除并返回
func (s *Session) Pop(key string) (interface{}, bool, error) {
	s.mu.Lock()

	v, exists := s.data[key]
	if exists == false {
		s.mu.Unlock()
		return nil, false, nil
	}
	delete(s.data, key)
	s.mu.Unlock()

	err := s.write()
	if err != nil {
		return nil, false, err
	}
	return v, true, nil
}

// 写入并更改相应的数据
func (s *Session) write(bs ...[]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastAccessTime = time.Now()
	if len(s.token) == 0 {
		return errors.New("scs: token is empty,can not write")
	}

	var j []byte
	var err error
	if len(bs) > 0 {
		j = bs[0]
	} else {
		j, err = encodeToJSON(s.id, s.data, s.deadline)
		if err != nil {
			return err
		}
	}
	expiry := s.GetExpiry()
	err = s.store.Save(s.token, j, expiry)
	if err != nil {
		return err
	}
	return nil
}

// 生成token
func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func gobEncode(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecode(b []byte, dst interface{}) error {
	buf := bytes.NewBuffer(b)
	return gob.NewDecoder(buf).Decode(dst)
}

func encodeToJSON(id string, data map[string]interface{}, deadline time.Time) ([]byte, error) {
	return json.Marshal(&struct {
		Data     map[string]interface{} `json:"data"`
		Deadline int64                  `json:"deadline"`
		ID       string                 `json:"id"`
	}{
		Data:     data,
		Deadline: deadline.UnixNano(),
		ID:       id,
	})
}

func decodeFromJSON(j []byte) (string, map[string]interface{}, time.Time, error) {
	aux := struct {
		Data     map[string]interface{} `json:"data"`
		Deadline int64                  `json:"deadline"`
		ID       string                 `json:"id"`
	}{}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.UseNumber()
	err := dec.Decode(&aux)
	if err != nil {
		return "", nil, time.Time{}, err
	}
	return aux.ID, aux.Data, time.Unix(0, aux.Deadline), nil
}
