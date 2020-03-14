package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func enviarArchivo (conn net.Conn, srcFile string) bool {

	// Abre el archivo a enviar
	arch, err := os.Open(srcFile)
	if err != nil {
		return false
	}
	defer arch.Close()

	buffer := make([]byte, BUFFERSIZE)
	for {
		_, err = arch.Read(buffer)
		if err == io.EOF {
			break
		}
		conn.Write(buffer)
	}

	defer conn.Close()

	return true
}

func main() {

	fmt.Println("Indica el puerto en  el que escuchar solicitudes")
	var port string
	fmt.Scanln(&port)
	fmt.Print(port)
	PORT := ":" + port
	socket, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer socket.Close()

	fmt.Println("Ingrese qué video quiere enviar:\n(1) video1.mp4\n(2) video2.mp4")
	var seleccion string
	var archivo string

	fmt.Scanln(&seleccion)
	if seleccion == "1" {
		archivo = "data/video1.mp4"
	} else {
		archivo = "data/video2.mp4"
	}

	fmt.Println("Ingrese el número de usuarios que espera atender")
	var numero int
	fmt.Scanln(&numero)

	i := 0
	for {
		conn, err := socket.Accept()
		i ++
		if err != nil {
			fmt.Println(err)
			return
		}
		go enviarArchivo(conn, archivo)
	}
}
