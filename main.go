package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
)

var numProcs = flag.Int("numProcs", runtime.GOMAXPROCS(0), "The number of Mapnik processes to run. More processes will allow more parallelism, but also consume more resources.")
var styleFile = flag.String("styleFile", "map.xml", "The Mapnik style file to load and serve.")
var port = flag.Int("port", 8080, "The port number to start the HTTP server listening on.")
var printHelp = flag.Bool("help", false, "Print a short usage message.")

func main() {
	flag.Parse()

	if *printHelp {
		fmt.Fprintf(os.Stderr, "Usage for go-vector-map-tile-server:\n\n")
		flag.PrintDefaults()
		return
	}

	h, err := NewVecMapsHandler(*styleFile, *numProcs)
	if err != nil {
		fmt.Printf("Ooops, start vector maps handler: %s\n", err.Error())
		return
	}
	defer h.Close()

	addr := fmt.Sprintf(":%d", *port)
	s := &http.Server{
		Addr: addr,
		Handler: h,
	}

	fmt.Printf("About to start serving tiles on %v.\n", s.Addr)
	log.Fatal(s.ListenAndServe())
}
