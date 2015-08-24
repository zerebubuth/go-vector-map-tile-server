# Go Vector Map Tile Server

A tile server using [vecamole](https://github.com/zerebubuth/vecamole) and [Mapnik](http://mapnik.org) to render vector tiles from a Mapnik data source and configuration.

## Installation

First, install [vecamole](https://github.com/zerebubuth/vecamole). That's the only hard bit. From then, it should be as simple as:

```
go install github.com/zerebubuth/go-vector-map-tile-server
```

## Running it

Currently it just dumps a static `tile.pbf` in the current directory. Working in progress to make it more useful.
