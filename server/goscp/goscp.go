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

// FindPoints - find corresponding point of center of view on screen
func FindPoints(viewImage image.Image, screenImage image.Image, points []image.Point) ([]image.Point, error) {
	view, err := gocv.ImageGrayToMatGray(rgbaToGray(viewImage))
	defer view.Close()

	if err != nil {
		return nil, err
	}

	screen, err := gocv.ImageGrayToMatGray(rgbaToGray(screenImage))
	defer screen.Close()

	if err != nil {
		return nil, err
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
		return nil, errors.New("Not enougt points found")
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

	M := gocv.FindHomography(pointsView, &pointsScreen, gocv.HomograpyMethodRANSAC, 5.0, &mask, 2000, 0.995)

	src := gocv.NewMatWithSize(1, 1, gocv.MatTypeCV64FC2)
	defer src.Close()

	dst := gocv.NewMat()
	defer dst.Close()

	res := make([]image.Point, len(points))

	for _, point := range points {
		src.SetDoubleAt(0, 0, float64(point.X))
		src.SetDoubleAt(0, 1, float64(point.Y))

		gocv.PerspectiveTransform(src, &dst, M)

		res = append(res, image.Point{int(dst.GetDoubleAt(0, 0)), int(dst.GetDoubleAt(0, 1))})
	}

	return res, nil
}

// DebugFindPoints - find points and open windows, where founded points are drawn
func DebugFindPoints(viewImage image.Image, screenImage image.Image) {
	window1 := gocv.NewWindow("test1")
	defer window1.Close()

	window2 := gocv.NewWindow("test2")
	defer window2.Close()

	points := make([]image.Point, 5)

	points[0] = image.Point{(viewImage.Bounds().Max.X - 1) / 2, (viewImage.Bounds().Max.Y - 1) / 2}
	points[1] = image.Point{points[0].X - 250, points[0].Y - 250}
	points[2] = image.Point{points[0].X + 250, points[0].Y - 250}
	points[3] = image.Point{points[0].X - 250, points[0].Y + 250}
	points[4] = image.Point{points[0].X + 250, points[0].Y + 250}

	matchedPoints, err := FindPoints(viewImage, screenImage, points)

	photoMat, _ := gocv.ImageToMatRGBA(viewImage)
	defer photoMat.Close()
	screenMat, _ := gocv.ImageToMatRGBA(screenImage)
	defer screenMat.Close()

	if err != nil {
		log.Fatal(err.Error())
	}

	for _, point := range points {
		gocv.Circle(&photoMat, point, 2, color.RGBA{0, 255, 0, 0}, 2)
	}

	for _, point := range matchedPoints {
		gocv.Circle(&screenMat, point, 2, color.RGBA{0, 255, 0, 0}, 2)
	}

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
