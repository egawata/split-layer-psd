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

var fName string
var outDir string

func init() {
	flag.StringVar(&fName, "file", "", "psd filename")
	flag.StringVar(&fName, "f", "", "psd filename (shorthand)")
	oUsage := `output directory (default: same directory with original psd)`
	flag.StringVar(&outDir, "out", "", oUsage)
	flag.StringVar(&outDir, "o", "", oUsage+` (shorthand)`)
}

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
	if fName == "" {
		log.Fatal("filename(-f) required")
	}

	if outDir == "" {
		outDir = filepath.Dir(fName)
	}
	if _, err := os.Stat(outDir); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("directory %s does not exist. create...\n", outDir)
		if err := os.MkdirAll(outDir, 0777); err != nil {
			log.Fatal(err)
		}
	}

	file, err := os.Open(fName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := psd.Decode(file, &psd.DecodeOptions{SkipMergedImage: true})
	if err != nil {
		log.Fatal(err)
	}
	for i, layer := range img.Layer {
		fn := filepath.Join(outDir, fmt.Sprintf("%03d", i))
		if err = processLayer(fn, layer.Name, &layer); err != nil {
			log.Printf("[WARN] %s: %v\n", fn, err)
		}
	}
}
