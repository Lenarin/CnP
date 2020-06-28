package goscp

import (
	"errors"
	"image"
	"image/color"
	"log"
	"sync"

	"gocv.io/x/gocv"

	x "gocv.io/x/gocv/contrib"
)

func rgbaToGray(img *image.Image) *image.Gray {
	var (
		bounds = (*img).Bounds()
		gray   = image.NewGray(bounds)
	)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			var rgba = (*img).At(x, y)
			gray.Set(x, y, rgba)
		}
	}
	return gray
}

// FindPoints - find corresponding point of center of view on screen
//
// Function will not corrupt your images
func FindPoints(viewImage *image.Image, screenImage *image.Image, points []image.Point) ([]image.Point, error) {
	var wg sync.WaitGroup
	wg.Add(2)

	var kpView, kpScreen []gocv.KeyPoint
	var desView, desScreen gocv.Mat

	go func() {
		sift := x.NewSIFT()
		defer sift.Close()

		view, _ := gocv.ImageGrayToMatGray(rgbaToGray(viewImage))
		defer view.Close()

		kpView, desView = sift.DetectAndCompute(view, gocv.NewMat())

		wg.Done()
	}()

	go func() {
		sift := x.NewSIFT()
		defer sift.Close()

		screen, _ := gocv.ImageGrayToMatGray(rgbaToGray(screenImage))
		defer screen.Close()

		kpScreen, desScreen = sift.DetectAndCompute(screen, gocv.NewMat())

		wg.Done()
	}()

	bf := gocv.NewBFMatcher()
	defer bf.Close()

	wg.Wait()

	matches := bf.KnnMatch(desView, desScreen, 2)

	tkpView := make([]gocv.KeyPoint, 1)
	tkpScreen := make([]gocv.KeyPoint, 1)

	good := make([]gocv.DMatch, 1)
	for _, m := range matches {
		if m[0].Distance < 0.7*m[1].Distance {
			good = append(good, m[0])
			tkpView = append(tkpView, kpView[m[0].QueryIdx])
			tkpScreen = append(tkpScreen, kpScreen[m[0].TrainIdx])
		}
	}

	log.Println(len(good))

	if len(good) < 6 {
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

	for i, point := range points {
		src.SetDoubleAt(0, 0, float64(point.X))
		src.SetDoubleAt(0, 1, float64(point.Y))

		gocv.PerspectiveTransform(src, &dst, M)

		res[i] = image.Point{int(dst.GetDoubleAt(0, 0)), int(dst.GetDoubleAt(0, 1))}
	}

	return res, nil
}

// DebugFindPoints - find points and open windows, where founded points are drawn
func DebugFindPoints(viewImage *image.Image, screenImage *image.Image, width int, height int) {
	window1 := gocv.NewWindow("test1")
	defer window1.Close()

	window2 := gocv.NewWindow("test2")
	defer window2.Close()

	points := make([]image.Point, 5)

	points[0] = image.Point{((*viewImage).Bounds().Max.X - 1) / 2, ((*viewImage).Bounds().Max.Y - 1) / 2}
	points[1] = image.Point{points[0].X - width, points[0].Y - height}
	points[2] = image.Point{points[0].X + width, points[0].Y - height}
	points[3] = image.Point{points[0].X - width, points[0].Y + height}
	points[4] = image.Point{points[0].X + width, points[0].Y + height}

	matchedPoints, err := FindPoints(viewImage, screenImage, points)

	if err != nil {
		log.Println(err.Error())
		return
	}

	photoMat, _ := gocv.ImageToMatRGBA(*viewImage)
	defer photoMat.Close()
	screenMat, _ := gocv.ImageToMatRGBA(*screenImage)
	defer screenMat.Close()

	if err != nil {
		log.Fatal(err.Error())
	}

	for _, point := range points {
		gocv.Circle(&photoMat, point, 5, color.RGBA{255, 0, 0, 0}, -1)
	}

	gocv.Line(&photoMat, points[1], points[2], color.RGBA{0, 255, 0, 0}, 2)
	gocv.Line(&photoMat, points[1], points[3], color.RGBA{0, 255, 0, 0}, 2)
	gocv.Line(&photoMat, points[3], points[4], color.RGBA{0, 255, 0, 0}, 2)
	gocv.Line(&photoMat, points[4], points[2], color.RGBA{0, 255, 0, 0}, 2)

	for _, point := range matchedPoints {
		gocv.Circle(&screenMat, point, 5, color.RGBA{255, 0, 0, 0}, -1)
	}

	gocv.Line(&screenMat, matchedPoints[1], matchedPoints[2], color.RGBA{0, 255, 0, 0}, 2)
	gocv.Line(&screenMat, matchedPoints[1], matchedPoints[3], color.RGBA{0, 255, 0, 0}, 2)
	gocv.Line(&screenMat, matchedPoints[3], matchedPoints[4], color.RGBA{0, 255, 0, 0}, 2)
	gocv.Line(&screenMat, matchedPoints[2], matchedPoints[4], color.RGBA{0, 255, 0, 0}, 2)

	newWidth := (matchedPoints[2].X - matchedPoints[1].X + matchedPoints[4].X - matchedPoints[3].X) / 2
	newHeight := (matchedPoints[3].Y - matchedPoints[1].Y + matchedPoints[4].Y - matchedPoints[2].Y) / 2
	log.Println("End Calculating")

	log.Printf("[DEBUG] Width: %v\n", newWidth)
	log.Printf("[DEBUG] Height: %v\n", newHeight)

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
