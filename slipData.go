package main

// Slip struct from Advice Slip REST Service
type Slip struct {
	ID     int    `json:"id"`
	Advice string `json:"advice"`
	Date   string `json:"date"`
}

// Message struct from Advice Slip REST Service
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SlipError struct from Advice Slip REST Service
type SlipError struct {
	Message Message `json:"message"`
}

// QueryResult struct from Advice Slip REST Service
type QueryResult struct {
	ResultsAmount string `json:"total_results"`
	Query         string `json:"query"`
	Slips         []Slip `json:"slips"`
}
