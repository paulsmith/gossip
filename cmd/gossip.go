package main

import (
	"flag"
	"log"
	"fmt"
	"os"
	
	"github.com/paulsmith/gossip"
)

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s src dest\n", os.Args[0])
		os.Exit(1)
	}
	site := gossip.NewSite(flag.Arg(0), flag.Arg(1))
	if err := site.Generate(); err != nil {
		log.Fatalf("error generating site: %v", err)
	}
}
