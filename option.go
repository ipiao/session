package session

import (
	"time"
)

var defaultName = "session"

// Options 大部分继承自http Cookie里字段
type Options struct {
	name        string
	domain      string
	httpOnly    bool
	idleTimeout time.Duration
	lifetime    time.Duration
	path        string
	persist     bool
	secure      bool
}

// NewOptions 新建Options
func NewOptions(opts ...Option) Options {
	var options = Options{
		httpOnly: true,
		path:     "/",
	}
	for _, o := range opts {
		o(&options)
	}
	if options.name == "" {
		options.name = defaultName
	}
	if options.lifetime == 0 {
		options.lifetime = time.Hour * 24
	}
	return options
}

// Option 处理option赋值
type Option func(o *Options)

// Name 设置name
func Name(name string) Option {
	return func(o *Options) {
		o.name = name
	}
}

// Domain 设置domain
func Domain(domain string) Option {
	return func(o *Options) {
		o.domain = domain
	}
}

// HTTPOnly ..
func HTTPOnly(b bool) Option {
	return func(o *Options) {
		o.httpOnly = b
	}
}

// IdleTime ..
func IdleTime(d time.Duration) Option {
	return func(o *Options) {
		o.idleTimeout = d
	}
}

// LifeTime ..
func LifeTime(d time.Duration) Option {
	return func(o *Options) {
		o.lifetime = d
	}
}

// Path ..
func Path(p string) Option {
	return func(o *Options) {
		o.path = p
	}
}

// Persist ..
func Persist(b bool) Option {
	return func(o *Options) {
		o.persist = b
	}
}

// Secure ..
func Secure(b bool) Option {
	return func(o *Options) {
		o.secure = b
	}
}
