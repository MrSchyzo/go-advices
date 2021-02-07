package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/tidwall/buntdb"
)

func poorManInjectDependencies(db *buntdb.DB) (*AdviceService, error) {

	mapping := &SimpleAdviceMapping{}
	limiter := &SimpleAdviceLimiter{}
	client := &http.Client{}
	query := &RESTAdviceQuery{*client}
	retriever := &SimpleAdviceRetriever{query, mapping}
	cache := &InMemoryCacheForAdvices{db}
	getter := &CachedAdviceGetter{cache, retriever, limiter}
	service := &AdviceService{getter}

	return service, nil
}

type customHandler struct {
	handler http.Handler
}

//This handler makes sure that the HTTPHandler is able to accept Accept header set to "*/*"
func (c *customHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Header != nil && req.Header.Get("Accept") == "*/*" {
		req.Header.Set("Accept", "application/json")
	}

	c.handler.ServeHTTP(w, req)
}

/**
This is a sort of decorator pattern for the jsonrpc2.HTTPHandler
*/
func wrap(handler http.Handler) http.Handler {
	return &customHandler{handler}
}

func main() {
	fmt.Println("Running application...")
	db, err := buntdb.Open(":memory:")
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer db.Close()

	getter, err := poorManInjectDependencies(db)
	if err != nil {
		log.Fatalln(err)
		return
	}

	rpc.Register(getter)

	http.Handle("/rpc", wrap(jsonrpc2.HTTPHandler(nil)))
	lnHTTP, err := net.Listen("tcp", "0.0.0.0:10000")
	if err != nil {
		panic(err)
	}
	defer lnHTTP.Close()
	http.Serve(lnHTTP, nil)
}
