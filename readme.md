<!-- 简单的session处理机制 -->

### session

>- [CSDN Go实现Session](http://blog.csdn.net/lzy_zhi_yuan/article/details/73127601)
>- [Github Alexedwards/scs](https://github.com/alexedwards/scs)

> - 将cookieStore分类为客户端存储器
> - 在manage中保存session信息,用于manager之间的交互，同时GC机制清理manager中保存的过期session


### TODO
>支持data查询

###  demo

```go
package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ipiao/session"
)

var sessionManager = session.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4")

func main() {
	sessionManager.Option(session.Persist(true))
	sessionManager.Option(session.LifeTime(time.Second * 30))

	mux := http.NewServeMux()
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/get", getHandler)

	http.ListenAndServe(":4000", mux)
}

func putHandler(w http.ResponseWriter, r *http.Request) {

	session, err := sessionManager.Load(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	err = session.PutToResponseWriter(w, "message", "Hello world!")
	// sessionManager.Write(session, w)
	// session.WriteToResponseWriter(w)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	log.Println("PUT:", session.GetData())
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessionManager.Load(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	message, err := session.GetString("message")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	log.Println("GET:", session.GetData())
	io.WriteString(w, message)
}

```