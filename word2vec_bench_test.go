package word2vec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var filename = "" //set this to the path of your model file to a more precise bench

func BenchmarkLazyModel(b *testing.B) {

	var (
		f   *os.File
		err error
	)

	if filename == "" {
		f, err = os.Create(filepath.Join(b.TempDir(), "go-word2vec-bench.bin"))
		if err != nil {
			b.Fatal(err)
		}
		defer f.Close()

		dim := 300
		numWords := 26 * 26 * 26 * 26

		buf := bytes.NewBuffer(make([]byte, 0, 500_000))

		_, err := fmt.Fprintln(buf, numWords, dim)
		if err != nil {
			b.Fatal(err)
		}

		emptyVector := make([]float32, dim)

		for i := 'a'; i <= 'z'; i++ {
			for j := 'a'; j <= 'z'; j++ {
				for k := 'a'; k <= 'z'; k++ {
					for l := 'a'; l <= 'z'; l++ {
						_, err := fmt.Fprintf(buf, "%s%s%s%s ", string(i), string(j), string(k), string(l))
						if err != nil {
							b.Fatal("unexpected error writing word")
						}
						err = binary.Write(buf, binary.LittleEndian, emptyVector)
						if err != nil {
							b.Fatal("unexpected error writing vector")
						}
						_, err = fmt.Fprintf(buf, "\n")
						if err != nil {
							b.Fatal("unexpected error writing new line")
						}
					}
				}
			}
		}
		bb := buf.Bytes()

		_, err = f.Write(bb)
		if err != nil {
			b.Fatal(err)
		}

		_, err = f.Seek(0, 0)
		if err != nil {
			b.Fatal(err)
		}

	} else {
		f, err = os.Open(filename)
		if err != nil {
			b.Fatal(err)
		}
		defer f.Close()
	}

	b.Run("InitializeEager", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := FromReader(f)
			if err != nil {
				b.Fatal(err)
			}
			_, err = f.Seek(0, 0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("InitializeLazy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := LazyFromReader(f)
			if err != nil {
				b.Fatal(err)
			}
			_, err = f.Seek(0, 0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	m, err := FromReader(f)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("LoadVectorEager", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			r := m.Map([]string{"abcd"})
			if len(r) == 0 {
				b.Fatal("empty return")
			}
		}

	})

	_, err = f.Seek(0, 0)
	if err != nil {
		b.Fatal(err)
	}

	l, err := LazyFromReader(f)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("LoadVectorLazy", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			r := l.Map([]string{"abcd"})
			if len(r) == 0 {
				b.Fatal("empty return")
			}
		}

	})

}
