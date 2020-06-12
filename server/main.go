package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"example.com/cnp/server/goscp"

	"github.com/kbinani/screenshot"

	"github.com/gofiber/fiber"
)

func main() {
	// _, file, _, _ := runtime.Caller(0)

	// res, _ := ps.PasteImage(filepath.Join(filepath.Dir(file), "test.jpg"), "test", 10, 10)

	app := fiber.New()

	var pasteImage image.Image

	mode := 2 // 0 - production mode; 1 - presentation mode; 2 - debug mode

	app.Post("/image", func(c *fiber.Ctx) {
		file, err := c.FormFile("data")

		if err != nil {
			log.Println(err.Error())
			return
		}

		c.SaveFile(file, "image.png")

		fileOpened, err := file.Open()
		defer fileOpened.Close()

		if err != nil {
			log.Println(err.Error())
			return
		}

		pasteImage, _, err = image.Decode(fileOpened)

		if err != nil {
			log.Println(err.Error())
			return
		}
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

			fmt.Println(reflectedPoints)
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
