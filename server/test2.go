package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
)

func main() {
	data, err := ioutil.ReadFile("./ExdKOOz.png")
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("python", "-u", "./u2net_run.py")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	go func() {
		defer stdin.Close()
		if _, err := stdin.Write(data); err != nil {
			panic(err)
		}
	}()

	out, err := cmd.Output()
	fmt.Println(string(out))
	if err != nil {
		log.Fatal(err)
	}
}
