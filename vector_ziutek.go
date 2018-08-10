// +build !go1.8

package word2vec

import "github.com/ziutek/blas"

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
