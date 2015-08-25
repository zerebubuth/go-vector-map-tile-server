# Go Vector Map Tile Server

A tile server using [vecamole](https://github.com/zerebubuth/vecamole) and [Mapnik](http://mapnik.org) to render vector tiles from a Mapnik data source and configuration.

## Installation

First, install [vecamole](https://github.com/zerebubuth/vecamole). That's the only hard bit. From then, it should be as simple as:

```
go install github.com/zerebubuth/go-vector-map-tile-server
```

## Running it

You can run it from your `$GOPATH` like this:

```
bin/go-vector-map-tile-server
```

It has a few command line options:

* `help`: Print a short usage message.
* `numProcs`: The number of Mapnik processes to run. More processes will allow more parallelism, but also consume more resources. The default is the same as `$GOMAXPROCS`.
* `port`: The port number to start the HTTP server listening on. Default 8080.
* `styleFile`: The Mapnik style file to load and serve. Default `map.xml`.

And will respond to HTTP requests of the form: `http://localhost:8080/$z/$x/$y.$fmt`, or whatever port you ended up running the server on. Note that, at the moment, the format has no effect - you always get back protocol buffers Mapnik vector tiles.
