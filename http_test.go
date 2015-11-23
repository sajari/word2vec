package word2vec

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type testCoser struct {
	cos   func(x, y Expr) (float32, error)
	coses func(pairs [][2]Expr) ([]float32, error)
	cosN  func(x Expr, n int) ([]Match, error)
}

func (t testCoser) Cos(x, y Expr) (float32, error)           { return t.cos(x, y) }
func (t testCoser) Coses(pairs [][2]Expr) ([]float32, error) { return t.coses(pairs) }
func (t testCoser) CosN(x Expr, n int) ([]Match, error)      { return t.cosN(x, n) }

func TestEndToEndCos(t *testing.T) {
	tc := &testCoser{}
	h := NewServer(tc)
	s := httptest.NewServer(h)
	defer s.Close()

	c := Client{
		Addr: strings.TrimPrefix(s.URL, "http://"),
	}

	cosTests := []struct {
		x, y Expr
		f    float32
		err  error
	}{
		{
			x:   Expr{"hello": 1.0},
			y:   Expr{"world": 1.0},
			f:   1.0,
			err: nil,
		},
		{
			x:   Expr{"": 1.0},
			y:   Expr{"empty": 1.0},
			f:   0,
			err: fmt.Errorf("suffix test"),
		},
		{
			x:   Expr{"hello": 1.0, "something": 2.0},
			y:   Expr{"world": 1.0, "else": 0.0},
			f:   1.0,
			err: nil,
		},
	}

	for _, tt := range cosTests {
		var cosX, cosY Expr

		tc.cos = func(x, y Expr) (float32, error) {
			cosX = x
			cosY = y
			return tt.f, tt.err
		}

		f, err := c.Cos(tt.x, tt.y)
		if !reflect.DeepEqual(cosX, tt.x) {
			t.Errorf("cosX = %#v, expected: %#v", cosX, tt.x)
		}
		if !reflect.DeepEqual(cosY, tt.y) {
			t.Errorf("cosY = %#v, expected: %#v", cosY, tt.y)
		}
		if f != tt.f {
			t.Errorf("f = %#v, expected: %#v", f, tt.f)
		}
		if tt.err != nil {
			if err == nil {
				t.Errorf("err = %#v, expected: %#v", err, tt.err)
			}
			if !strings.HasSuffix(err.Error(), tt.err.Error()) {
				t.Errorf("err = %q, expected suffix: %q", err, tt.err)
			}
		}
	}
}

func TestEndToEndCoses(t *testing.T) {
	tc := &testCoser{}
	h := NewServer(tc)
	s := httptest.NewServer(h)
	defer s.Close()

	c := Client{
		Addr: strings.TrimPrefix(s.URL, "http://"),
	}

	cosesTests := []struct {
		pairs [][2]Expr
		f     []float32
		err   error
	}{
		{
			pairs: [][2]Expr{
				{{"hello": 1.0}, {"world": 1.0}},
			},
			f:   []float32{1.0},
			err: nil,
		},
		{
			pairs: [][2]Expr{
				{{"hello": 1.0}, {"world": 1.0}},
			},
			f:   nil,
			err: errors.New("suffix"),
		},
	}

	for _, tt := range cosesTests {
		var cosesPairs [][2]Expr
		tc.coses = func(pairs [][2]Expr) ([]float32, error) {
			cosesPairs = pairs
			return tt.f, tt.err
		}

		f, err := c.Coses(tt.pairs)
		if !reflect.DeepEqual(cosesPairs, tt.pairs) {
			t.Errorf("cosesPairs = %#v, expected: %#v", cosesPairs, tt.pairs)
		}
		if !reflect.DeepEqual(f, tt.f) {
			t.Errorf("f = %#v, expected: %#v", f, tt.f)
		}
		if tt.err != nil {
			if err == nil {
				t.Errorf("err = %#v, expected: %#v", err, tt.err)
			}
			if !strings.HasSuffix(err.Error(), tt.err.Error()) {
				t.Errorf("err = %q, expected suffix: %q", err, tt.err)
			}
		}
	}
}

func TestEndToEndCosN(t *testing.T) {
	tc := &testCoser{}
	h := NewServer(tc)
	s := httptest.NewServer(h)
	defer s.Close()

	c := Client{
		Addr: strings.TrimPrefix(s.URL, "http://"),
	}

	cosNTests := []struct {
		x   Expr
		n   int
		m   []Match
		err error
	}{
		{
			x:   Expr{"hello": 1.0},
			n:   0,
			m:   []Match{},
			err: nil,
		},
		{
			x: Expr{"hello": 1.0},
			n: 10,
			m: []Match{
				{"hello", 1.0},
				{"something", 2.0},
				{"else", 1.0},
				{"is", 0.5},
				{"in", 0.1},
				{"the", 0.4},
			},
			err: nil,
		},
		{
			x:   Expr{"hello": 1.0},
			n:   10,
			m:   nil,
			err: fmt.Errorf("suffix error"),
		},
	}

	for _, tt := range cosNTests {
		var cosNN int
		var cosNX Expr
		tc.cosN = func(x Expr, n int) ([]Match, error) {
			cosNX = x
			cosNN = n
			return tt.m, tt.err
		}

		m, err := c.CosN(tt.x, tt.n)
		if !reflect.DeepEqual(cosNX, tt.x) {
			t.Errorf("cosNX = %#v, expected: %#v", cosNX, tt.x)
		}
		if !reflect.DeepEqual(cosNN, tt.n) {
			t.Errorf("cosNN = %#v, expected: %#v", cosNN, tt.n)
		}
		if !reflect.DeepEqual(m, tt.m) {
			t.Errorf("m = %#v, expected: %#v", m, tt.m)
		}
		if tt.err != nil {
			if err == nil {
				t.Errorf("err = %#v, expected: %#v", err, tt.err)
			}
			if !strings.HasSuffix(err.Error(), tt.err.Error()) {
				t.Errorf("err = %q, expected suffix: %q", err, tt.err)
			}
		}
	}
}
