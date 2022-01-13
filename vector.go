//go:build !amd64
// +build !amd64

package word2vec

import (
	"math"
)

// Normalise normalises the vector in-place.
func (v Vector) Normalise() {
	n := v.Norm()
	for i := range v {
		v[i] /= n
	}
}

// Norm computes the Euclidean norm of the vector.
func (v Vector) Norm() float32 {
	var out float32
	for _, vx := range v {
		out += vx * vx
	}
	return float32(math.Sqrt(float64(out)))
}

// Add performs v += a * u (in-place).
func (v Vector) Add(a float32, u Vector) {
	for i := range v {
		v[i] += a * u[i]
	}
}

// Dot computes the dot product with u.
func (v Vector) Dot(u Vector) float32 {
	var out float32
	for i, vx := range v {
		out += vx * u[i]
	}
	return out
}
