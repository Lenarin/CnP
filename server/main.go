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
	"strconv"

	"example.com/cnp/server/goscp"
	"example.com/cnp/server/ps"

	"github.com/anthonynsimon/bild/blend"
	"github.com/anthonynsimon/bild/fcolor"
	"github.com/kbinani/screenshot"

	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
	"github.com/gofiber/logger"
)

func main() {
	_, mainPath, _, _ := runtime.Caller(0)

	app := fiber.New()

	app.Settings.BodyLimit = 16 * 1024 * 1024

	app.Use(logger.New(logger.Config{
		// Optional
		Format: "${time} ${method} ${path} - ${ip} - ${status} - ${latency}\n",
	}))

	app.Use(cors.New())

	var pasteImage image.Image

	mode := 2 // 0 - production mode; 1 - presentation mode; 2 - debug mode

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
			if (b.R + b.G + b.B) < 2.8 {
				a.A = 0
			} else {
				a.A = (b.R + b.G + b.B) / 3.0
			}
			return a
		})

		origBounds := cutedImage.Bounds()
		res := image.Rectangle{image.Point{origBounds.Min.X - 1, origBounds.Min.Y - 1}, image.Point{origBounds.Min.X - 1, origBounds.Min.Y - 1}}
		for x := origBounds.Min.X; x < origBounds.Max.X; x++ {
			for y := origBounds.Min.Y; y < origBounds.Max.Y; y++ {
				if cutedImage.RGBAAt(x, y).A != 0 {
					if res.Min.X == origBounds.Min.X-1 || x < res.Min.X {
						res.Min.X = x
					}
					if res.Min.Y == origBounds.Min.Y-1 || y < res.Min.Y {
						res.Min.Y = y
					}
					if y > res.Max.Y {
						res.Max.Y = y
					}
					if x > res.Max.X {
						res.Max.X = x
					}
				}
			}
		}

		pasteImage = cutedImage.SubImage(res)

		out, err := os.Create("./output.png")
		defer out.Close()
		if err != nil {
			log.Println(err.Error())
			return
		}

		err = png.Encode(out, pasteImage)
		if err != nil {
			log.Println(err.Error())
			return
		}

		c.Set("Connection", "close")
		c.SendFile("./output.png")
	})

	app.Post("/view", func(c *fiber.Ctx) {
		if pasteImage == nil {
			file, err := os.Open("output.png")
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

		width, _ := strconv.Atoi(c.FormValue("width"))
		height, _ := strconv.Atoi(c.FormValue("height"))
		width /= 2
		height /= 2

		file, err := c.FormFile("data")

		if err != nil {
			log.Println(err.Error())
			return
		}

		// c.SaveFile(file, "view.jpg")

		go func() {
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

			log.Println("Start Calculating")
			if mode == 0 {
				points := make([]image.Point, 5)

				points[0] = image.Point{(view.Bounds().Max.X - 1) / 2, (view.Bounds().Max.Y - 1) / 2} // Center
				points[1] = image.Point{points[0].X - width/2, points[0].Y - height/2}                // Left-Top
				points[2] = image.Point{points[0].X + width/2, points[0].Y - height/2}                // Right-Top
				points[3] = image.Point{points[0].X - width/2, points[0].Y + height/2}                // Left-Bottom
				points[4] = image.Point{points[0].X + width/2, points[0].Y + height/2}                // Right-Bottom

				temp := image.Image(screen)

				reflectedPoints, err := goscp.FindPoints(&view, &temp, points)

				if err != nil {
					log.Println(err.Error())
					return
				}

				//newWidth := (reflectedPoints[2].X - reflectedPoints[1].X + reflectedPoints[4].X - reflectedPoints[3].X) / 2
				//newHeight := (reflectedPoints[3].Y - reflectedPoints[1].Y + reflectedPoints[4].Y - reflectedPoints[2].Y) / 2
				log.Println("End Calculating")

				log.Println("Send to Photoshop")
				_, err = ps.PasteImage(filepath.Join(filepath.Dir(mainPath), "output.png"), "test", reflectedPoints[1].X, reflectedPoints[1].Y, width, height)

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
				goscp.DebugFindPoints(&view, &temp, width, height)
			}
		}()

		return

	})

	// Start server
	log.Fatal(app.Listen(80))

}
