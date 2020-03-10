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
)

var logger *log.Logger

func enviarArchivo (conn net.Conn, srcFile string, nCLiente int) bool {

	// Abre el archivo a enviar
	arch, err := os.Open(srcFile)
	if err != nil {
		fmt.Println(err)
		logger.Println("Hubo un error enviando el archivo al cliente " + strconv.Itoa(nCLiente))
		return false
	}
	defer arch.Close()

	hash := crearHash(arch)

	if hash != "" {
		// Envía el archivo al cliente
		_, err = io.Copy(conn, arch)
		if err != nil {
			fmt.Println(err)
			logger.Println("Hubo un error enviando el archivo al cliente " + strconv.Itoa(nCLiente))
			return false
		}

		fmt.Fprintf(conn, hash)

	} else {
		logger.Println("Hubo un error enviando el archivo al cliente " + strconv.Itoa(nCLiente))
		return false
	}

	defer conn.Close()

	logger.Println("Se envió correctamente el archivo al cliente " + strconv.Itoa(nCLiente))

	return true
}

func crearHash (archivo *os.File) string {

	// Crea una interfaz para realizar el hash
	hash := md5.New()

	// Copia el archivo a la interfaz
	_, err := io.Copy(hash, archivo)
	if err != nil {
		return ""
	}

	// Obtiene el hash en 16 bytes
	hashInBytes := hash.Sum(nil)[:16]

	// Convierte el hash a String
	return hex.EncodeToString(hashInBytes)
}

func main() {

	logFile, err := os.Create("data/log.txt")
	logger = log.New(logFile, ">>", log.LstdFlags)

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

	fmt.Println("Ingrese qué seleccion quiere enviar:\n(1) video1.mp4\n(2) video2.mp4")
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

	for i := 0; i < numero; i++ {
		conn, err := socket.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		inCliente, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		if inCliente == "Listo" {
			go enviarArchivo(conn, archivo, i)
		}

	}
}

