<!-- 简单的session处理机制 -->

### session

>- [CSDN Go实现Session](http://blog.csdn.net/lzy_zhi_yuan/article/details/73127601)
>- [Github Alexedwards/scs](https://github.com/alexedwards/scs)

> - 将cookieStore分类为客户端存储器
> - 在manage中保存session信息,用于manager之间的交互，同时GC机制清理manager中保存的过期session
> - 添加session-id,用于在管理器中查找已存在的session进行返回
> - 添加Finder,用于在管理器中查找符合条件的session


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

	scs "github.com/ipiao/session"
)

var sessionManager = scs.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4")

func main() {
	sessionManager.Option(scs.Persist(true))
	sessionManager.Option(scs.LifeTime(time.Second * 30))

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
	session.WriteToResponseWriter(w)
	sessions := sessionManager.FindSeesion()
	log.Println("GET:", len(sessions))
	sessions1 := sessionManager.FindSeesion(scs.FindByKVEq("message", "Hello world!"))
	log.Println("GET KVEq:", len(sessions1))
	sessions2 := sessionManager.FindSeesion(scs.FindTimeOut())
	log.Println("GET TimeOut:", len(sessions2))
	message, err := session.GetString("message")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	log.Println("GET:", session.GetData())
	io.WriteString(w, message)
}
```