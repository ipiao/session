package session

// Handle 操作session,错误内部处理
type Handle func(*Session)

// MakeHandle 将多个变成一个
func MakeHandle(fds ...Handle) Handle {
	return func(s *Session) {
		for _, fd := range fds {
			fd(s)
		}
	}
}

// HandleSetKV 设置键值
func HandleSetKV(key string, value interface{}) Handle {
	return func(s *Session) {
		s.mu.Lock()
		s.data[key] = value
		s.mu.Unlock()
	}
}
