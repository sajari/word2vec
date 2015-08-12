package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sajari/word2vec"
)

var listen string
var path string

func init() {
	flag.StringVar(&listen, "listen", "", "bind address for HTTP server")
	flag.StringVar(&path, "p", "", "path to binary model data")
}

func main() {
	flag.Parse()

	if path == "" {
		fmt.Println("must specify -p; see -h for more details")
		os.Exit(1)
	}

	log.Println("Loading model...")
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("error opening binary model data file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	m, err := word2vec.FromReader(f)
	if err != nil {
		fmt.Printf("error reading binary model data: %v\n", err)
		os.Exit(1)
	}

	ms := word2vec.ModelServer{m}

	http.HandleFunc("/most-sim", ms.HandleMostSimQuery)
	http.HandleFunc("/sim", ms.HandleSimQuery)
	
	log.Printf("Server listening on %v", listen)
	log.Println("Hit Ctrl-C to quit.")
	
	log.Fatal(http.ListenAndServe(listen, nil))
}
