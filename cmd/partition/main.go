/*
partition is a tool which reads word2vec classes output and allows you to query the data.

   $ partition -p /path/to/classes.txt -q something
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"code.sajari.com/word2vec/partition"
)

var path string
var query string
var class int
var dist bool
var info bool

func init() {
	flag.StringVar(&path, "p", "", "`path` to classes text data")
	flag.StringVar(&query, "q", "", "fetch the class for `word`")
	flag.IntVar(&class, "c", 0, "`class` to fetch")
	flag.BoolVar(&dist, "d", false, "show distribution of classes")
	flag.BoolVar(&info, "i", false, "only show info about the word class and partition")
}

type classCount struct {
	class, total int
}

type classCountSlice []classCount

func (c classCountSlice) Less(i, j int) bool {
	return c[i].total < c[j].total
}

func (c classCountSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c classCountSlice) Len() int {
	return len(c)
}

func main() {
	flag.Parse()

	if path == "" {
		fmt.Println("must specify -p; see -h for more details")
		os.Exit(1)
	}

	if query == "" && class == 0 && !dist {
		fmt.Println("must specify -q, -c or -d; see -h for more details")
		os.Exit(1)
	}

	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("error opening classes data file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	p, err := partition.NewPartition(f)
	if err != nil {
		fmt.Printf("error reading classes data: %v\n", err)
		os.Exit(1)
	}

	if dist {
		n := p.Classes()
		s := make([]classCount, 0, n)
		for i := 0; i < n; i++ {
			cl, err := p.EquivClassIndex(i)
			if err != nil {
				fmt.Printf("no class for index: %v\n", i)
				os.Exit(1)
			}
			s = append(s, classCount{i, len(cl)})
		}

		sort.Sort(sort.Reverse(classCountSlice(s)))

		fmt.Println("class\t\ttotal")
		for _, cc := range s {
			fmt.Printf("%v\t\t%v\n", cc.class, cc.total)
		}
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
