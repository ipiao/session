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

	// 废弃，这里要求所有的存储器自带GC
	//// 返回存储器是否会自动GC
	//AutoGC() bool

	// session数据落地,与Loads主要是针对memoryStore
	Dumps() (err error)

	// 加载保存的session数据
	Loads() (bs [][]byte, err error)
}

// 自带token生成方法的，将数据存储在token里面的存储器，如cookie存储器,客户端存储器
type clientStore interface {
	MakeToken(b []byte, expiry time.Time) (token string, err error)
}
