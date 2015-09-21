package word2vec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type CosineQuery struct {
	A Expr `json:"a,omitempty"`
	B Expr `json:"b,omitempty"`
}

type CosineResponse struct {
	Value float32 `json:"value"`
}

func (q CosineQuery) Eval(m *Model) (*CosineResponse, error) {
	v, err := m.Cosine(q.A, q.B)
	if err != nil {
		return nil, err
	}

	return &CosineResponse{
		Value: v,
	}, nil
}

type CosinesQuery struct {
	Queries []CosineQuery `json:"queries"`
}

type CosinesResponse struct {
	Values []CosineResponse `json:"values"`
}

func (qs CosinesQuery) Eval(m *Model) (*CosinesResponse, error) {
	values := make([]CosineResponse, len(qs.Queries))
	for i, q := range qs.Queries {
		r, err := q.Eval(m)
		if err != nil {
			return nil, err
		}
		values[i] = *r
	}
	return &CosinesResponse{
		Values: values,
	}, nil
}

type CosineNQuery struct {
	Expr Expr `json:"expr"`
	N    int  `json:"n"`
}

type CosineNResponse struct {
	Matches []Match `json:"matches"`
}

func (q CosineNQuery) Eval(m *Model) (*CosineNResponse, error) {
	r, err := m.CosineN(q.Expr, q.N)
	if err != nil {
		return nil, err
	}

	return &CosineNResponse{
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
	mux.HandleFunc("/cosine-n", ms.handleCosineNQuery)
	mux.HandleFunc("/cosine", ms.handleCosineQuery)
	mux.HandleFunc("/cosines", ms.handleCosinesQuery)

	ms.ServeMux = mux
	return ms
}

func handleError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	log.Printf(msg)
	w.WriteHeader(status)
	w.Write([]byte(msg))
	return
}

func (m *Server) handleCosineQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var q CosineQuery
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

func (m *Server) handleCosinesQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var q CosinesQuery
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

func (m *Server) handleCosineNQuery(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var q CosineNQuery
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

type Client struct {
	Addr string
}

func (c Client) Cosine(x, y Expr) (float32, error) {
	req := CosineQuery{A: x, B: y}

	b, err := json.Marshal(req)
	if err != nil {
		return 0.0, err
	}

	r, err := http.NewRequest("GET", "http://"+c.Addr+"/cosine", bytes.NewReader(b))
	if err != nil {
		return 0.0, err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0.0, fmt.Errorf("non-%v status code: %v", http.StatusOK, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0.0, fmt.Errorf("error reading response: %v", err)
	}

	var data CosineResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0.0, fmt.Errorf("error unmarshalling result: %v", err)
	}

	return data.Value, nil
}

func (c Client) Cosines(pairs [][2]Expr) ([]float32, error) {
	req := CosinesQuery{
		Queries: make([]CosineQuery, len(pairs)),
	}
	for _, pair := range pairs {
		req.Queries = append(req.Queries, CosineQuery{
			A: pair[0],
			B: pair[1],
		})
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", "http://"+c.Addr+"/cosines", bytes.NewReader(b))
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

	var data CosinesResponse
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

func (c Client) CosineN(e Expr, n int) ([]Match, error) {
	req := CosineNQuery{Expr: e, N: n}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", "http://"+c.Addr+"/cosine-n", bytes.NewReader(b))
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

	var data CosineNResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling result: %v", err)
	}

	return data.Matches, nil
}
