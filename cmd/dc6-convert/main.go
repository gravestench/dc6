package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	dc6 "github.com/gravestench/dc6/pkg"
	gpl "github.com/gravestench/gpl/pkg"
)

type options struct {
	dc6Path *string
	palPath *string
	pngPath *string
}

func main() {
	var o options

	parseOptions(&o)

	//dc6BaseName := path.Base(*o.dc6Path)
	//dc6FileName := fileNameWithoutExt(dc6BaseName)

	//palBaseName := path.Base(*o.palPath)
	//palFileName := fileNameWithoutExt(palBaseName)

	dc6Data, err := ioutil.ReadFile(*o.dc6Path)
	if err != nil {
		const fmtErr = "could not read file, %v"
		fmt.Print(fmt.Errorf(fmtErr, err))

		return
	}

	d, err := dc6.FromBytes(dc6Data)
	if err != nil {
		fmt.Println(err)
		return
	}

	palData, err := ioutil.ReadFile(*o.palPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	gplInstance, err := gpl.Decode(bytes.NewBuffer(palData))
	if err != nil {
		fmt.Println("palette is not a GIMP palette file...")
		return
	}

	outfilePath := *o.pngPath
	if d.Directions > 1 || d.FramesPerDirection > 1 {
		noExt := fileNameWithoutExt(outfilePath)
		outfilePath = noExt + "_d%v_f%v.png"
	}

	d.SetPalette(color.Palette(*gplInstance))

	for dirIdx := 0; dirIdx < int(d.Directions); dirIdx++ {
		startIdx := dirIdx * int(d.FramesPerDirection)
		stopIdx := startIdx + int(d.FramesPerDirection)
		frames := d.Frames[startIdx:stopIdx]

		for frameIdx := range frames {
			outPath := outfilePath
			if d.Directions > 1 || d.FramesPerDirection > 1 {
				outPath = fmt.Sprintf(outfilePath, dirIdx, frameIdx)
			}

			f, err := os.Create(outPath)
			if err != nil {
				log.Fatal(err)
			}

			if err := png.Encode(f, frames[frameIdx]); err != nil {
				_ = f.Close()
				log.Fatal(err)
			}

			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func parseOptions(o *options) (terminate bool) {
	o.dc6Path = flag.String("dc6", "", "input dc6 file (required)")
	o.palPath = flag.String("pal", "", "input pal file (optional)")
	o.pngPath = flag.String("png", "", "path to png file (optional)")

	flag.Parse()

	if *o.dc6Path == "" {
		flag.Usage()
		return true
	}

	return false
}

func fileNameWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
