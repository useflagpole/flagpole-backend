package main

import (
	"flag"
	"fmt"
)

type Flags struct {
	port      string
	jwtSecret string
}

func (f *Flags) parseFlags() {
	port      := flag.String("port", "4000", "Port in which the Flagpole API will serve")
	jwtSecret := flag.String("jwt-secret", "change-me", "Secret key used to sign JWT tokens")
	flag.Parse()

	f.port      = *port
	f.jwtSecret = *jwtSecret
}

func (f *Flags) printFlags() {
	fmt.Printf("%+v\n", f)
}
