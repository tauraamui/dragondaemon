package videobackend

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"math"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/xerror"
	"gocv.io/x/gocv"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

type mockVideoBackend struct{}

func (b *mockVideoBackend) Connect(cancel context.Context, addr string) (Connection, error) {
	return &mockVideoConnection{}, nil
}

func (b *mockVideoBackend) NewFrame() videoframe.Frame {
	return &openCVFrame{mat: gocv.NewMat()}
}

func (b *mockVideoBackend) NewFrameFromBytes(d []byte) (videoframe.Frame, error) {
	openCVBackend := openCVBackend{}
	return openCVBackend.NewFrameFromBytes(d)
}

func (b *mockVideoBackend) NewWriter() videoclip.Writer {
	return &openCVClipWriter{}
}

type mockVideoConnection struct {
	uuid                    string
	cameraTitle             string
	renderedBaseFrameCanvas bool
	baseFrameCanvas         image.Image
}

func (mvc *mockVideoConnection) UUID() string {
	if len(mvc.uuid) == 0 {
		mvc.uuid = uuid.NewString()
	}
	return mvc.uuid
}

func (mvc *mockVideoConnection) Read(frame videoframe.Frame) error {
	frameMatRef, ok := frame.DataRef().(*gocv.Mat)
	if !ok {
		return xerror.New("must pass OpenCV frame to MockVideo connection read")
	}

	if !mvc.renderedBaseFrameCanvas {
		mvc.baseFrameCanvas = renderBaseFrameCanvas()
	}

	img, err := drawTextLayerOntoBaseFrameClone(
		mvc.baseFrameCanvas, mvc.cameraTitle,
	)

	if err != nil {
		return err
	}

	mat, err := gocv.ImageToMatRGB(img)
	if err != nil {
		return xerror.Errorf("unable to convert Go image into OpenCV mat: %w", err)
	}
	defer mat.Close()

	mat.CopyTo(frameMatRef)

	return nil
}

func (mvc *mockVideoConnection) IsOpen() bool {
	return true
}

// Close the video capture instance
func (mvc *mockVideoConnection) Close() error {
	mvc.renderedBaseFrameCanvas = false
	mvc.baseFrameCanvas = nil
	return nil
}

func drawTextLayerOntoBaseFrameClone(base image.Image, title string) (image.Image, error) {
	baseClone := cloneImage(base)
	err := drawText(baseClone, 5, 50, "DD_OFFLINE_STREAM")
	if err != nil {
		return nil, xerror.Errorf("unable to draw text onto in-mem image for offline stream: %w", err)
	}

	err = drawText(baseClone, 5, 180, title)
	if err != nil {
		return nil, xerror.Errorf("unable to draw text onto in-mem image for offline stream: %w", err) //nolint
	}
	err = drawText(baseClone, 5, 310, time.Now().Format("2006-01-02 15:04:05.999999999"))
	if err != nil {
		return nil, xerror.Errorf("unable to draw text onto in-mem image for offline stream: %w", err) //nolint
	}
	return baseClone, nil
}

func renderBaseFrameCanvas() image.Image {
	var w, h int = 600, 400
	var hw, hh float64 = float64(w / 2), float64(h / 2)
	r := 200.0
	θ := 2 * math.Pi / 3
	cr := &circle{hw - r*math.Sin(0), hh - r*math.Cos(0), 300}
	cg := &circle{hw - r*math.Sin(θ), hh - r*math.Cos(θ), 300}
	cb := &circle{hw - r*math.Sin(-θ), hh - r*math.Cos(-θ), 300}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c := color.RGBA{
				cr.Brightness(float64(x), float64(y)),
				cg.Brightness(float64(x), float64(y)),
				cb.Brightness(float64(x), float64(y)),
				255,
			}
			img.Set(x, y, c)
		}
	}
	return img
}

func cloneImage(src image.Image) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}

func drawText(canvas *image.RGBA, x, y int, text string) error {
	var (
		fgColor  image.Image
		fontFace *truetype.Font
		err      error
		fontSize = 64.0
	)
	fgColor = image.White
	fontFace, err = freetype.ParseFont(goregular.TTF)
	fontDrawer := &font.Drawer{
		Dst: canvas,
		Src: fgColor,
		Face: truetype.NewFace(fontFace, &truetype.Options{
			Size:    fontSize,
			Hinting: font.HintingFull,
		}),
	}
	textBounds, _ := fontDrawer.BoundString(text)
	textHeight := textBounds.Max.Y - textBounds.Min.Y
	yPosition := fixed.I((y)-textHeight.Ceil())/2 + fixed.I(textHeight.Ceil())
	fontDrawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: yPosition,
	}
	fontDrawer.DrawString(text)
	return err
}

type circle struct {
	X, Y, R float64
}

func (c *circle) Brightness(x, y float64) uint8 {
	var dx, dy float64 = c.X - x, c.Y - y
	d := math.Sqrt(dx*dx+dy*dy) / c.R
	if d > 1 {
		return 0
	} else {
		return 255
	}
}
