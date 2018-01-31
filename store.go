package session

import "time"

// Store 存储session
// 存储的实际是Session的data和deadline
type Store interface {
	// 存储session，如果token一致，则更新session,同时改写过期时间
	Save(token string, b []byte, expiry time.Time) (err error)

	// 移除给定token的session并且获取，如果不存在，返回nil
	Delete(token string) (err error)

	// 查找给定token的session
	Find(token string) (b []byte, found bool, err error)

	// // FineAll 查找所有数据，用于分布式/系统重启的session重载
	// FineAll(bs [][]byte, err error)
}

type cookieStore interface {
	MakeToken(b []byte, expiry time.Time) (token string, err error)
}
