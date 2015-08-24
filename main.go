package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"github.com/zerebubuth/govecamole"
)

var outputFile = flag.String("outputFile", "tile.pbf", "The file name to write the output to.")

func main() {
	flag.Parse()

	err := govecamole.RegisterDefaultDatasources()
	if err != nil {
		fmt.Printf("Ooops, can't register datasources: %s\n", err.Error())
		return
	}

	m, err := govecamole.New(256, 256)
	if err != nil {
		fmt.Printf("Ooops, got an error: %s\n", err.Error())
		return
	}
	defer m.Close()

	sampleconf := `<!DOCTYPE Map>
<Map srs="+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs">
  <Style name="point">
  </Style>
  <Layer name="point" srs="+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs">
    <StyleName>point</StyleName>
    <Datasource>
    <Parameter name="type">csv</Parameter>
    <Parameter name="inline">
type,WKT
point,"POINT (0 0)"
</Parameter>
    </Datasource>
  </Layer>
</Map>`

	err = m.LoadString(sampleconf, true, "sampleconf")
	if err != nil {
		fmt.Printf("Unable to load sample config string into map: %s\n", err.Error())
		return
	}

	req, err := govecamole.NewRequestZXY(256, 256, 0, 0, 0)
	if err != nil {
		fmt.Printf("Unable to create a request: %s\n", err.Error())
		return
	}
	defer req.Close()

	opts, err := govecamole.DefaultOptions()
	if err != nil {
		fmt.Printf("Unable to create default options: %s\n", err.Error())
		return
	}
	defer opts.Close()

	var buf bytes.Buffer
	err = govecamole.Render(&buf, m, req, opts)
	if err != nil {
		fmt.Printf("Unable render tile: %s\n", err.Error())
		return
	}

	fmt.Printf("Got tile size=%v\n", buf.Len())

	file, err := os.Create(*outputFile)
	if err != nil {
		fmt.Printf("Unable to create output file %v: %s\n", *outputFile, err.Error())
		return
	}
	defer file.Close()

	_, err = io.Copy(file, &buf)
	if err != nil {
		fmt.Printf("Unable to copy tile to file: %s\n", err.Error())
		return
	}
}
