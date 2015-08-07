/*
partition is a tool which reads word2vec classes output and allows you to query the data.

   $ partition -p /path/to/classes.txt -q something
*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sajari/word2vec"
)

var path string
var query string
var class int
var info bool

func init() {
	flag.StringVar(&path, "p", "", "path to classes text data")
	flag.StringVar(&query, "q", "", "word to fetch class")
	flag.IntVar(&class, "c", 0, "class to fetch")
	flag.BoolVar(&info, "i", false, "only show info about the word class and partition")
}

func main() {
	flag.Parse()

	if path == "" {
		fmt.Println("must specify -p; see -h for more details")
		os.Exit(1)
	}

	if query == "" && class == 0 {
		fmt.Println("must specify -q or -c; see -h for more details")
		os.Exit(1)
	}

	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("error opening classes data file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	p, err := word2vec.NewPartition(f)
	if err != nil {
		fmt.Printf("error reading classes data: %v\n", err)
		os.Exit(1)
	}

	var c []string
	var cl int
	if query != "" {
		c, err = p.EquivClass(query)
		if err != nil {
			fmt.Printf("error fetching equivalence class: %v\n", err)
			os.Exit(1)
		}

		cl, err = p.Class(query)
		if err != nil {
			fmt.Printf("error fetching class index: %v\n", err)
			os.Exit(1)
		}
	}

	if class != 0 {
		c, err = p.EquivClassIndex(class)
		if err != nil {
			fmt.Printf("error fetching equivalence class: %v\n", err)
			os.Exit(1)
		}
		cl = class
	}

	fmt.Printf("class index:    %v\n", cl)
	if info {
		fmt.Printf("word:           %#v\n", query)
		fmt.Printf("words in class: %v\n", len(c))
		fmt.Printf("total classes   %v\n", p.Classes())
		fmt.Printf("total words     %v\n", p.Size())
		return
	}

	for _, w := range c {
		fmt.Printf("%#v\n", w)
	}
}
