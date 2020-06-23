package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"example.com/cnp/server/goscp"
	"example.com/cnp/server/ps"

	"github.com/anthonynsimon/bild/blend"
	"github.com/anthonynsimon/bild/fcolor"
	"github.com/kbinani/screenshot"

	"github.com/gofiber/fiber"
)

func main() {
	_, mainPath, _, _ := runtime.Caller(0)

	app := fiber.New()

	var pasteImage image.Image

	mode := 0 // 0 - production mode; 1 - presentation mode; 2 - debug mode

	app.Post("/image", func(c *fiber.Ctx) {
		file, err := c.FormFile("data")
		if err != nil {
			log.Println(err.Error())
			return
		}

		fileOpened, err := file.Open()
		defer fileOpened.Close()

		if err != nil {
			log.Println(err.Error())
			return
		}

		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, fileOpened)
		if err != nil {
			log.Println(err.Error())
			return
		}

		fmt.Println("Proceding image")
		masked, err := ProcImage(buf.Bytes())
		if err != nil {
			log.Println(err.Error())
			return
		}

		fmt.Println("Decode mask")
		imageMask, err := png.Decode(bytes.NewReader(masked))
		if err != nil {
			log.Println(err.Error())
			return
		}

		fileOpened.Seek(0, 0)

		fmt.Println("Trying to open image..")
		pasteImage, _, err = image.Decode(fileOpened)
		if err != nil {
			log.Println(err.Error())
			return
		}

		cutedImage := blend.Blend(pasteImage, imageMask, func(a fcolor.RGBAF64, b fcolor.RGBAF64) fcolor.RGBAF64 {
			a.A = (b.R + b.G + b.B) / 3.0
			return a
		})

		out, err := os.Create("./output.png")
		defer out.Close()
		if err != nil {
			log.Println(err.Error())
			return
		}

		err = png.Encode(out, cutedImage)
		if err != nil {
			log.Println(err.Error())
			return
		}

		c.SendFile("./output.png")
	})

	app.Post("/view", func(c *fiber.Ctx) {
		if pasteImage == nil {
			file, err := os.Open("image.png")
			defer file.Close()

			if err != nil {
				log.Println(err.Error())
				return
			}

			pasteImage, _, err = image.Decode(file)

			if err != nil {
				log.Println(err.Error())
				return
			}
		}

		file, err := c.FormFile("data")

		if err != nil {
			log.Println(err.Error())
			return
		}

		// c.SaveFile(file, "view.jpg")

		fileOpened, err := file.Open()
		defer fileOpened.Close()

		if err != nil {
			log.Println(err.Error())
			return
		}

		view, _, err := image.Decode(fileOpened)

		if err != nil {
			log.Println(err.Error())
			return
		}

		screen, err := screenshot.CaptureDisplay(0)

		if err != nil {
			log.Println(err.Error())
			return
		}

		if mode == 0 {
			points := make([]image.Point, 5)

			imageWidth := pasteImage.Bounds().Max.X
			imageHeight := pasteImage.Bounds().Max.Y

			points[0] = image.Point{(view.Bounds().Max.X - 1) / 2, (view.Bounds().Max.Y - 1) / 2} // Center
			points[1] = image.Point{points[0].X - imageWidth/2, points[0].Y - imageHeight/2}      // Left-Top
			points[2] = image.Point{points[0].X + imageWidth/2, points[0].Y - imageHeight/2}      // Right-Top
			points[3] = image.Point{points[0].X - imageWidth/2, points[0].Y + imageHeight/2}      // Left-Bottom
			points[4] = image.Point{points[0].X + imageWidth/2, points[0].Y + imageHeight/2}      // Right-Bottom

			temp := image.Image(screen)

			reflectedPoints, err := goscp.FindPoints(&view, &temp, points)

			if err != nil {
				log.Println(err.Error())
				return
			}

			newWidth := (reflectedPoints[2].X - reflectedPoints[1].X + reflectedPoints[4].X - reflectedPoints[3].X) / 2
			newHeight := (reflectedPoints[3].Y - reflectedPoints[1].Y + reflectedPoints[4].Y - reflectedPoints[2].Y) / 2

			_, err = ps.PasteImage(filepath.Join(filepath.Dir(mainPath), "image.png"), "test", reflectedPoints[1].X, reflectedPoints[1].Y, newWidth, newHeight)

			if err != nil {
				log.Println(err.Error())
				return
			}
		}

		// TODO: Add presentation mode
		if mode == 1 {

		}

		// TODO: Add debug mode
		if mode == 2 {
			temp := image.Image(screen)
			goscp.DebugFindPoints(&view, &temp)
		}

	})

	// Start server
	log.Fatal(app.Listen(80))

}
