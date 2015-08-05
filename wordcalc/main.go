/*
wordcalc is a tool which reads word2vec binary models and allows you to do basic calculations
with lists of query words.  For instance vec(king) - vec(man) + vec(woman) would be equivalent
to:

   $ wordcalc -p /path/to/model.bin -a king,woman -s man
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sajari/word2vec"
)

var path string
var addList, subList string
var verbose bool
var n int

func init() {
	flag.StringVar(&path, "p", "", "path to binary model data")
	flag.StringVar(&addList, "a", "", "comma separated list of model words to add to the target vector")
	flag.StringVar(&subList, "s", "", "comma separated list of model words to subtract form the target vector")
	flag.BoolVar(&verbose, "v", false, "show verbose output")
	flag.IntVar(&n, "n", 10, "number of most similar matches to show")
}

func main() {
	flag.Parse()

	if path == "" {
		fmt.Println("must provide path to binary model data")
		os.Exit(1)
	}

	if addList == "" && subList == "" {
		fmt.Println("must specify -add or -sub")
		os.Exit(1)
	}

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

	var adds, subs []string
	if addList != "" {
		adds = strings.Split(addList, ",")
	}
	if subList != "" {
		subs = strings.Split(subList, ",")
	}

	if verbose {
		fmt.Printf("Add:        %#v\n", adds)
		fmt.Printf("Subtract:   %#v\n", subs)
	}

	v, err := m.Eval(adds, subs)
	if err != nil {
		fmt.Printf("error creating target vector: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("Target vector: %#v\n", v)
	}

	before := time.Now()
	pairs, err := m.MostSimilar(v, n)
	if err != nil {
		fmt.Printf("error finding most similar: %v\n", err)
		os.Exit(1)
	}
	after := time.Now()

	if verbose {
		fmt.Println("Total time: ", after.Sub(before))
	}

	for _, k := range pairs {
		fmt.Printf("%v\t %#v\n", k.Score, k.Word)
	}
}
