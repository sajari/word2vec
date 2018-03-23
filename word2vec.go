// Package word2vec provides functionality for reading binary word2vec models
// and performing cosine similarity queries (see https://code.google.com/p/word2vec/).
package word2vec // import "code.sajari.com/word2vec"

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/ziutek/blas"
)

// Model is a type which represents a word2vec Model and implements the Coser
// and Mapper interfaces.
type Model struct {
	dim   int
	words map[string]Vector
}

var (
	_ Coser  = (*Model)(nil)
	_ Mapper = (*Model)(nil)
)

// FromReader creates a Model using the binary model data provided by the io.Reader.
func FromReader(r io.Reader) (*Model, error) {
	br := bufio.NewReader(r)
	var size, dim int
	n, err := fmt.Fscanln(r, &size, &dim)
	if err != nil {
		return nil, err
	}
	if n != 2 {
		return nil, fmt.Errorf("could not extract size/dim from binary model data")
	}

	m := &Model{
		words: make(map[string]Vector, size),
		dim:   dim,
	}

	raw := make([]float32, size*dim)

	for i := 0; i < size; i++ {
		w, err := br.ReadString(' ')
		if err != nil {
			return nil, err
		}
		w = w[:len(w)-1]

		v := Vector(raw[dim*i : m.dim*(i+1)])
		if err := binary.Read(br, binary.LittleEndian, v); err != nil {
			return nil, err
		}

		v.Normalise()

		m.words[w] = v

		b, err := br.ReadByte()
		if err != nil {
			if i == size-1 && err == io.EOF {
				break
			}
			return nil, err
		}
		if b != byte('\n') {
			if err := br.UnreadByte(); err != nil {
				return nil, err
			}
		}
	}
	return m, nil
}

// Vector is a type which represents a word vector.
type Vector []float32

// Normalise normalises the vector in-place.
func (v Vector) Normalise() {
	w := blas.Snrm2(len(v), v, 1)
	blas.Sscal(len(v), 1/w, v, 1)
}

// Norm computes the Euclidean norm of the vector.
func (v Vector) Norm() float32 {
	return blas.Snrm2(len(v), v, 1)
}

// Add performs v += a * u (in-place).
func (v Vector) Add(a float32, u Vector) {
	blas.Saxpy(len(v), a, u, 1, v, 1)
}

// Dot computes the dot product with u.
func (v Vector) Dot(u Vector) float32 {
	return blas.Sdot(len(v), u, 1, v, 1)
}

// NotFoundError is an error returned from Model functions when an input
// word is not in the model.
type NotFoundError struct {
	Word string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("word not found: %q", e.Word)
}

// Expr is a type which represents a linear expresssion of (weight, word) pairs
// which can be evaluated to a vector by a word2vec Model.
type Expr map[string]float32

// Add appends the given word with specified weight to the expression.  If the word already
// exists in the expression, then the weights are added.
func (e Expr) Add(weight float32, word string) {
	e[word] += weight
}

// Eval evaluates the Expr to a Vector using a Model.
func (e Expr) Eval(m *Model) (Vector, error) {
	if len(e) == 0 {
		return nil, fmt.Errorf("must specify at least one word to evaluate")
	}
	return m.Eval(e)
}

// Add is a convenience method for adding multiple words to an Expr.
func Add(e Expr, weight float32, words []string) {
	for _, w := range words {
		e.Add(weight, w)
	}
}

// AddWeight is a convenience method for adding multiple weighted words to an Expr.
func AddWeight(e Expr, weights []float32, words []string) {
	if len(weights) != len(words) {
		panic("weight and words must be the same length")
	}

	for i, w := range weights {
		e.Add(w, words[i])
	}
}

// Coser is an interface which defines methods which can evaluate cosine similarity
// between Exprs.
type Coser interface {
	// Cos computes the cosine similarity of the expressions.
	Cos(e, f Expr) (float32, error)

	// Coses computes the cosine similarity of pairs of expressions.
	Coses(pairs [][2]Expr) ([]float32, error)

	// CosN computes the N most similar words to the expression.
	CosN(e Expr, n int) ([]Match, error)
}

// Size returns the number of words in the model.
func (m *Model) Size() int {
	return len(m.words)
}

// Dim returns the dimention of the vectors in the model.
func (m *Model) Dim() int {
	return m.dim
}

// Mapper is an interface which defines a method which can return a mapping of
// word -> Vector for each word in words.
type Mapper interface {
	Map(words []string) map[string]Vector
}

// Map returns a mapping word -> Vector for each word in `words`.
// Unknown words are ignored.
func (m *Model) Map(words []string) map[string]Vector {
	result := make(map[string]Vector)
	for _, w := range words {
		if v, ok := m.words[w]; ok {
			result[w] = v
		}
	}
	return result
}

// Cos returns the cosine similarity of the given expressions.
func (m *Model) Cos(a, b Expr) (float32, error) {
	u, err := a.Eval(m)
	if err != nil {
		return 0, err
	}

	v, err := b.Eval(m)
	if err != nil {
		return 0, err
	}
	return u.Dot(v), nil
}

// Coses returns the cosine similarity of each pair of expressions in the list.  Returns
// immediately if an error occurs.
func (m *Model) Coses(pairs [][2]Expr) ([]float32, error) {
	out := make([]float32, len(pairs))
	for i, p := range pairs {
		c, err := m.Cos(p[0], p[1])
		if err != nil {
			return nil, err
		}
		out[i] = c
	}
	return out, nil
}

// Eval constructs a vector by evaluating the expression
// vector.  Returns an error if a word is not in the model.
func (m *Model) Eval(expr Expr) (Vector, error) {
	v := Vector(make([]float32, m.dim))
	for w, c := range expr {
		u, ok := m.words[w]
		if !ok {
			return nil, &NotFoundError{w}
		}
		v.Add(c, u)
	}
	v.Normalise()
	return v, nil
}

// Match is a type which represents a pairing of a word and score indicating
// the similarity of this word against a search word.
type Match struct {
	Word  string  `json:"word"`
	Score float32 `json:"score"`
}

// CosN computes the n most similar words to the expression.  Returns an error if the
// expression could not be evaluated.
func (m *Model) CosN(e Expr, n int) ([]Match, error) {
	if n == 0 {
		return nil, nil
	}

	v, err := e.Eval(m)
	if err != nil {
		return nil, err
	}

	v.Normalise()
	return m.cosineN(v, n), nil
}

// cosineN is a method which returns a list of `n` most similar vectors to `v` in the model.
func (m *Model) cosineN(v Vector, n int) []Match {
	r := make([]Match, n)
	for w, u := range m.words {
		score := v.Dot(u)
		p := Match{w, score}
		// TODO(dhowden): MaxHeap would be better here if n is large.
		if r[n-1].Score > p.Score {
			continue
		}
		r[n-1] = p
		for j := n - 2; j >= 0; j-- {
			if r[j].Score > p.Score {
				break
			}
			r[j], r[j+1] = p, r[j]
		}
	}
	return r
}

type matchHeap []Match

func (h matchHeap) Len() int           { return len(h) }
func (h matchHeap) Less(i, j int) bool { return h[i].Score < h[j].Score }
func (h matchHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *matchHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(Match))
}

func (h *matchHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type multiMatches struct {
	N       int
	Matches []Match
}

// MultiCosN takes a list of expressions and computes the
// n most similar words for each.
func MultiCosN(m *Model, exprs []Expr, n int) ([][]Match, error) {
	if n == 0 {
		return make([][]Match, len(exprs)), nil
	}

	vecs := make([]Vector, len(exprs))
	for i, e := range exprs {
		v, err := e.Eval(m)
		if err != nil {
			return nil, err
		}
		vecs[i] = v
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(vecs))
	ch := make(chan multiMatches, len(vecs))
	for i, v := range vecs {
		go func(i int, v Vector) {
			ch <- multiMatches{N: i, Matches: m.cosineN(v, n)}
			wg.Done()
		}(i, v)
	}
	wg.Wait()
	close(ch)

	result := make([][]Match, len(vecs))
	for r := range ch {
		result[r.N] = r.Matches
	}
	return result, nil
}
