package word2vec

import (
	"crypto/sha1"
	"sort"
	"strconv"
)

// NewCache returns a Coser which will cache repeated calls to the Cos method,
// particularly useful when using Client.
func NewCache(c Coser) Coser {
	return &cache{
		Coser:     c,
		cache:     make(map[string]float32),
		errCache:  make(map[string]error),
		cosnCache: make(map[string][]Match),
	}
}

type cache struct {
	Coser

	errCache  map[string]error
	cache     map[string]float32
	cosnCache map[string][]Match
}

func hashExpr(x Expr) string {
	list := make([]string, 0, len(x))
	for k := range x {
		list = append(list, k)
	}
	sort.Strings(list)

	h := sha1.New()
	for _, w := range list {
		h.Write([]byte(w))
		h.Write([]byte(strconv.FormatFloat(float64(x[w]), 'f', -1, 64)))
	}
	return string(h.Sum(nil))
}

// Cos implements Coser.
func (c *cache) Cos(x, y Expr) (float32, error) {
	xh := hashExpr(x)
	if err, ok := c.errCache[xh]; ok {
		return 0, err
	}
	yh := hashExpr(y)
	if err, ok := c.errCache[yh]; ok {
		return 0, err
	}

	if f, ok := c.cache[xh+yh]; ok {
		return f, nil
	}

	f, err := c.Coser.Cos(x, y)
	if err != nil {
		if errNotFound, ok := err.(NotFoundError); ok {
			w := errNotFound.Word
			if _, ok := x[w]; ok {
				c.errCache[xh] = err
			}
			if _, ok := y[w]; ok {
				c.errCache[yh] = err
			}
		}
		return 0, err
	}
	c.cache[xh+yh] = f
	return f, nil
}

// CosN implements Coser.
func (c *cache) CosN(e Expr, n int) ([]Match, error) {
	eh := hashExpr(e)
	if err, ok := c.errCache[eh]; ok {
		return nil, err
	}

	if result, ok := c.cosnCache[eh]; ok {
		return result, nil
	}

	result, err := c.Coser.CosN(e, n)
	if err != nil {
		if errNotFound, ok := err.(NotFoundError); ok {
			if _, ok := e[errNotFound.Word]; ok {
				c.errCache[eh] = err
			}
		}
		return nil, err
	}
	c.cosnCache[eh] = result
	return result, nil
}
