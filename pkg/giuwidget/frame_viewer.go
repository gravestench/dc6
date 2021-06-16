package giu_widget

import (
	"fmt"
	"image"

	"github.com/AllenDang/giu"

	dc6 "github.com/gravestench/dc6/pkg"
)

func FrameViewer(id string, d *dc6.DC6, tl TextureLoader) giu.Widget {
	return &frameViewer{
		id:  id,
		dc6: d,
		tl:  tl,
	}
}

var _ giu.Widget = &frameViewer{}

type frameViewerState struct {
	direction, frame int
	scale            float64
	textures         []*giu.Texture
}

func (fvs *frameViewerState) Dispose() {

}

type frameViewer struct {
	id    string
	dc6   *dc6.DC6
	state *frameViewerState
	tl    TextureLoader
}

func (fv *frameViewer) Build() {
	s := fv.getState()
	if s == nil {
		return
	}

	fv.state = s

	absIndex := (int(fv.dc6.Directions) * s.direction) + s.frame
	frame := fv.dc6.Frames[absIndex]
	w, h := float32(float64(frame.Width)*s.scale), float32(float64(frame.Height)*s.scale)

	var texture *giu.Texture

	if s.textures != nil && absIndex < len(s.textures) {
		texture = s.textures[absIndex]
	}

	layout := giu.Layout{
		giu.Custom(func() {
			if texture == nil {
				return
			}

			giu.Image(texture)
		}),
		giu.Dummy(w, h),
	}

	layout.Build()
}

func (fv *frameViewer) getStateID() string {
	return fmt.Sprintf("widget_%s", fv.id)
}

func (fv *frameViewer) getState() *frameViewerState {
	var state *frameViewerState

	s := giu.Context.GetState(fv.getStateID())

	if s != nil {
		state = s.(*frameViewerState)
	} else {
		fv.initState()
		state = fv.getState()
	}

	return state
}

func (fv *frameViewer) initState() {
	if fv.dc6 == nil {
		return
	}

	// Prevent multiple invocation to LoadImage.
	newState := &frameViewerState{
		direction: 0,
		frame:     0,
		scale:     1.0,
		textures:  make([]*giu.Texture, 0),
	}

	fv.setState(newState)

	go func() {
		numFrames := int(fv.dc6.Directions * fv.dc6.FramesPerDirection)
		textures := make([]*giu.Texture, numFrames)

		for frameIndex := 0; frameIndex < numFrames; frameIndex++ {
			fidx := frameIndex
			frame := fv.dc6.Frames[fidx]

			img := image.NewRGBA(image.Rectangle{
				Max: frame.Bounds().Size(),
			})

			for py := 0; py < int(frame.Height); py++ {
				for px := 0; px < int(frame.Width); px++ {
					img.Set(px, py, frame.At(px, py))
				}
			}

			fv.tl.CreateTextureFromARGB(img, func(t *giu.Texture) {
				textures[fidx] = t
			})
		}

		s := fv.getState()
		s.textures = textures
		fv.setState(s)
	}()
}

func (fv *frameViewer) setState(s giu.Disposable) {
	giu.Context.SetState(fv.getStateID(), s)
}
