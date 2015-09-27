# word2vec

word2vec provides functionality for loading binary [word2vec](https://code.google.com/p/word2vec) models, and basic manipulation.

    $ go get github.com/sajari/word2vec/...

This will build the `word-calc` tool (in `$GOPATH/bin`), which lets you to do basic calculations with lists of query words.  For instance: `vec(king) - vec(man) + vec(woman)` would be equivalent to:

    $ word-calc -model /path/to/model.bin -add king,woman -sub man
