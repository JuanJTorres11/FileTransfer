package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func uploadFile(srcFile, serverAddr string) {
	// connect to server
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// open file to upload
	fi, err := os.Open(srcFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fi.Close()

	// upload
	_, err = io.Copy(conn, fi)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func downloadFile(dstFile, serverAddr string) {
	// create new file to hold response
	fo, err := os.Create(dstFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fo.Close()

	// connect to server
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	_, err = io.Copy(fo, conn)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	fmt.Println("Indica la direccion:puerto a la cual desea conectarse")

	var port string
	fmt.Scanln(&port)
	fmt.Print(port)

	CONNECT := port
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(c, text+"\n")

		message, _ := bufio.NewReader(c).ReadString('\n')
		fmt.Print("->: " + message)
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
