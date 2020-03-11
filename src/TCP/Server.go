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
	"time"
)

var logger *log.Logger
const BUFFERSIZE = 512

func enviarArchivo (conn net.Conn, in *bufio.Reader, srcFile string, nCLiente int) bool {

	// Abre el archivo a enviar
	arch, err := os.Open(srcFile)
	if err != nil {
		fmt.Println(err)
		logger.Println("Hubo un error al abrir el archivo para enviar al cliente " + strconv.Itoa(nCLiente))
		return false
	}
	defer arch.Close()

	hash := crearHash(arch)
	fmt.Println(hash)

	if hash != "" {

		antes := time.Now()

		/**
		// Envía el archivo al cliente
		info, err := arch.Stat()
		if err != nil {
			fmt.Println(err)
			return false
		}
		tamanio := completarString(strconv.FormatInt(info.Size(), 10), 10)
		fmt.Println("Enviando el tamaño del archivo")
		conn.Write([]byte(tamanio))
		buffer := make([]byte, BUFFERSIZE)
		fmt.Println("Empieza envío de archivo")
		for {
			_, err = arch.Read(buffer)
			if err == io.EOF {
				break
			}
			conn.Write(buffer)
		}
		*/

		// Envía el archivo al cliente
		arch, _ = os.Open(srcFile)
		n , err := io.Copy(conn, arch)
		fmt.Println("copió ", n)
		if err != nil {
			fmt.Println(err)
			logger.Println("Hubo un error enviando el archivo al cliente " + strconv.Itoa(nCLiente))
			return false
		}

		msj0, _ := in.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			logger.Println("El cliente ", strconv.Itoa(nCLiente), " no recibió el archivo")
			return false
		}
		if msj0 == "OK" {
			fmt.Fprintln(conn, hash)
			despues := time.Now()
			fmt.Println("se recibió el archivo")

			seg := fmt.Sprintf("%f", despues.Sub(antes).Seconds())
			logger.Println("Se envió el archivo al cliente ", strconv.Itoa(nCLiente), "y tardó ", seg, " segundos")
		} else {
			logger.Println("El cliente ", strconv.Itoa(nCLiente), " no recibe el archivo")
			return false
		}

	} else {
		logger.Println("Hubo un error al calcular el hash y enviarlo al cliente " + strconv.Itoa(nCLiente))
		return false
	}

	defer conn.Close()

	msj, err := in.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		logger.Println("El cliente " + strconv.Itoa(nCLiente) + " no pudo verificar la integridad del archivo")
		return false
	} else if msj != "Verificado" {
		logger.Println("El cliente " + strconv.Itoa(nCLiente) + " no pudo verificar la integridad del archivo")
		return false
	}

	logger.Println("Se envió correctamente el archivo al cliente " + strconv.Itoa(nCLiente))

	return true
}

func completarString (retorno string, largo int) string {
	for {
		largoRetorno := len(retorno)
		if largoRetorno < largo {
			retorno = retorno + ":"
			continue
		}
		break
	}
	return retorno
}

func crearHash (archivo *os.File) string {

	// Crea una interfaz para realizar el hash
	hash := md5.New()

	// Copia el archivo a la interfaz
	n, err := io.Copy(hash, archivo)
	fmt.Println("copió en hash ", n)
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
	logger.Println("Inicio")

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

		inCliente := bufio.NewReader(conn)
		msj1, err := inCliente.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		if msj1 == "Listo\n" || msj1 == "Listo" {
			fmt.Println(i)
			go enviarArchivo(conn, inCliente, archivo, i)
		}

	}
}

