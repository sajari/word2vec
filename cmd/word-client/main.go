/*
word2vec-client is a tool which queries a `word-server` HTTP server to do computations with a word2vec
model.
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

var addr string
var addListA, subListA string
var addListB, subListB string
var sim bool
var n int

func init() {
	flag.StringVar(&addr, "addr", "localhost:1234", "server `address`")
	flag.StringVar(&addListA, "addA", "", "comma separated list of model `words` to add to the target vector A")
	flag.StringVar(&subListA, "subA", "", "comma separated list of model `words` to subtract from the target vector A")
	flag.StringVar(&addListB, "addB", "", "comma separated list of model `words` to add to the target vector B")
	flag.StringVar(&subListB, "subB", "", "comma separated list of model `words` to subtract from the target vector B")
	flag.BoolVar(&sim, "sim", false, "similarity query")
	flag.IntVar(&n, "n", 10, "return `N` similar items in similarity query")
}

func makeExpr(addList, subList string) (word2vec.Expr, error) {
	if addList == "" && subList == "" {
		return word2vec.Expr{}, fmt.Errorf("must specify 'add' and/or 'sub' component for each target vector; see -h for more details")
	}

	result := word2vec.Expr{}
	if addList != "" {
		for _, w := range strings.Split(addList, ",") {
			result.Add(1, w)
		}
	}
	if subList != "" {
		for _, w := range strings.Split(subList, ",") {
			result.Add(-1, w)
		}
	}
	return result, nil
}

func main() {
	flag.Parse()

	if addr == "" {
		fmt.Println("must specify -addr; see -h for more details")
		os.Exit(1)
	}

	exprA, err := makeExpr(addListA, subListA)
	if err != nil {
		fmt.Printf("error creating target vector for 'A': %v\n", err)
		os.Exit(1)
	}

	if sim {
		c := word2vec.Client{Addr: addr}
		r, err := c.CosN(exprA, n)
		if err != nil {
			fmt.Printf("error looking up similar items: %v\n", err)
			os.Exit(1)
		}
		for _, x := range r {
			fmt.Printf("%9f %#v\n", x.Score, x.Word)
		}
		return
	}

	exprB, err := makeExpr(addListB, subListB)
	if err != nil {
		fmt.Printf("error creating target vector for 'B': %v\n", err)
		os.Exit(1)
	}

	c := word2vec.Client{Addr: addr}

	start := time.Now()
	v, err := c.Cos(exprA, exprB)
	totalTime := time.Since(start)
	if err != nil {
		fmt.Printf("error looking up similarity: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("cosine similarity: %v (took: %v)\n", v, totalTime)
}
