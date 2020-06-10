package main

import (
	"image"
	"image/color"
	"path/filepath"
	"runtime"

	"gocv.io/x/gocv"
)

func main() {
	_, file, _, _ := runtime.Caller(0)

	// res, _ := ps.PasteImage(filepath.Join(filepath.Dir(file), "test.jpg"), "test", 10, 10)

	view := gocv.IMRead(filepath.Join(filepath.Dir(file), "test_photo.jpg"), gocv.IMReadGrayScale)
	defer view.Close()

	screen := gocv.IMRead(filepath.Join(filepath.Dir(file), "test_screen.jpg"), gocv.IMReadGrayScale)
	defer screen.Close()

	window1 := gocv.NewWindow("View")
	defer window1.Close()

	window2 := gocv.NewWindow("Screen")
	defer window2.Close()

	orb := gocv.NewORB()
	defer orb.Close()

	bf := gocv.NewBFMatcher()
	defer bf.Close()

	kpView, desView := orb.DetectAndCompute(view, gocv.NewMat())
	kpScreen, desScreen := orb.DetectAndCompute(screen, gocv.NewMat())

	matches := bf.KnnMatch(desView, desScreen, 2)

	tkpView := make([]gocv.KeyPoint, 1)
	tkpScreen := make([]gocv.KeyPoint, 1)

	good := make([]gocv.DMatch, 1)
	for _, m := range matches {
		if m[0].Distance < 0.75*m[1].Distance {
			good = append(good, m[0])
			tkpView = append(tkpView, kpView[m[0].QueryIdx])
			tkpScreen = append(tkpScreen, kpScreen[m[0].TrainIdx])
		}
	}

	pointsView := gocv.NewMatWithSize(len(good), 1, gocv.MatTypeCV64FC2)
	defer pointsView.Close()
	pointsScreen := gocv.NewMatWithSize(len(good), 1, gocv.MatTypeCV64FC2)
	defer pointsScreen.Close()

	for i, keypoint := range tkpView {
		pointsView.SetDoubleAt(i, 0, float64(keypoint.X))
		pointsView.SetDoubleAt(i, 1, float64(keypoint.Y))
	}

	for i, keypoint := range tkpScreen {
		pointsScreen.SetDoubleAt(i, 0, float64(keypoint.X))
		pointsScreen.SetDoubleAt(i, 1, float64(keypoint.Y))
	}

	mask := gocv.NewMat()
	defer mask.Close()

	h := view.Rows()
	w := view.Cols()
	M := gocv.FindHomography(pointsView, &pointsScreen, gocv.HomograpyMethodRANSAC, 5.0, &mask, 2000, 0.995)

	gocv.DrawKeyPoints(view, tkpView, &view, color.RGBA{0, 0, 255, 0}, 2)
	gocv.DrawKeyPoints(screen, tkpScreen, &screen, color.RGBA{0, 0, 255, 0}, 2)

	gocv.Circle(&view, image.Point{(w - 1) / 2, (h - 1) / 2}, 5, color.RGBA{0, 255, 0, 0}, 5)

	src := gocv.NewMatWithSize(1, 1, gocv.MatTypeCV64FC2)
	defer src.Close()
	src.SetDoubleAt(0, 0, float64((w-1)/2))
	src.SetDoubleAt(0, 1, float64((h-1)/2))

	dst := gocv.NewMat()
	defer dst.Close()

	gocv.PerspectiveTransform(src, &dst, M)

	gocv.Circle(&screen, image.Point{int(dst.GetDoubleAt(0, 0)), int(dst.GetDoubleAt(0, 1))}, 5, color.RGBA{0, 255, 0, 0}, 5)

	window1.IMShow(view)
	window2.IMShow(screen)

	for {

		if window1.WaitKey(1) >= 0 {
			break
		}

		if window2.WaitKey(1) >= 0 {
			break
		}
	}

}
