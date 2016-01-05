package word2vec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type cosQuery struct {
	A Expr `json:"a,omitempty"`
	B Expr `json:"b,omitempty"`
}

type cosResponse struct {
	Value float32 `json:"value"`
}

func (q cosQuery) Eval(c Coser) (interface{}, error) {
	v, err := c.Cos(q.A, q.B)
	if err != nil {
		return nil, err
	}

	return &cosResponse{
		Value: v,
	}, nil
}

type cosesQuery struct {
	A []Expr `json:"a"`
	B []Expr `json:"b"`
}

type cosesResponse struct {
	Values []float32 `json:"values"`
}

func (qs cosesQuery) Eval(c Coser) (interface{}, error) {
	query := make([][2]Expr, len(qs.A))
	for i := range qs.A {
		query[i] = [2]Expr{qs.A[i], qs.B[i]}
	}
	result, err := c.Coses(query)
	if err != nil {
		return nil, err
	}
	return struct {
		Values []float32
	}{
		Values: result,
	}, nil
}

type cosNQuery struct {
	Expr Expr `json:"expr"`
	N    int  `json:"n"`
}

type cosNResponse struct {
	Matches []Match `json:"matches"`
}

func (q cosNQuery) Eval(c Coser) (interface{}, error) {
	r, err := c.CosN(q.Expr, q.N)
	if err != nil {
		return nil, err
	}

	return &cosNResponse{
		Matches: r,
	}, nil
}

// server is a type which implements http.Handler and exports endpoints
// for performing similarity queries on a word2vec model.
type server struct {
	Coser
	*http.ServeMux
}

// NewServer creates a new word2vec server which exports endpoints for performing
// similarity queries on a word2vec Model.
func NewServer(c Coser) http.Handler {
	ms := &server{
		Coser: c,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/cos-n", ms.handleCosNQuery)
	mux.HandleFunc("/cos", ms.handleCosQuery)
	mux.HandleFunc("/coses", ms.handleCosesQuery)

	ms.ServeMux = mux
	return ms
}

func handleError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	log.Printf(msg)
	w.WriteHeader(status)
	w.Write([]byte(msg))
	return
}

type evaler interface {
	Eval(Coser) (interface{}, error)
}

func (s *server) handleEval(e evaler, w http.ResponseWriter, r *http.Request) {
	resp, err := e.Eval(s.Coser)
	if err != nil {
		msg := fmt.Sprintf("error evaluating query: %v", err)
		handleError(w, r, http.StatusBadRequest, msg)
		return
	}

	b, err := json.Marshal(resp)
	if err != nil {
		msg := fmt.Sprintf("error encoding response %#v to JSON: %v", resp, err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Printf("error writing response: %v", err)
	}
}

func (s *server) handleCosQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var q cosQuery
	err := dec.Decode(&q)
	if err != nil {
		msg := fmt.Sprintf("error decoding query: %v", err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}
	s.handleEval(q, w, r)
}

func (s *server) handleCosesQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var q cosesQuery
	err := dec.Decode(&q)
	if err != nil {
		msg := fmt.Sprintf("error decoding query: %v", err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}
	s.handleEval(q, w, r)
}

func (s *server) handleCosNQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var q cosNQuery
	err := dec.Decode(&q)
	if err != nil {
		msg := fmt.Sprintf("error decoding query: %v", err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}
	s.handleEval(q, w, r)
}

// Client is type which implements Coser and evaluates Expr similarity queries
// using a word2vec Server (see above).
type Client struct {
	Addr string
}

func (c Client) fetch(x interface{}, suffix string) ([]byte, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s", c.Addr, suffix), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		body := string(body)
		if strings.HasPrefix(body, "error evaluating query: word not found:") {
			var w string
			if _, err := fmt.Sscanf(body, "error evaluating query: word not found: %q", &w); err == nil {
				return nil, NotFoundError{Word: w}
			}
		}
		return nil, errors.New(body)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-%v status code: %v msg: %v", http.StatusOK, resp.Status, string(body))
	}

	return body, nil
}

// Cos implements Coser.
func (c Client) Cos(x, y Expr) (float32, error) {
	req := cosQuery{A: x, B: y}
	body, err := c.fetch(req, "cos")
	if err != nil {
		return 0.0, err
	}

	var data cosResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0.0, fmt.Errorf("error unmarshalling result: %v", err)
	}

	return data.Value, nil
}

// Coses implements Coser.
func (c Client) Coses(pairs [][2]Expr) ([]float32, error) {
	req := cosesQuery{
		A: make([]Expr, len(pairs)),
		B: make([]Expr, len(pairs)),
	}
	for i, pair := range pairs {
		req.A[i] = pair[0]
		req.B[i] = pair[1]
	}

	body, err := c.fetch(req, "coses")
	if err != nil {
		return nil, err
	}

	var data cosesResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling result: %v", err)
	}
	return data.Values, nil
}

// CosN implements Coser.
func (c Client) CosN(e Expr, n int) ([]Match, error) {
	req := cosNQuery{Expr: e, N: n}
	body, err := c.fetch(req, "cos-n")
	if err != nil {
		return nil, err
	}

	var data cosNResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling result: %v", err)
	}
	return data.Matches, nil
}
