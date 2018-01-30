package session

import (
	"sync"
)

// Session 会话管理
type Session interface {
	Set(key, value interface{}) error
	Get(key interface{}) interface{}
	Del(key interface{}) interface{}
}

type session struct {
	id             string                      // session 唯一id
	data           map[interface{}]interface{} // 可存储的数据
	idleTime       int64                       // 最大空闲时间
	lastAccessTime int64                       // 最后访问时间
	mu             sync.Mutex                  // 访问锁
}
