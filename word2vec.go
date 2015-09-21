// Package word2vec provides functionality for reading binary word2vec models
// and basic usage (see https://code.google.com/p/word2vec/).
package word2vec

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/ziutek/blas"
)

// Model is a type which represents a word2vec Model.
type Model struct {
	dim   int
	words map[string]Vector
}

// FromReader creates a Model using the binary model data provided by the io.Reader.
func FromReader(r io.Reader) (*Model, error) {
	br := bufio.NewReader(r)
	var size, dim int
	n, err := fmt.Fscanln(r, &size, &dim)
	if err != nil {
		return nil, err
	}
	if n != 2 {
		return nil, fmt.Errorf("could not extract size/dim from binary Data")
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
		err = binary.Read(br, binary.LittleEndian, v)
		if err != nil {
			return nil, err
		}

		v.Normalise()

		_, err = br.ReadByte()
		if err != nil {
			return nil, err
		}

		m.words[w] = v
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
	return fmt.Sprintf("word not found: %v", e.Word)
}

// Expr is a type which represents a linear expresssion which can be evaluated to a vector
// by a word2vec Model.
type Expr map[string]float32

// Add appends the given word with coefficient to the expression.  If the word already exists
// in the expression, then the coefficients are added.
func (e Expr) Add(f float32, w string) {
	e[w] += f
}

// AddAll is a convenience method which adds all the words in the slice to the Expr, using the given
// coefficient.
func AddAll(e Expr, f float32, ws []string) {
	for _, w := range ws {
		e.Add(f, w)
	}
}

// Eval evaluates the Expr to a Vector using a Model.
func (e Expr) Eval(m *Model) (Vector, error) {
	if len(e) == 0 {
		return nil, fmt.Errorf("must specify at least one word to evaluate")
	}
	return m.Eval(e)
}

// Size returns the number of words in the model.
func (m *Model) Size() int {
	return len(m.words)
}

// Dim returns the dimention of the vectors in the model.
func (m *Model) Dim() int {
	return m.dim
}

// Similarity returns the similarity between the two words.
func (m *Model) Similarity(x, y string) (float32, error) {
	u, ok := m.words[x]
	if !ok {
		return 0.0, &NotFoundError{x}
	}
	v, ok := m.words[y]
	if !ok {
		return 0.0, &NotFoundError{y}
	}
	return u.Dot(v), nil
}

// Vectors returns a mapping word -> Vector for each word in `w`,
// unknown words are ignored.
func (m *Model) Vectors(words []string) map[string]Vector {
	result := make(map[string]Vector)
	for _, w := range words {
		if v, ok := m.words[w]; ok {
			result[w] = v
		}
	}
	return result
}

func (m *Model) Sim(u, v Vector) float32 {
	return u.Dot(v)
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
	return v, nil
}

// Eval is a convenience method which
func Evaluate(m *Model, add []string, sub []string) (Vector, error) {
	e := Expr{}
	AddAll(e, 1, add)
	AddAll(e, -1, sub)
	return e.Eval(m)
}

// Match is a type which represents a pairing of a word and score indicating
// the similarity of this word against a search word.
type Match struct {
	Word  string  `json:"word"`
	Score float32 `json:"score"`
}

// SimN is a method which returns a list of `n` most similar vectors
// to `v` in the model.
func (m *Model) SimN(v Vector, n int) []Match {
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

type multiMatches struct {
	Word    string
	Matches []Match
}

// MultiSimN takes a map of word -> vector (see Vectors) and computes the
// n most similar words for each.
func MultiSimN(m *Model, vecs map[string]Vector, n int) map[string][]Match {
	wg := &sync.WaitGroup{}
	wg.Add(len(vecs))
	ch := make(chan multiMatches, len(vecs))
	for k, v := range vecs {
		go func(k string, v Vector) {
			ch <- multiMatches{Word: k, Matches: m.SimN(v, n)}
			wg.Done()
		}(k, v)
	}
	wg.Wait()
	close(ch)

	result := make(map[string][]Match, len(vecs))
	for r := range ch {
		result[r.Word] = r.Matches
	}
	return result
}
