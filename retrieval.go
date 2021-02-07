package main

import (
	j "encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// AdviceRetriever interface
type AdviceRetriever interface {
	RetrieveForTopic(topic string) ([]string, error)
}

// SimpleAdviceRetriever struct is the simplest implementation of the interface above
type SimpleAdviceRetriever struct {
	adviceQuery   AdviceQuery
	adviceMapping AdviceMapping
}

// RetrieveForTopic function returns all the advices' text for the given topic
func (r *SimpleAdviceRetriever) RetrieveForTopic(topic string) ([]string, error) {
	resp, errMsg, err := r.adviceQuery.GetByTopic(topic)

	if err != nil {
		return nil, err
	}
	if errMsg.Message.Text != "" {
		fmt.Println("Unable to find anything for \"" + topic + "\"")
		return make([]string, 0), nil
	}

	ret := make([]string, len(resp.Slips))
	for i := 0; i < len(resp.Slips); i++ {
		ret[i] = resp.Slips[i].Advice
	}

	return ret, nil
}

// AdviceQuery interface
type AdviceQuery interface {
	GetByTopic(topic string) (QueryResult, SlipError, error)
}

// RESTAdviceQuery struct hides the HTTP details behind a function
type RESTAdviceQuery struct {
	client http.Client
}

// GetByTopic function contains the HTTP details for the query to api.adviceslip.com
func (q *RESTAdviceQuery) GetByTopic(topic string) (QueryResult, SlipError, error) {
	fmt.Println("(HTTP call for querying " + topic + ")")

	query := url.PathEscape(topic)

	req, err := http.NewRequest("GET", "https://api.adviceslip.com/advice/search/"+query, nil)
	if err != nil {
		return QueryResult{}, SlipError{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := q.client.Do(req)
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

// AdviceMapping interface
type AdviceMapping interface {
	Map(result QueryResult) []string
}

// SimpleAdviceMapping struct
type SimpleAdviceMapping struct{}

// Map function
func (m *SimpleAdviceMapping) Map(result QueryResult) []string {
	slips := result.Slips
	amount := len(slips)

	ret := make([]string, amount)
	for i := 0; i < amount; i++ {
		ret[i] = slips[i].Advice
	}

	return ret
}
