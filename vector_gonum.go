// +build go1.8

package word2vec

import "gonum.org/v1/gonum/blas/blas32"

// Vector is a type which represents a word vector.
type Vector []float32

// Normalise normalises the vector in-place.
func (v Vector) Normalise() {
	w := blas32.Nrm2(len(v), blas32.Vector{Inc: 1, Data: v})
	blas32.Scal(len(v), 1/w, blas32.Vector{Inc: 1, Data: v})
}

// Norm computes the Euclidean norm of the vector.
func (v Vector) Norm() float32 {
	return blas32.Nrm2(len(v), blas32.Vector{Inc: 1, Data: v})
}

// Add performs v += a * u (in-place).
func (v Vector) Add(a float32, u Vector) {
	blas32.Axpy(len(v), a, blas32.Vector{Inc: 1, Data: u}, blas32.Vector{Inc: 1, Data: v})
}

// Dot computes the dot product with u.
func (v Vector) Dot(u Vector) float32 {
	x := blas32.Vector{
		Inc:  1,
		Data: u,
	}
	y := blas32.Vector{
		Inc:  1,
		Data: v,
	}
	return blas32.Dot(len(v), x, y)
}
