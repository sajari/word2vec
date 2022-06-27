# word2vec
[![Build Status](https://travis-ci.org/sajari/word2vec.svg?branch=master)](https://travis-ci.org/sajari/word2vec)
[![GoDoc](https://godoc.org/code.sajari.com/word2vec?status.svg)](https://godoc.org/code.sajari.com/word2vec)

word2vec is a Go package which provides functions for querying word2vec models (see [https://code.google.com/p/word2vec](https://code.google.com/p/word2vec)).  Any binary word2vec model file can be loaded and queried.

## Requirements

- [Go 1.4+](http://golang.org/dl/) (only tested on 1.4+)
- A [word2vec](https://code.google.com/p/word2vec) model (binary format)

## Installation

If you haven't setup Go before, you need to first set a `GOPATH` (see [https://golang.org/doc/code.html#GOPATH](https://golang.org/doc/code.html#GOPATH)).

To fetch and build the code:

    $ go get code.sajari.com/word2vec/...

This will build the command line tools (in particular `word-calc`, `word-server`, `word-client`) into `$GOPATH/bin` (assumed to be in your `PATH` already).

## Usage

### word-calc

The `word-calc` tool is a quick way to perform basic word calculations on a word2vec model.  For instance: `vec(king) - vec(man) + vec(woman)` would be equivalent to:

    $ word-calc -model /path/to/model.bin -add king,woman -sub man

See `word-calc -h` for full more details.  Note that `word-calc` first loads the model every time,  and so can appear to be quite slow. Use `word-server` and `word-client` to get better performance when running multiple queries on the same model.

###  word-server and word-client

The `word-server` tool (see `cmd/word-server`) creates an HTTP server which wraps a word2vec model which can be queried from Go using a [Client](http://godoc.org/code.sajari.com/word2vec#Client), or using the `word-client` tool (see `cmd/word-client`).

    $ word-server -model /path/to/model.bin -listen localhost:1234

A simple code example using `Client`:

```go
c := word2vec.Client{Addr: "localhost:1234"}

// Create an expression.
expr := word2vec.Expr{}
expr.Add(1, "king")
expr.Add(-1, "man")
expr.Add(1, "woman")

// Find the most similar result by cosine similarity.
matches, err := c.CosN(expr, 1)
if err != nil {
	log.Fatalf("error evaluating cosine similarity: %v", err)
}
```

### API Example
Alternatively you can interact with a word2vec model directly in your code:

```go
// Load the model from an io.Reader (i.e. a file).
model, err := word2vec.FromReader(r)
if err != nil {
	log.Fatalf("error loading model: %v", err)
}

// Create an expression.
expr := word2vec.Expr{}
expr.Add(1, "king")
expr.Add(-1, "man")
expr.Add(1, "woman")

// Find the most similar result by cosine similarity.
matches, err := model.CosN(expr, 1)
if err != nil {
	log.Fatalf("error evaluating cosine similarity: %v", err)
}

// Create two expressions.
x := word2vec.Expr{"king": 1.0}
y := word2vec.Expr{"queen": 1.0}

// Compute similarity between the expressions
cosineSimilarity, err := m.Cos(x, y)
if err != nil {
    log.Fatalf("error evaluating cosine similarity: %v", err)
}

```

If you only wanna to compute similarity between some words, but not to find the n most similar of a given word,
you can use a lazy model, which initializes faster and uses less memory:

```go
// Lazy load the model from an io.Reader (i.e. a file).
model, err := word2vec.LazyFromReader(r)
if err != nil {
	log.Fatalf("error loading model: %v", err)
}

// Create two expressions.
x := word2vec.Expr{"king": 1.0}
y := word2vec.Expr{"queen": 1.0}

// Compute similarity between the expressions
cosineSimilarity, err := m.Cos(x, y)
if err != nil {
    log.Fatalf("error evaluating cosine similarity: %v", err)
}

```

Below is a benchmark between the normal and the lazy model using a model file with 456.976 (26 ^ 4) words of 300 dimensions.
You can run the benchmark with your model declaring the var ```filename``` of ```word2vec_bench_test.go``` to the path of your binary model.


```
BenchmarkLazyModel
BenchmarkLazyModel/InitializeEager
BenchmarkLazyModel/InitializeEager-4         	       1	4217573971 ns/op	1193506272 B/op	 1370948 allocs/op
BenchmarkLazyModel/InitializeLazy
BenchmarkLazyModel/InitializeLazy-4            	       4	 321390344 ns/op	  31409760 B/op	  456996 allocs/op
BenchmarkLazyModel/LoadVectorEager
BenchmarkLazyModel/LoadVectorEager-4                  4769024	       266.2 ns/op	     400 B/op	       2 allocs/op
BenchmarkLazyModel/LoadVectorLazy
BenchmarkLazyModel/LoadVectorLazy-4                    151197	        8163 ns/op	    3032 B/op	       6 allocs/op

4217573971ns = 4,21s
321390344ns  = 0,32s
1193506272B  = 1,19GB
31409760B    = 0,03GB(31MB)

```
