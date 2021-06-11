package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gravestench/dc6/pkg"
	widget "github.com/gravestench/dc6/pkg/giu-widget"

	"github.com/AllenDang/giu"
)

const (
	title               = "dcc viewer"
	windowFlags         = giu.MasterWindowFlagsFloating & giu.MasterWindowFlagsNotResizable
	minWidth, minHeight = 1, 1
)

func main() {
	if len(os.Args) < 2 {
		return
	}

	srcPath := os.Args[1]

	fileContents, err := ioutil.ReadFile(srcPath)
	if err != nil {
		const fmtErr = "could not read file, %w"

		fmt.Print(fmt.Errorf(fmtErr, err))

		return
	}

	dc6, err := pkg.FromBytes(fileContents)
	if err != nil {
		fmt.Print(err)
		return
	}

	f0 := dc6.Frames[0]
	w, h := int(f0.Width), int(f0.Height)

	if w < minWidth {
		w = minWidth
	}

	if h < minHeight {
		h = minHeight
	}

	windowTitle := fmt.Sprintf("%s - %s", title, path.Base(srcPath))

	window := giu.NewMasterWindow(windowTitle, w, h, windowFlags, nil)
	id := fmt.Sprintf("%s##%s", windowTitle, "dc6")

	tl := widget.NewTextureLoader()

	viewer := widget.FrameViewer(id, dc6, tl)

	window.Run(func() {
		tl.ResumeLoadingTextures()
		tl.ProcessTextureLoadRequests()
		giu.SingleWindow(windowTitle).Layout(viewer)
	})
}
