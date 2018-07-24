// +build !go1.8

package word2vec

import (
	"fmt"
	"math/rand"
	"testing"
)

func data(n int) []float32 {
	rand.Seed(10)
	data := []float32{}
	for index := 0; index < n; index++ {
		data = append(data, rand.Float32())
	}
	return data
}

func BenchmarkDotFloat32(b *testing.B) {
	benchmarks := []struct {
		name string
		dim  int
	}{}
	for _, dim := range []int{10, 50, 100, 150, 200, 250, 300, 350, 400, 450, 500} {
		benchmarks = append(benchmarks, struct {
			name string
			dim  int
		}{
			name: fmt.Sprintf("test with dimension %d", dim),
			dim:  dim,
		})
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			vData := data(bm.dim)
			uData := data(bm.dim)
			for i := 0; i < b.N; i++ {
				v := Vector(vData)
				u := Vector(uData)
				v.Dot(u)
			}
		})
	}
}

func BenchmarkAddFloat32(b *testing.B) {
	benchmarks := []struct {
		name string
		dim  int
		a    float32
	}{}
	for i, dim := range []int{10, 50, 100, 150, 200, 250, 300, 350, 400, 450, 500} {
		benchmarks = append(benchmarks, struct {
			name string
			dim  int
			a    float32
		}{
			name: fmt.Sprintf("test with dimension %d", dim),
			dim:  dim,
			a:    float32(i * 10),
		})
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			vData := data(bm.dim)
			uData := data(bm.dim)
			for i := 0; i < b.N; i++ {
				v := Vector(vData)
				u := Vector(uData)
				v.Add(bm.a, u)
			}
		})
	}
}

func BenchmarkNormFloat32(b *testing.B) {
	benchmarks := []struct {
		name string
		dim  int
	}{}
	for _, dim := range []int{10, 50, 100, 150, 200, 250, 300, 350, 400, 450, 500} {
		benchmarks = append(benchmarks, struct {
			name string
			dim  int
		}{
			name: fmt.Sprintf("test with dimension %d", dim),
			dim:  dim,
		})
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			d := data(bm.dim)
			for i := 0; i < b.N; i++ {
				v := Vector(d)
				v.Norm()
			}
		})
	}
}
