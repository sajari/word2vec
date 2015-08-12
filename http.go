package word2vec

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Query struct {
	Add []string `json:"add"`
	Sub []string `json:"sub"`
	N   int      `json:"n"`
}

type Queries struct {
	Queries []Query `json:"queries"`
}

type Response struct {
	Query  Query   `json:"query"`
	Result []Match `json:"result"`
}

type Client interface {
	Get(q Query) Response
}

type ModelServer struct {
	*Model
}

func handleError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	log.Printf(msg)
	w.WriteHeader(status)
	w.Write([]byte(msg))
	return
}

func (m *ModelServer) HandleQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var query Query
	err := dec.Decode(&query)
	if err != nil {
		msg := fmt.Sprintf("error decoding query: %v", err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}

	if len(query.Add) == 0 && len(query.Sub) == 0 {
		msg := fmt.Sprintf("must specify either 'add' or 'sub'")
		handleError(w, r, http.StatusBadRequest, msg)
	}

	v, err := m.Eval(query.Add, query.Sub)
	if err != nil {
		msg := fmt.Sprintf("error creating target vector: %v", err)
		handleError(w, r, http.StatusBadRequest, msg)
		return
	}

	resp := Response{
		Query:  query,
		Result: m.MostSimilar(v, query.N),
	}

	b, err := json.Marshal(resp)
	if err != nil {
		msg := fmt.Sprintf("error encoding response %#v to JSON: %v", resp, err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Println("error writing response: %v", err)
	}
}

func (m *ModelServer) HandleQueries(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var query Query
	err := dec.Decode(&query)
	if err != nil {
		msg := fmt.Sprintf("error decoding query: %v", err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}

	if len(query.Add) == 0 && len(query.Sub) == 0 {
		msg := fmt.Sprintf("must specify either 'add' or 'sub'")
		handleError(w, r, http.StatusBadRequest, msg)
	}

	v, err := m.Eval(query.Add, query.Sub)
	if err != nil {
		msg := fmt.Sprintf("error creating target vector: %v", err)
		handleError(w, r, http.StatusBadRequest, msg)
		return
	}

	resp := Response{
		Query:  query,
		Result: m.MostSimilar(v, query.N),
	}

	b, err := json.Marshal(resp)
	if err != nil {
		msg := fmt.Sprintf("error encoding response %#v to JSON: %v", resp, err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Println("error writing response: %v", err)
	}
}
