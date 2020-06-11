package goscp

import (
	"errors"
	"image"
	"image/color"
	"log"

	"gocv.io/x/gocv"
)

func rgbaToGray(img image.Image) *image.Gray {
	var (
		bounds = img.Bounds()
		gray   = image.NewGray(bounds)
	)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			var rgba = img.At(x, y)
			gray.Set(x, y, rgba)
		}
	}
	return gray
}

// FindPoint - find corresponding point of center of view on screen
func FindPoint(viewImage image.Image, screenImage image.Image) (int, int, error) {
	view, err := gocv.ImageGrayToMatGray(rgbaToGray(viewImage))
	defer view.Close()

	if err != nil {
		return 0, 0, err
	}

	screen, err := gocv.ImageGrayToMatGray(rgbaToGray(screenImage))
	defer screen.Close()

	if err != nil {
		return 0, 0, err
	}

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

	if len(good) < 3 {
		return 0, 0, errors.New("Not enougt points found")
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

	src := gocv.NewMatWithSize(1, 1, gocv.MatTypeCV64FC2)
	defer src.Close()
	src.SetDoubleAt(0, 0, float64((w-1)/2))
	src.SetDoubleAt(0, 1, float64((h-1)/2))

	dst := gocv.NewMat()
	defer dst.Close()

	gocv.PerspectiveTransform(src, &dst, M)

	return int(dst.GetDoubleAt(0, 0)), int(dst.GetDoubleAt(0, 1)), nil
}

// DebugFindPoint - find points and open windows, where founded points are drawn
func DebugFindPoint(viewImage image.Image, screenImage image.Image) {
	window1 := gocv.NewWindow("test1")
	defer window1.Close()

	window2 := gocv.NewWindow("test2")
	defer window2.Close()

	x, y, err := FindPoint(viewImage, screenImage)

	photoMat, _ := gocv.ImageToMatRGBA(viewImage)
	defer photoMat.Close()
	screenMat, _ := gocv.ImageToMatRGBA(screenImage)
	defer screenMat.Close()

	w := photoMat.Cols()
	h := photoMat.Rows()

	if err != nil {
		log.Fatal(err.Error())
	}

	gocv.Circle(&photoMat, image.Point{(w - 1) / 2, (h - 1) / 2}, 5, color.RGBA{0, 255, 0, 0}, 5)

	gocv.Circle(&screenMat, image.Point{x, y}, 5, color.RGBA{0, 255, 0, 0}, 5)

	window1.IMShow(photoMat)
	window2.IMShow(screenMat)

	for {

		if window1.WaitKey(1) >= 0 {
			break
		}

		if window2.WaitKey(1) >= 0 {
			break
		}

	}
}
