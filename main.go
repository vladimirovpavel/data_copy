package main

import (
	"flag"
	"fmt"
)

var src, dst string
var offset, count uint64

func init() {
	flag.StringVar(&src, "src", "", "source copy file")
	flag.StringVar(&dst, "dst", "", "destination copy file")
	flag.Uint64Var(&offset, "offset", 0, "offset in input file")
	flag.Uint64Var(&count, "count", 0, "count of bytes to write")
}

func main() {
	flag.Parse()
	if src == "" || dst == "" {
		fmt.Printf("Usage:\ndata_copy --src <filename> --dst <filename> " +
			"[--offset <digit>] [--count <digit>]\n" +
			"\tsrc - path to source filename\n\tdst - path to destination file\n" +
			"\toffset - offset in source from start\n" +
			"\tcount - count of bytes to write (if 0 - all file)\n")
	}
	Copy(src, dst, offset, count)
}

//dst path, but not name - name same the src
//progressbarr from github.com/cheggaaa/pb
