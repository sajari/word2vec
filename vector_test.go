package word2vec

import (
	"fmt"
	"reflect"
	"testing"
)

func TestVectorAdd(t *testing.T) {
	tests := []struct {
		x, y, ans []float32
	}{
		{
			x:   []float32{0},
			y:   []float32{0},
			ans: []float32{0},
		},
	}

	for _, tt := range tests {
		v := Vector(tt.x)
		u := Vector(tt.y)
		v.Add(1.0, u)

		vans := Vector(tt.ans)
		if !reflect.DeepEqual(v, vans) {
			t.Errorf("x.Add(y) = %v, expected %v", v, vans)
		}
	}

}

func TestVectorDot(t *testing.T) {
	tests := []struct {
		x, y []float32
		ans  float32
	}{
		{
			x:   []float32{0},
			y:   []float32{0},
			ans: 0,
		},
		{
			x:   []float32{1},
			y:   []float32{0},
			ans: 0,
		},
		{
			x:   []float32{0},
			y:   []float32{1},
			ans: 0,
		},
		{
			x:   []float32{1},
			y:   []float32{1},
			ans: 1,
		},
	}

	for _, tt := range tests {
		v := Vector(tt.x)
		u := Vector(tt.y)
		ans := v.Dot(u)

		if ans != tt.ans {
			t.Errorf("x.Dot(y) = %v, expected %v", ans, tt.ans)
		}
	}
}

func BenchmarkDotFloat32(b *testing.B) {
	for _, bm := range []int{10, 50, 100, 150, 200, 250, 300, 350, 400, 450, 500} {
		b.Run(fmt.Sprintf("test with dimension %d", bm), func(b *testing.B) {
			b.ReportAllocs()

			vData := data(bm)
			uData := data(bm)
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
		dim int
		a   float32
	}{}
	for i, dim := range []int{10, 50, 100, 150, 200, 250, 300, 350, 400, 450, 500} {
		benchmarks = append(benchmarks, struct {
			dim int
			a   float32
		}{
			dim: dim,
			a:   float32(i * 10),
		})
	}
	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("test with dimension %d and a %.2f", bm.dim, bm.a), func(b *testing.B) {
			b.ReportAllocs()

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
	for _, bm := range []int{10, 50, 100, 150, 200, 250, 300, 350, 400, 450, 500} {
		b.Run(fmt.Sprintf("test with dimension %d", bm), func(b *testing.B) {
			b.ReportAllocs()

			d := data(bm)
			for i := 0; i < b.N; i++ {
				v := Vector(d)
				v.Norm()
			}
		})
	}
}
