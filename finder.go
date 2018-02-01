package session

// Finder 查找session
type Finder func(*Session) bool

// MakeFinder 将多个变成一个
func MakeFinder(fds ...Finder) Finder {
	return func(s *Session) bool {
		for _, fd := range fds {
			if !fd(s) {
				return false
			}
		}
		return true
	}
}

// FindByKVEq 按键值查找,值相等
func FindByKVEq(key string, value interface{}) Finder {
	return func(s *Session) bool {
		if v, ok := s.data[key]; ok {
			return v == value
		}
		return false
	}
}

// FindTimeOut 查找超时
func FindTimeOut() Finder {
	return func(s *Session) bool {
		return s.TimeOut()
	}
}

// FindByToken 按token查找
func FindByToken(token string) Finder {
	return func(s *Session) bool {
		return s.token == token
	}
}
