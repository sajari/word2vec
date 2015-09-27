package word2vec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type CosQuery struct {
	A Expr `json:"a,omitempty"`
	B Expr `json:"b,omitempty"`
}

type CosResponse struct {
	Value float32 `json:"value"`
}

func (q CosQuery) Eval(m *Model) (*CosResponse, error) {
	v, err := m.Cos(q.A, q.B)
	if err != nil {
		return nil, err
	}

	return &CosResponse{
		Value: v,
	}, nil
}

type CosesQuery struct {
	Queries []CosQuery `json:"queries"`
}

type CosesResponse struct {
	Values []CosResponse `json:"values"`
}

func (qs CosesQuery) Eval(m *Model) (*CosesResponse, error) {
	values := make([]CosResponse, len(qs.Queries))
	for i, q := range qs.Queries {
		r, err := q.Eval(m)
		if err != nil {
			return nil, err
		}
		values[i] = *r
	}
	return &CosesResponse{
		Values: values,
	}, nil
}

type CosNQuery struct {
	Expr Expr `json:"expr"`
	N    int  `json:"n"`
}

type CosNResponse struct {
	Matches []Match `json:"matches"`
}

func (q CosNQuery) Eval(m *Model) (*CosNResponse, error) {
	r, err := m.CosN(q.Expr, q.N)
	if err != nil {
		return nil, err
	}

	return &CosNResponse{
		Matches: r,
	}, nil
}

type Server struct {
	*Model
	*http.ServeMux
}

func NewServer(m *Model) http.Handler {
	ms := &Server{
		Model: m,
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

func (m *Server) handleCosQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var q CosQuery
	err := dec.Decode(&q)
	if err != nil {
		msg := fmt.Sprintf("error decoding query: %v", err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}

	resp, err := q.Eval(m.Model)
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

func (m *Server) handleCosesQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var q CosesQuery
	err := dec.Decode(&q)
	if err != nil {
		msg := fmt.Sprintf("error decoding query: %v", err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}

	resp, err := q.Eval(m.Model)
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

func (m *Server) handleCosNQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var q CosNQuery
	err := dec.Decode(&q)
	if err != nil {
		msg := fmt.Sprintf("error decoding query: %v", err)
		handleError(w, r, http.StatusInternalServerError, msg)
		return
	}

	resp, err := q.Eval(m.Model)
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

// Client is type which implements Coser and evaluates Expr similarity queries
// using a word2vec Server (see above).
type Client struct {
	Addr string
}

// Cos implements Coser.
func (c Client) Cos(x, y Expr) (float32, error) {
	req := CosQuery{A: x, B: y}

	b, err := json.Marshal(req)
	if err != nil {
		return 0.0, err
	}

	r, err := http.NewRequest("GET", "http://"+c.Addr+"/cos", bytes.NewReader(b))
	if err != nil {
		return 0.0, err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0.0, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return 0.0, fmt.Errorf("error: %v", string(b))
	}

	if resp.StatusCode != http.StatusOK {
		return 0.0, fmt.Errorf("non-%v status code: %v msg: %v", http.StatusOK, resp.Status, string(b))
	}

	var data CosResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0.0, fmt.Errorf("error unmarshalling result: %v", err)
	}

	return data.Value, nil
}

// Coses implements Coser.
func (c Client) Coses(pairs [][2]Expr) ([]float32, error) {
	req := CosesQuery{
		Queries: make([]CosQuery, len(pairs)),
	}
	for _, pair := range pairs {
		req.Queries = append(req.Queries, CosQuery{
			A: pair[0],
			B: pair[1],
		})
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", "http://"+c.Addr+"/coses", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-%v status code: %v", http.StatusOK, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var data CosesResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling result: %v", err)
	}

	result := make([]float32, len(data.Values))
	for i, v := range data.Values {
		result[i] = v.Value
	}
	return result, nil
}

// CosN implements Coser.
func (c Client) CosN(e Expr, n int) ([]Match, error) {
	req := CosNQuery{Expr: e, N: n}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", "http://"+c.Addr+"/cos-n", bytes.NewReader(b))
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
		return nil, fmt.Errorf("error: %v", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-%v status code: %v msg: %v", http.StatusOK, resp.Status, string(body))
	}

	var data CosNResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling result: %v", err)
	}

	return data.Matches, nil
}
