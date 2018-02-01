package session

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//--------------
//--- writer ---
//--------------

// WriteToResponseWriter 将session数据写入到返回中
func (s *Session) WriteToResponseWriter(w http.ResponseWriter) error {
	expiry := s.GetExpiry()
	// 如果设置了闲置时间
	j, err := encodeToJSON(s.id, s.data, s.deadline)
	if err != nil {
		return err
	}
	err = s.Write(j)
	if err != nil {
		return err
	}
	// 如果是客户端存储,要更新token值
	ce, ok := s.store.(clientStore)
	if ok {
		s.token, err = ce.MakeToken(j, expiry)
		if err != nil {
			return err
		}
	}
	// 设置cookie
	cookie := &http.Cookie{
		Name:     s.opts.name,
		Value:    s.token,
		Path:     s.opts.path,
		Domain:   s.opts.domain,
		Secure:   s.opts.secure,
		HttpOnly: s.opts.httpOnly,
	}
	if s.opts.persist == true {
		// Round up expiry time to the nearest second.
		cookie.Expires = time.Unix(expiry.Unix()+1, 0)
		cookie.MaxAge = int(expiry.Sub(time.Now()).Seconds() + 1)
	}

	// 重写存在的cookie
	var set bool
	for i, h := range w.Header()["Set-Cookie"] {
		if strings.HasPrefix(h, fmt.Sprintf("%s=", s.opts.name)) {
			w.Header()["Set-Cookie"][i] = cookie.String()
			set = true
			break
		}
	}
	// 如果不存在，则新生成一个
	if !set {
		http.SetCookie(w, cookie)
	}
	return nil
}

// PutToResponseWriter 存入，存在则替换
func (s *Session) PutToResponseWriter(w http.ResponseWriter, key string, val interface{}) error {
	s.mu.Lock()
	s.data[key] = val
	s.mu.Unlock()
	return s.WriteToResponseWriter(w)
}

// PopFromResponseWriter 移除并返回
func (s *Session) PopFromResponseWriter(w http.ResponseWriter, key string) (interface{}, bool, error) {
	s.mu.Lock()
	v, exists := s.data[key]
	if exists == false {
		s.mu.Unlock()
		return nil, false, nil
	}
	delete(s.data, key)
	s.mu.Unlock()
	err := s.WriteToResponseWriter(w)
	if err != nil {
		return nil, false, err
	}
	return v, true, nil
}

// PutStringToResponseWriter 存储string
func (s *Session) PutStringToResponseWriter(w http.ResponseWriter, key string, val string) error {
	return s.PutToResponseWriter(w, key, val)
}

// PopStringFromResponseWriter 移除并返回
func (s *Session) PopStringFromResponseWriter(w http.ResponseWriter, key string) (string, error) {
	v, exists, err := s.PopFromResponseWriter(w, key)
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

// PutBoolToResponseWriter 存储
func (s *Session) PutBoolToResponseWriter(w http.ResponseWriter, key string, val bool) error {
	return s.PutToResponseWriter(w, key, val)
}

// PopBoolFromResponseWriter 移除并返回
func (s *Session) PopBoolFromResponseWriter(w http.ResponseWriter, key string) (bool, error) {
	v, exists, err := s.PopFromResponseWriter(w, key)
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

// PutIntToResponseWriter 存储
func (s *Session) PutIntToResponseWriter(w http.ResponseWriter, key string, val int) error {
	return s.PutToResponseWriter(w, key, val)
}

// PopIntFromResponseWriter 移除并返回
func (s *Session) PopIntFromResponseWriter(w http.ResponseWriter, key string) (int, error) {
	v, exists, err := s.PopFromResponseWriter(w, key)
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

// PutInt64ToResponseWriter 存储
func (s *Session) PutInt64ToResponseWriter(w http.ResponseWriter, key string, val int64) error {
	return s.PutToResponseWriter(w, key, val)
}

// PopInt64FromResponseWriter 移除并返回
func (s *Session) PopInt64FromResponseWriter(w http.ResponseWriter, key string) (int64, error) {
	v, exists, err := s.PopFromResponseWriter(w, key)
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

// PutFloat64ToResponseWriter 存储
func (s *Session) PutFloat64ToResponseWriter(w http.ResponseWriter, key string, val float64) error {
	return s.PutToResponseWriter(w, key, val)
}

// PopFloat64FromResponseWriter 移除并返回
func (s *Session) PopFloat64FromResponseWriter(w http.ResponseWriter, key string) (float64, error) {
	v, exists, err := s.PopFromResponseWriter(w, key)
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

// PutTimeToResponseWriter 存储
func (s *Session) PutTimeToResponseWriter(w http.ResponseWriter, key string, val time.Time) error {
	return s.PutToResponseWriter(w, key, val)
}

// PopTimeFromResponseWriter 移除并返回
func (s *Session) PopTimeFromResponseWriter(w http.ResponseWriter, key string) (time.Time, error) {
	v, exists, err := s.PopFromResponseWriter(w, key)
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

// PutBytesToResponseWriter 存储
func (s *Session) PutBytesToResponseWriter(w http.ResponseWriter, key string, val []byte) error {
	return s.PutToResponseWriter(w, key, val)
}

// PopBytesFromResponseWriter 移除并返回
func (s *Session) PopBytesFromResponseWriter(w http.ResponseWriter, key string) ([]byte, error) {
	v, exists, err := s.PopFromResponseWriter(w, key)
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

// PutObject 存入，存在则替换
func (s *Session) PutObjectToResponseWriter(w http.ResponseWriter, key string, val interface{}) error {
	if val == nil {
		return errors.New("value must not be nil")
	}
	b, err := gobEncode(val)
	if err != nil {
		return err
	}
	return s.PutBytesToResponseWriter(w, key, b)
}

// PopObject 移除并返回
func (s *Session) PopObjectFromResponseWriter(w http.ResponseWriter, key string, dst interface{}) error {
	b, err := s.PopBytesFromResponseWriter(w, key)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}
	return gobDecode(b, dst)
}
