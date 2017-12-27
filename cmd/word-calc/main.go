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

	"code.sajari.com/word2vec"
)

var path string
var addList, subList string
var multiQuery string
var verbose bool
var n int

func init() {
	flag.StringVar(&path, "model", "", "`path` to binary model data")
	flag.StringVar(&multiQuery, "words", "", "comma separated list of model `words` to query at the same time")
	flag.StringVar(&addList, "add", "", "comma separated list of model `words` to add to the target vector")
	flag.StringVar(&subList, "sub", "", "comma separated list of model `words` to subtract from the target vector")
	flag.BoolVar(&verbose, "v", false, "show verbose output")
	flag.IntVar(&n, "n", 10, "show `N` similar matches")
}

func main() {
	flag.Parse()

	if path == "" {
		fmt.Println("must specify -model; see -h for more details")
		os.Exit(1)
	}

	if addList == "" && subList == "" && multiQuery == "" {
		fmt.Println("must specify -add, -sub, or -words; see -h for more details")
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
		var exprs []word2vec.Expr

		multiWords := strings.Split(multiQuery, ",")
		for _, w := range multiWords {
			e := word2vec.Expr{}
			e.Add(1, w)

			exprs = append(exprs, e)
		}

		before := time.Now()
		res, err := word2vec.MultiCosN(m, exprs, 10)
		if err != nil {
			fmt.Printf("error retrieving multi cos: %v", err)
			os.Exit(1)
		}
		fmt.Println("Total time:", time.Since(before))
		fmt.Println(res)
		return
	}

	expr := word2vec.Expr{}
	if addList != "" {
		word2vec.Add(expr, 1, strings.Split(addList, ","))
	}
	if subList != "" {
		word2vec.Add(expr, -1, strings.Split(subList, ","))
	}

	if verbose {
		fmt.Printf("Expr: %#v\n", expr)
	}

	if verbose {
		v, err := expr.Eval(m)
		if err != nil {
			fmt.Printf("error creating target vector: %v", err)
			os.Exit(1)
		}
		fmt.Printf("Target vector: %#v\n", v)
	}

	before := time.Now()
	pairs, err := m.CosN(expr, n)
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
