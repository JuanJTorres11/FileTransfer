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

var logger *log.Logger
const BUFFERSIZE = 1024

func enviarArchivo (conn net.Conn, in *bufio.Reader, srcFile string, nCLiente int) bool {

	// Abre el archivo a enviar
	arch, err := os.Open(srcFile)
	if err != nil {
		logger.Println("Hubo un error: ", err, " al abrir el archivo para enviar al cliente ", nCLiente)
		return false
	}
	defer arch.Close()

	hash := crearHash(arch)

	if hash != "" {

		// Abre de nuevo el archivo a enviar
		arch, _ = os.Open(srcFile)

		// Envía el archivo al cliente
		info, err := arch.Stat()
		if err != nil {
			logger.Println("Hubo un error al sacar el tamaño del archivo ", err)
			return false
		}
		// Esto se usa para llenar el buffer del cliente y que este no se quede esperando
		tamanio := completarString(strconv.FormatInt(info.Size(), 10), 10)
		logger.Println("Enviando el tamaño del archivo")
		antes := time.Now()
		conn.Write([]byte(tamanio))
		buffer := make([]byte, BUFFERSIZE)
		logger.Println("Empieza envío de archivo")
		for {
			_, err = arch.Read(buffer)
			if err == io.EOF {
				break
			}
			conn.Write(buffer)
		}

		msj0, err := in.ReadString('\n')
		if err != nil {
			logger.Println("Se produjo un error: ", err, "El cliente ", nCLiente, " no recibió el archivo")
			return false
		}
		if msj0 == "OK" || msj0=="OK\n" {
			fmt.Fprintln(conn, hash)
			despues := time.Now()
			seg := fmt.Sprintf("%f", despues.Sub(antes).Seconds())
			logger.Println("Se envió el archivo al cliente ", strconv.Itoa(nCLiente), "y tardó ", seg, " segundos")
		} else {
			logger.Println("El cliente ", nCLiente, " no recibe el archivo")
			return false
		}
	} else {
		logger.Println("Hubo un error al calcular el hash y enviarlo al cliente ", nCLiente)
		return false
	}

	defer conn.Close()

	msj, err := in.ReadString('\n')
	msj3 := strings.TrimSuffix(msj, "\n")
	if err != nil {
		logger.Println("Se produjo un error: ", err, " y el cliente ", nCLiente, " no pudo verificar la integridad del archivo")
		return false
	} else if msj3 != "Verificado" {
		logger.Println("El cliente ", nCLiente, " no pudo verificar la integridad del archivo")
		return false
	}

	logger.Println("Se envió correctamente el archivo al cliente ", nCLiente)

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

	logFile, err := os.Create("data/logServer.txt")
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
			go enviarArchivo(conn, inCliente, archivo, i)
		}

	}
}

