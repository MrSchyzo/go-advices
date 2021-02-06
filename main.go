package main

import (
	j "encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"time"

	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/tidwall/buntdb"
)

// Slip struct
type Slip struct {
	ID     int    `json:"id"`
	Advice string `json:"advice"`
	Date   string `json:"date"`
}

// Message struct
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SlipError struct
type SlipError struct {
	Message Message `json:"message"`
}

// QueryResult struct
type QueryResult struct {
	ResultsAmount string `json:"total_results"`
	Query         string `json:"query"`
	Slips         []Slip `json:"slips"`
}

// AdviceArgs struct
type AdviceArgs struct {
	Topic       string `json:"topic"`
	MaybeAmount *int   `json:"amount,omitempty"`
}

// AdviceReply struct
type AdviceReply struct {
	AdviceList []string `json:"adviceList"`
}

// AdviceGetter interface
type AdviceGetter interface {
	GetAdvicesLimitedFor(topic string, amount int) ([]string, error)
	GetAdvicesFor(topic string) ([]string, error)
}

// AdviceService struct
type AdviceService struct {
	getter AdviceGetter
}

// GiveMeAdvice What is this?
func (a *AdviceService) GiveMeAdvice(args *AdviceArgs, reply *AdviceReply) error {

	var advices []string
	var err error

	if args.MaybeAmount == nil {
		advices, err = a.getter.GetAdvicesFor(args.Topic)
	} else if *(args.MaybeAmount) >= 0 {
		advices, err = a.getter.GetAdvicesLimitedFor(args.Topic, *(args.MaybeAmount))
	} else {
		return errors.New("Cannot accept an amount that is less than 0")
	}

	if err != nil {
		return err
	}

	reply.AdviceList = advices

	return nil
}

// CachedAdviceGetter struct
type CachedAdviceGetter struct {
	db *buntdb.DB
}

func (g *CachedAdviceGetter) GetAdvicesLimitedFor(topic string, amount int) ([]string, error) {
	advices, err := g.GetAdvicesFor(topic)

	if err != nil {
		return nil, err
	}

	return limitSliceTo(advices, amount), nil
}

func (g *CachedAdviceGetter) GetAdvicesFor(topic string) ([]string, error) {
	cached, _ := getFrom(topic, g.db)
	if cached != nil {
		return cached, nil
	}

	advices, err := retrieveAllAdvices(topic)
	if err != nil {
		return nil, err
	}

	stored, err := storeInto(topic, advices, g.db)
	if err != nil {
		return nil, err
	}

	return stored, err
}

func storeInto(key string, content []string, db *buntdb.DB) ([]string, error) {
	value, failure := j.Marshal(content)
	if failure != nil {
		return nil, failure
	}

	err := db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, string(value), &buntdb.SetOptions{Expires: true, TTL: time.Minute * 5})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return content, nil
}

func getFrom(key string, db *buntdb.DB) ([]string, error) {
	var cachedString string
	var cached []string

	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key, false)
		if err != nil {
			return err
		}
		cachedString = val
		return nil
	})

	if err != nil {
		return nil, err
	}

	j.Unmarshal([]byte(cachedString), &cached)
	return cached, nil
}

func retrieveAllAdvices(s string) ([]string, error) {
	resp, errMsg, err := queryForAdviceWith(s)

	if err != nil {
		return nil, err
	}
	if errMsg.Message.Text != "" {
		fmt.Println("Unable to find anything for \"" + s + "\"")
		return make([]string, 0), nil
	}

	ret := make([]string, len(resp.Slips))
	for i := 0; i < len(resp.Slips); i++ {
		ret[i] = resp.Slips[i].Advice
	}

	return ret, nil
}

func limitSliceTo(slice []string, i int) []string {
	if i < 0 {
		return slice
	}

	l := len(slice)

	var amount int
	if l > i {
		amount = i
	} else {
		amount = l
	}

	return slice[0:amount]
}

func getAdvicesFor(s string, db *buntdb.DB) ([]string, error) {
	cached, _ := getFrom(s, db)
	if cached != nil {
		return cached, nil
	}

	advices, err := retrieveAllAdvices(s)
	if err != nil {
		return nil, err
	}

	stored, err := storeInto(s, advices, db)
	if err != nil {
		return nil, err
	}

	return stored, err
}

func queryForAdviceWith(s string) (QueryResult, SlipError, error) {
	fmt.Println("(HTTP call for querying " + s + ")")
	client := &http.Client{}
	defer client.CloseIdleConnections()

	query := url.PathEscape(s)

	req, err := http.NewRequest("GET", "https://api.adviceslip.com/advice/search/"+query, nil)
	if err != nil {
		return QueryResult{}, SlipError{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return QueryResult{}, SlipError{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return QueryResult{}, SlipError{}, err
	}

	var error SlipError
	j.Unmarshal(bodyBytes, &error)
	if error.Message.Text != "" {
		return QueryResult{}, error, nil
	}

	var result QueryResult
	resultError := j.Unmarshal(bodyBytes, &result)
	if resultError == nil {
		return result, SlipError{}, nil
	}

	return QueryResult{}, SlipError{}, resultError
}

func main() {
	fmt.Println("Welcome!")
	db, err := buntdb.Open(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	getter := &AdviceService{&CachedAdviceGetter{db}}
	rpc.Register(getter)

	// Server provide a HTTP transport on /rpc endpoint.
	http.Handle("/rpc", wrap(jsonrpc2.HTTPHandler(nil)))
	lnHTTP, err := net.Listen("tcp", "localhost:10000")
	if err != nil {
		panic(err)
	}
	defer lnHTTP.Close()
	http.Serve(lnHTTP, nil)
}

type customHandler struct {
	handler http.Handler
}

func (c *customHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Header != nil && req.Header.Get("Accept") == "*/*" {
		req.Header.Set("Accept", "application/json")
	}

	c.handler.ServeHTTP(w, req)
}

func wrap(handler http.Handler) http.Handler {
	return &customHandler{handler}
}
