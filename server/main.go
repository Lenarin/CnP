package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"example.com/cnp/server/goscp"

	"github.com/kbinani/screenshot"

	"github.com/gofiber/fiber"
)

func main() {
	// _, file, _, _ := runtime.Caller(0)

	// res, _ := ps.PasteImage(filepath.Join(filepath.Dir(file), "test.jpg"), "test", 10, 10)

	app := fiber.New()

	app.Post("/image", func(c *fiber.Ctx) {
		file, err := c.FormFile("data")

		if err == nil {
			c.SaveFile(file, "image.jpg")
		}
	})

	app.Post("/view", func(c *fiber.Ctx) {
		file, err := c.FormFile("data")

		if err != nil {
			log.Println(err.Error())
			return
		}

		// c.SaveFile(file, "view.jpg")

		fileOpened, err := file.Open()

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
		/*
			x, y, err := goscp.FindPoint(view, screen)

			if err != nil {
				log.Println(err.Error())
				return
			}
		*/

		goscp.DebugFindPoint(view, screen)
	})

	// Start server
	log.Fatal(app.Listen(80))

}
