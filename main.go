package main

import (
	"flag"
	"fmt"
	"log"
	"myCache/mycach"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Mycache server port")
	flag.BoolVar(&api, "api", false, "Start a papi server?")
	flag.Parse()
	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	group := createGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], []string(addrs), group)
}

func createGroup() *myCach.Group {
	return myCach.NewGroup("scores", 2<<10, myCach.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))
}

func startAPIServer(apiAddr string, group *myCach.Group) {
	http.Handle("/api", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")
		view, err := group.Get(key)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(view.ByteSlice())
	}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func startCacheServer(addr string, addrs []string, group *myCach.Group) {
	peers := myCach.NewHTTPPool(addr)
	peers.Set(addrs...)
	group.RegisterPeers(peers)
	log.Println("myCache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}
