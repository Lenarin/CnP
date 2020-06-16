package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("python", "./u2net_run.py", "./image_orig.png")

	out, err := cmd.Output()
	fmt.Println(string(out))
	if err != nil {
		log.Fatal(err)
	}
}
