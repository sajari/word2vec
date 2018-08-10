package word2vec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestFromReader(t *testing.T) {
	vecs := map[string]Vector{
		"hello": Vector{0, 1},
		"world": Vector{1, 0},
	}
	dim := 2

	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, len(vecs), dim)

	for k, v := range vecs {
		fmt.Fprintf(buf, "%s ", k)
		err := binary.Write(buf, binary.LittleEndian, v)
		if err != nil {
			t.Errorf("unexpected error writing vector")
		}
		fmt.Fprintf(buf, "\n")
	}

	m, err := FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Errorf("unexpected error from FromReader: %v", err)
	}

	if m.Size() != len(vecs) {
		t.Errorf("m.Size() = %d, expected %d", m.Size(), len(vecs))
	}

	if m.Dim() != 2 {
		t.Errorf("m.Dim() = %d, expected 2", m.Dim())
	}

	mVecs := m.Map([]string{"hello", "world"})
	if !reflect.DeepEqual(vecs, mVecs) {
		t.Errorf("m.Map() = %v, expected %v", mVecs, vecs)
	}

	x := Expr{"hello": 1.0}
	expectedMatches := []Match{
		{Word: "hello", Score: 1.0},
		{Word: "world", Score: 0.0},
	}
	matches, err := m.CosN(x, 2)
	if err != nil {
		t.Errorf("unexpected error from m.CosN(x, 2): %v", err)
	}
	if len(matches) != 2 {
		t.Errorf("len(matches) = %d, expected %d", len(matches), 2)
	}

	if !reflect.DeepEqual(matches, expectedMatches) {
		t.Errorf("m.CosN(x, 2) = %v, expected: %v", matches, expectedMatches)
	}

	y := Expr{"world": 1.0}
	expectedCos := float32(0.0)
	c, err := m.Cos(x, y)
	if err != nil {
		t.Errorf("unexpected error from m.Cos(x, y): %v", err)
	}
	if c != expectedCos {
		t.Errorf("Cos(x, y) = %f, expected %f", c, expectedCos)
	}
}

func TestAddWeight(t *testing.T) {
	x := Expr{}
	y := Expr{}

	weights := []float32{0.1, 0.2}
	words := []string{"one", "two"}

	AddWeight(x, weights, words)

	Add(y, 0.1, []string{"one", "two"})
	Add(y, 0.1, []string{"two"})

	if !reflect.DeepEqual(x, y) {
		t.Errorf("x = %v, y = %v", x, y)
	}
}

func data(n int) []float32 {
	rand.Seed(10)
	data := []float32{}
	for index := 0; index < n; index++ {
		data = append(data, rand.Float32())
	}
	return data
}
