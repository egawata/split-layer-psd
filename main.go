package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/oov/psd"
)

var fName string
var outDir string
var optBgcolor string
var optBgcolorWhite bool
var optKeepOriginalBound bool

var originalBound image.Rectangle

var col = map[string][]uint16{
	"black":   {0, 0, 0},
	"blue":    {0, 0, 65535},
	"red":     {65535, 0, 0},
	"magenta": {65535, 0, 65535},
	"green":   {0, 65535, 0},
	"cyan":    {0, 65535, 65535},
	"yellow":  {65535, 65535, 0},
	"white":   {65535, 65535, 65535},
}

func init() {
	flag.StringVar(&fName, "file", "", "psd filename")
	flag.StringVar(&fName, "f", "", "psd filename (shorthand)")
	oUsage := `output directory (default: same directory with original psd)`
	flag.StringVar(&outDir, "out", "", oUsage)
	flag.StringVar(&outDir, "o", "", oUsage+` (shorthand)`)
	bgcolorUsage := `fill background with color`
	flag.StringVar(&optBgcolor, "bgcolor", "", bgcolorUsage)
	flag.BoolVar(&optBgcolorWhite, "bw", false, "set bgcolor to white. shorthand for `-bgcolor white`")
	flag.BoolVar(&optKeepOriginalBound, "keep-original-bound", false, "keep original bound")
}

func parseBgcolor() color.RGBA64 {
	var bgcolor color.RGBA64
	if bg, ok := col[optBgcolor]; ok {
		return color.RGBA64{bg[0], bg[1], bg[2], 65535}
	}

	// color code like `f7ca94`
	if c, err := hex.DecodeString(optBgcolor); err == nil {
		f := func(b byte) uint16 {
			u := uint16(b)
			return u*256 + u
		}

		return color.RGBA64{f(c[0]), f(c[1]), f(c[2]), 65535}
	}

	log.Fatalf("invalid bgcolor: %s", optBgcolor)
	return bgcolor
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

	pick := l.Picker

	if pick.Bounds().Empty() {
		fmt.Printf("[warn] empty layer: %s\n", layerName)
		return nil
	}

	fmt.Printf("%s -> %s.png\n", layerName, filename)

	if optKeepOriginalBound {
		pick = adjustBound(originalBound, pick)
	}

	var outImage image.Image

	if optBgcolor == "" {
		outImage = pick
	} else {
		bgColor := parseBgcolor()
		bounds := pick.Bounds()
		outImage = image.NewRGBA64(bounds)
		max := float32(math.MaxUint16)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c := pick.At(x, y)
				r, g, b, a := c.RGBA()
				af := float32(a)
				ra := af / max
				newR := uint16(float32(r)*ra + float32(bgColor.R)*(1-ra))
				newG := uint16(float32(g)*ra + float32(bgColor.G)*(1-ra))
				newB := uint16(float32(b)*ra + float32(bgColor.B)*(1-ra))
				outImage.(*image.RGBA64).Set(x, y, color.RGBA64{newR, newG, newB, uint16(math.MaxUint16)})
			}
		}
	}

	out, err := os.Create(fmt.Sprintf("%s.png", filename))
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, outImage)
}

func main() {
	flag.Parse()
	if fName == "" {
		log.Fatal("filename(-f) required")
	}

	// optBgcolor has precedence to optBgcolorWhite
	if optBgcolorWhite && optBgcolor == "" {
		optBgcolor = "white"
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
	originalBound = img.Config.Rect

	for i, layer := range img.Layer {
		fn := filepath.Join(outDir, fmt.Sprintf("%03d", i))
		if err = processLayer(fn, layer.Name, &layer); err != nil {
			log.Printf("[WARN] %s: %v\n", fn, err)
		}
	}
}

// adjust image's bound.
func adjustBound(dstBound image.Rectangle, src image.Image) *image.RGBA64 {
	n := image.NewRGBA64(dstBound)
	i := n.Bounds().Intersect(src.Bounds())
	for x := i.Bounds().Min.X; x < i.Bounds().Max.X; x++ {
		for y := i.Bounds().Min.Y; y < i.Bounds().Max.Y; y++ {
			n.Set(x, y, src.At(x, y))
		}
	}
	return n
}
