# word2vec
[![Build Status](https://travis-ci.org/sajari/word2vec.svg?branch=master)](https://travis-ci.org/sajari/word2vec)

word2vec provides functionality for loading binary [word2vec](https://code.google.com/p/word2vec) models and performing cosine similarity of combinations of words.

    $ go get github.com/sajari/word2vec/...

This will fetch the code and build several tools into `$GOPATH/bin`, including `word-calc` which performs basic calculations with lists of query words.  For instance: `vec(king) - vec(man) + vec(woman)` would be equivalent to:

    $ word-calc -model /path/to/model.bin -add king,woman -sub man

Alternatively, you can use the `word2vec` package directly:

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
matches, err := m.CosN(expr, 1)
if err != nil {
	log.Fatalf("error evaluating cosine similarity: %v", err)
}
```

##  word2vec server

The `word-server` tool (see `cmd/word-server`) creates an HTTP server which wraps a word2vec model which can be queried from Go using a [Client](http://godoc.org/github.com/sajari/word2vec#Client), or using the `word-client` tool (see `cmd/word-client`).

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
