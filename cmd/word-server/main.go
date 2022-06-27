/*
word-server creates an HTTP server which exports endpoints for querying a word2vec model.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"code.sajari.com/word2vec"
)

var listen, modelPath string
var lazy bool

func init() {
	flag.StringVar(&listen, "listen", "localhost:1234", "bind `address` for HTTP server")
	flag.StringVar(&modelPath, "model", "", "`path` to binary model data")
	flag.BoolVar(&lazy, "lazy", false, "whether to lazy load model, without cos-n support, defaults to true, which eager load it")
}

func main() {
	flag.Parse()

	if modelPath == "" {
		fmt.Println("must specify -model; see -h for more details")
		os.Exit(1)
	}

	if lazy {
		log.Println("Lazy loading model...")
	} else {
		log.Println("Loading model...")
	}
	f, err := os.Open(modelPath)
	if err != nil {
		fmt.Printf("error opening binary model data file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var m *word2vec.Model

	if lazy {
		m, err = word2vec.LazyFromReader(f)
	} else {
		m, err = word2vec.FromReader(f)
	}
	if err != nil {
		fmt.Printf("error reading binary model data: %v\n", err)
		f.Close()
		os.Exit(1)
	}

	ms := word2vec.NewServer(word2vec.NewCache(m))

	log.Printf("Server listening on %v", listen)
	log.Println("Hit Ctrl-C to quit.")

	log.Fatal(http.ListenAndServe(listen, ms))
}
