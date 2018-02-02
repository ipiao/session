package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	scs "github.com/ipiao/session"
	"github.com/ipiao/session/stores/memstore"
)

// var sessionManager = scs.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4")
var sessionManager *scs.Manager

func main() {
	store := memstore.New(time.Second * 30)
	store.SetDumpFile("memdump.dmp")
	sessionManager = scs.NewManager(store)
	go notifySign()

	sessionManager.Option(scs.IdleTime(time.Second * 6))
	sessionManager.Option(scs.Persist(true))
	sessionManager.Option(scs.LifeTime(time.Second * 3000))
	sessionManager.Option(scs.TouchInterval(time.Second * 2))

	mux := http.NewServeMux()
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/get", getHandler)

	http.ListenAndServe(":4000", sessionManager.Use(mux))
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
	// sessionManager.FindSeesion(scs.FindByKVEq("message", "Hello world!"))
	log.Println("PUT:", session.GetData())
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessionManager.Load(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	log.Println("LastAccessTime:", session.LastAccessTime())

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

func notifySign() {
	var sigRecv = make(chan os.Signal, 1)
	sigs := []os.Signal{syscall.SIGINT, syscall.SIGQUIT}
	signal.Notify(sigRecv, sigs...)
	go func() {
		for range sigRecv {
			sessionManager.Close()
			os.Exit(0)
		}
	}()
}
