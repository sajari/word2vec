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
var multiQuery string
var verbose bool
var n int

func init() {
	flag.StringVar(&path, "p", "", "path to binary model data")
	flag.StringVar(&multiQuery, "m", "", "comma separated list of model words to query at the same time")
	flag.StringVar(&addList, "a", "", "comma separated list of model words to add to the target vector")
	flag.StringVar(&subList, "s", "", "comma separated list of model words to subtract from the target vector")
	flag.BoolVar(&verbose, "v", false, "show verbose output")
	flag.IntVar(&n, "n", 10, "number of most similar matches to show")
}

func main() {
	flag.Parse()

	if path == "" {
		fmt.Println("must specify -p; see -h for more details")
		os.Exit(1)
	}

	if addList == "" && subList == "" && multiQuery == "" {
		fmt.Println("must specify -a, -s, or -m; see -h for more details")
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

	// TODO(dhowden): Tidy this up, it's rather hacked in here!
	if multiQuery != "" {
		multiWords := strings.Split(multiQuery, ",")
		vecs := m.Vectors(multiWords)

		before := time.Now()
		res := word2vec.MultiSimN(m, vecs, 10)
		fmt.Println("Total time:", time.Since(before))
		fmt.Println(res)
		return
	}

	expr := word2vec.Expr{}
	if addList != "" {
		word2vec.AddAll(expr, 1, strings.Split(addList, ","))
	}
	if subList != "" {
		word2vec.AddAll(expr, -1, strings.Split(subList, ","))
	}

	if verbose {
		fmt.Printf("Expr: %#v\n", expr)
	}

	v, err := expr.Eval(m)
	if err != nil {
		fmt.Printf("error creating target vector: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("Target vector: %#v\n", v)
	}

	before := time.Now()
	pairs := m.SimN(v, n)
	if err != nil {
		fmt.Printf("error finding most similar: %v\n", err)
		os.Exit(1)
	}
	after := time.Now()

	if verbose {
		fmt.Println("Total time: ", after.Sub(before))
	}

	for _, k := range pairs {
		fmt.Printf("%9f\t%#v\n", k.Score, k.Word)
	}
}
