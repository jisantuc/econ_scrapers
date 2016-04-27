econ scrapers
=============

_scrapes EconPapers_, do you get it

## Installation

Run `go get github.com/jisantuc/econ_scrapers` then `go install econ_scrapers` (or just `go install` if you're in `$GOPATH/src/github.com/jisantuc/econ_scrapers` but probably you knew that)

There aren't tests right now because I'm a serious rook at go and bad at thinking about testing for scrapers. Someday.

## Running

After `go install`ing, run `econ_scrapers` to get a nice JSON with a bunch of abstracts and JEL codes in whatever directory you're in.

## Caveats

I worked on this as a problem to solve with go. If there are improvements (I'm sure there are improvements), feel free to submit issues, but I might not get to them quickly because I really just wanted a dataset to play with for some `gensim` fun.
