package main

import (
	"errors"
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
		fn := fmt.Sprintf("%s_%03d", filename, i)
		if err := processLayer(fn, layerName+"/"+ll.Name, &ll); err != nil {
			return err
		}
	}
	if !l.HasImage() {
		return nil
	}
	if l.Picker.Bounds().Empty() {
		fmt.Printf("[warn] empty layer: %s\n", layerName)
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
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("directory %s does not exist. create...\n", dir)
		if err := os.MkdirAll(dir, 0777); err != nil {
			log.Fatal(err)
		}
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
		fn := filepath.Join(dir, fmt.Sprintf("%03d", i))
		if err = processLayer(fn, layer.Name, &layer); err != nil {
			log.Printf("[WARN] %s: %v\n", fn, err)
		}
	}
}
