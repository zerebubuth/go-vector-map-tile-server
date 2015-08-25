package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/zerebubuth/govecamole"
	"net/http"
)

type response struct {
	tile []byte
	err  error
}

type request struct {
	z, x, y int
	reply   chan<- response
}

// renderTile renders a single tile
func renderTile(z, x, y int, m *govecamole.VecMap) ([]byte, error) {
	req, err := govecamole.NewRequestZXY(256, 256, z, x, y)
	if err != nil {
		return nil, err
	}
	defer req.Close()

	opts, err := govecamole.DefaultOptions()
	if err != nil {
		return nil, err
	}
	defer opts.Close()

	var buf bytes.Buffer
	err = govecamole.Render(&buf, m, req, opts)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// renderLoop listens on a channel for requests, renders them and writes the
// result back on the channel in the request.
func renderLoop(ch <-chan request, m *govecamole.VecMap) {
	// see note in startRenderer about this.
	defer m.Close()

	for req := range ch {
		var r response
		r.tile, r.err = renderTile(req.z, req.x, req.y, m)
		req.reply<- r
	}
}

// startRenderer starts a single renderer, returning an error if the setup
// didn't work right.
func startRenderer(ch <-chan request, configFile string) error {
	m, err := govecamole.New(256, 256)
	if err != nil {
		return err
	}
	// note: not deferring m.Close() here, as we want ownership passed to
	// the goroutine. this probably isn't the best way of doing this, so
	// TODO: figure out how to do this in a nicer way, perhaps on the
	// goroutine itself?

	err = m.LoadFile(configFile, true, configFile)
	if err != nil {
		m.Close()
		return err
	}

	go renderLoop(ch, m)

	return nil
}

// setupVecMaps sets up numProcs mapnik objects and spawns goroutines to handle
// creating vector tiles for each of them. this returns a channel to write
// requests to.
func setupVecMaps(configFile string, numProcs int) (chan<- request, error) {
	err := govecamole.RegisterDefaultDatasources()
	if err != nil {
		return nil, err
	}

	ch := make(chan request)

	for i := 0; i < numProcs; i++ {
		err = startRenderer(ch, configFile)
		if err != nil {
			close(ch)
			return nil, err
		}
	}

	return ch, nil
}

// handleRequest creates a request and sends a request to a mapnik object and
// handles writing the response (tile or error) back to the client.
func handleRequest(z, x, y int, pool chan<- request, writer http.ResponseWriter) error {
	ch := make(chan response)
	req := request{z: z, x: x, y: y, reply: ch}
	pool <- req
	res := <-ch
	if res.err != nil {
		return res.err
	}
	_, err := writer.Write(res.tile)
	return err
}

// parsePath parses a z/x/y.fmt path string into z, x & y components.
func parsePath(p string) (int, int, int, error) {
	var z, x, y int
	var f string

	n, err := fmt.Sscanf(p, "/%d/%d/%d.%s", &z, &x, &y, &f)
	if err != nil {
		return 0, 0, 0, err
	}

	if n != 4 {
		return -1, 0, 0, errors.New("Expecting a path of format /z/x/y.fmt, but didn't match it.")
	}

	if z < 0 {
		return -1, 0, 0, errors.New("Zoom level must be non-negative.")
	}

	if z > 30 {
		return -1, 0, 0, errors.New("Zoom levels > 30 are not supported.")
	}

	if x < 0 {
		return -1, 0, 0, errors.New("X coordinate must be non-negative.")
	}

	if y < 0 {
		return -1, 0, 0, errors.New("Y coordinate must be non-negative.")
	}

	maxcoord := 1 << uint(z)
	if x >= maxcoord {
		return -1, 0, 0, errors.New("X coordinate out of range at this zoom.")
	}

	if y >= maxcoord {
		return -1, 0, 0, errors.New("Y coordinate out of range at this zoom.")
	}

	return z, x, y, nil
}

type vecMapsHandler struct {
	pool chan<- request
}

func (self *vecMapsHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	z, x, y, err := parsePath(req.URL.Path)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	err = handleRequest(z, x, y, self.pool, rw)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (self *vecMapsHandler) Close() {
	close(self.pool)
}

func NewVecMapsHandler(configFile string, numProcs int) (*vecMapsHandler, error) {
	ch, err := setupVecMaps(configFile, numProcs)
	if err != nil {
		return nil, err
	}

	h := new(vecMapsHandler)
	h.pool = ch
	return h, nil
}
