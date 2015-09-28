package partition

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)


// NotFoundError is an error returned from Partition functions when an input
// word can not be found.
type NotFoundError struct {
	Word string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("word not found: %v", e.Word)
}

// Partition represents an arrangement of words from a word2vec model into a partition of
// equivalency classes.
type Partition struct {
	words   map[string]int
	classes map[int][]string
	size    int
}

// NewPartition returns a Partition instance which is loaded with data from r (assumed
// to be of the output format from the word2vec command with -classes arg).
func NewPartition(r io.Reader) (*Partition, error) {
	scanner := bufio.NewScanner(r)
	words := make(map[string]int)
	classes := make(map[int][]string)

	i := 0
	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("[line: %d] expected 2 fields, instead got: %v", i+1, len(fields))
		}

		w := fields[0]
		c, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("[line: %d] error parsing integer %#v: %v", i+1, fields[1], err)
		}

		words[w] = c
		classes[c] = append(classes[c], w)
		i++
	}

	err := scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("[line: %d] scanner error: %v", i+1, err)
	}

	return &Partition{
		words:   words,
		classes: classes,
	}, nil
}

// Classes returns the number of classes in the partition.
func (p *Partition) Classes() int {
	return len(p.classes)
}

// Size returns the number of words in the partition.
func (p *Partition) Size() int {
	return len(p.words)
}

// Class returns the internal index used to represent he equivalence class for w.
func (p *Partition) Class(w string) (int, error) {
	n, ok := p.words[w]
	if !ok {
		return 0, &NotFoundError{w}
	}
	return n, nil
}

// Equiv returns true iff w and v are equivalent.
func (p *Partition) Equiv(w, v string) (bool, error) {
	i, ok := p.words[w]
	if !ok {
		return false, &NotFoundError{w}
	}
	j, ok := p.words[v]
	if !ok {
		return false, &NotFoundError{v}
	}
	return i == j, nil
}

// EquivClass returns the equivalence class of w as the list of strings which are
// equivalent to it.
func (p *Partition) EquivClass(w string) ([]string, error) {
	i, ok := p.words[w]
	if !ok {
		return nil, &NotFoundError{w}
	}
	return p.classes[i], nil
}

func (p *Partition) EquivClassIndex(i int) ([]string, error) {
	c, ok := p.classes[i]
	if !ok {
		return nil, fmt.Errorf("equivalence class not found: %v", i)
	}
	return c, nil
}
