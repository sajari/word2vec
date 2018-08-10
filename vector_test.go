package word2vec

import (
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
