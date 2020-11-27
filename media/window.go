package media

import "gocv.io/x/gocv"

func (s *Server) OpenInWindow(conn *gocv.VideoCapture, title string) {
	window := gocv.NewWindow(title)
	defer window.Close()
	img := gocv.NewMat()
	defer img.Close()

	for s.IsRunning() {
		conn.Read(&img)
		window.IMShow(img)
		window.WaitKey(1)
	}
}
