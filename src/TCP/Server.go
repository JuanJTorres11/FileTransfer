package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var count = 0

func handleConnection(c net.Conn) {
	fmt.Print(".")
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}
		fmt.Println(temp)
		counter := strconv.Itoa(count) + "\n"
		c.Write([]byte(string(counter)))
	}
	c.Close()
}

func recieveFile(server net.Listener, dstFile string) {
	// accept connection
	conn, err := server.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	// create new file
	fo, err := os.Create(dstFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fo.Close()

	// accept file from client & write to new file
	_, err = io.Copy(fo, conn)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func sendFile(server net.Listener, srcFile string) {
	// accept connection
	conn, err := server.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	// open file to send
	fi, err := os.Open(srcFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fi.Close()

	// send file to client
	_, err = io.Copy(conn, fi)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	fmt.Println("Indica el puerto en  el que escuchar solicitudes")
	var port string
	fmt.Scanln(&port)
	fmt.Print(port)
	PORT := ":" + port
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
		count++
	}
}

