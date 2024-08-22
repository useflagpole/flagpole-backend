package main

import (
	"flag"
	"fmt"
)

type Flags struct {
	port string
}

func (f *Flags) parseFlags() {
	port  := flag.String("port", "4000", "Port in which the Flagpole API will serve")
	flag.Parse()

	f.port = *port

}

func (f *Flags) printFlags() {
	fmt.Printf("%+v\n", f)
}
