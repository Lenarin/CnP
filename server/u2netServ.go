package main

import (
	"encoding/binary"
	"io"
	"net"
	"os/exec"
)

var address = "localhost:8800"
var cmd *exec.Cmd

// No connections :c
/*
func init() {
	log.Println("Start python server...")

	cmd = exec.Command("python", "./u-2-net-socketwrap/start.py")
	err := cmd.Start()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}
*/
func sendMessage(conn net.Conn, data []byte) error {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(len(data)))
	_, err := conn.Write(append(buff, data...))
	if err != nil {
		return err
	}
	return nil
}

func getMessage(conn net.Conn) ([]byte, error) {
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lenBuf)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, binary.BigEndian.Uint32(lenBuf))
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// ProcImage - send image to service and return it mask
func ProcImage(imageData []byte) ([]byte, error) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = sendMessage(conn, imageData)
	if err != nil {
		return nil, err
	}
	buf, err := getMessage(conn)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
