package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/oov/psd"
)

var (
	fName  = flag.String("f", "", "psd file")
	outDir = flag.String("o", "", "output directory (default: same directory with original psd)")
)

func processLayer(filename string, layerName string, l *psd.Layer) error {
	for i, ll := range l.Layer {
		if err := processLayer(
			fmt.Sprintf("%s_%03d", filename, i),
			layerName+"/"+ll.Name, &ll); err != nil {
			return err
		}
	}
	if !l.HasImage() {
		return nil
	}

	fmt.Printf("%s -> %s.png\n", layerName, filename)

	out, err := os.Create(fmt.Sprintf("%s.png", filename))
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, l.Picker)
}

func main() {
	flag.Parse()
	if *fName == "" {
		log.Fatal("filename(-f) required")
	}

	dir := *outDir
	if dir == "" {
		dir = filepath.Dir(*fName)
	}

	file, err := os.Open(*fName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := psd.Decode(file, &psd.DecodeOptions{SkipMergedImage: true})
	if err != nil {
		log.Fatal(err)
	}
	for i, layer := range img.Layer {
		if err = processLayer(filepath.Join(dir, fmt.Sprintf("%03d", i)), layer.Name, &layer); err != nil {
			log.Fatal(err)
		}
	}
}
