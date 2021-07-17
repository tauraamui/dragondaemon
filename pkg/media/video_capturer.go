package media

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"gocv.io/x/gocv"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

type VideoCapturable interface {
	SetP(*gocv.VideoCapture)
	IsOpened() bool
	Read(*gocv.Mat) bool
	Close() error
}

var openVideoCapture = func(
	rtspStream string,
	title string,
	fps int,
	dateTimeLabel bool,
	dateTimeFormat string,
	mock bool,
) (VideoCapturable, error) {
	if mock {
		println("OPENING MOCK CAPTURER")
		return &mockVideoCapture{title: title}, nil
	}

	vc, err := gocv.OpenVideoCapture(rtspStream)
	if err != nil {
		return nil, err
	}

	vc.Set(gocv.VideoCaptureFPS, float64(fps))
	return &videoCapture{
		p:                   vc,
		drawDateTimeLabel:   dateTimeLabel,
		dateTimeLabelFormat: dateTimeFormat,
	}, err
}

type videoCapture struct {
	p *gocv.VideoCapture

	drawDateTimeLabel   bool
	dateTimeLabelFormat string
}

// SetP updates the internal pointer to the video capture instance.
func (vc *videoCapture) SetP(c *gocv.VideoCapture) {
	vc.p = c
}

// IsOpened reports whether the video capture instance is open.
func (vc *videoCapture) IsOpened() bool {
	return vc.p.IsOpened()
}

// Read reads the next from the video capture to the Mat. It
// returns false if the video capture cannot read the frame.
func (vc *videoCapture) Read(m *gocv.Mat) bool {
	read := vc.p.Read(m)
	if read && vc.drawDateTimeLabel {
		gocv.PutText(
			m,
			time.Now().Format(vc.dateTimeLabelFormat),
			image.Pt(15, 50),
			gocv.FontHersheyPlain,
			3,
			color.RGBA{255, 255, 255, 255},
			int(gocv.Line4),
		)
	}
	return read
}

// Close the video capture instance
func (vc *videoCapture) Close() error {
	return vc.p.Close()
}

type mockVideoCapture struct {
	title       string
	initialised bool
	baseImage   image.Image
}

// SetP doesn't do anything it exists to satisfy VideoCapturable interface
func (mvc *mockVideoCapture) SetP(_ *gocv.VideoCapture) {}

// IsOpened always returns true.
func (mvc *mockVideoCapture) IsOpened() bool {
	return true
}

// Read reads the next from the video capture to the Mat. It
// returns false if the video capture cannot read the frame.
func (mvc *mockVideoCapture) Read(m *gocv.Mat) bool {
	if !mvc.initialised {
		var w, h int = 1400, 1200
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
		mvc.baseImage = img
		mvc.initialised = true
	}

	baseClone := cloneImage(mvc.baseImage)
	err := drawText(baseClone, 5, 50, "DD_OFFLINE_STREAM")
	if err != nil {
		log.Error("unable to draw text onto in-mem image for offline stream: %w", err) //nolint
	}
	err = drawText(baseClone, 5, 180, mvc.title)
	if err != nil {
		log.Error("unable to draw text onto in-mem image for offline stream: %w", err) //nolint
	}
	err = drawText(baseClone, 5, 310, time.Now().Format("2006-01-02 15:04:05.999999999"))
	if err != nil {
		log.Error("unable to draw text onto in-mem image for offline stream: %w", err) //nolint
	}

	mat, err := gocv.ImageToMatRGB(baseClone)
	if err != nil {
		log.Fatal("Unable to convert Go image into OpenCV mat") //nolint
	}
	defer mat.Close()

	time.Sleep(time.Millisecond * 10)
	mat.CopyTo(m)
	return mvc.initialised
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

// Close the video capture instance
func (mvc *mockVideoCapture) Close() error {
	mvc.initialised = false
	return nil
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
