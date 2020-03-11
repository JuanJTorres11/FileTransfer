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

const BUFFERSIZE = 1024

func descargar (dstFile string, conn net.Conn, logger *log.Logger) bool {


	bufferFileSize := make([]byte, 10)
	conn.Read(bufferFileSize)
	tamanio, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	// Crea el archivo para luego guardar en este lo que se descarga
	archivo, err := os.Create(dstFile)
	if err != nil {
		logger.Println("Hubo un error al crear el archivo  donde se descargará: ", err)
		return false
	}

	antes := time.Now()
	var bytes int64

	for {
		if (tamanio - bytes) < BUFFERSIZE {
			io.CopyN(archivo, conn, (tamanio - bytes))
			conn.Read(make([]byte, (bytes+BUFFERSIZE)-tamanio))
			break
		}
		io.CopyN(archivo, conn, BUFFERSIZE)
		bytes += BUFFERSIZE
	}
	despues := time.Now()
	archivo.Close()

	logger.Println("Se recibió un archivo (sin verificar) y tardó ", despues.Sub(antes).Seconds(), " segundos")
	_ , err = fmt.Fprintln(conn, "OK")
	if err != nil {
		logger.Println("Hubo un error al enviar el mensaje de control", err)
		return false
	}
	// Obtiene el hash del servidor
	hash, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		logger.Println("Hubo un error al leer el hash del archivo", err)
		return false
	}
	return verificarIntegridad(dstFile, strings.TrimSuffix(hash, "\n"), logger)
}

func verificarIntegridad (ruta, hashRecibido string, logger *log.Logger) bool {

	// Crea una interfaz para realizar el hash
	hash := md5.New()

	archivo, err  := os.Open(ruta)
	if err != nil {
		logger.Println("Hubo un problema al abrir el archivo para calcular su hash")
		return false
	}

	defer archivo.Close()

	// Copia el archivo a la interfaz
	_, err = io.Copy(hash, archivo)
	if err != nil {
		logger.Println("Hubo un problema al copiar el archivo para calcular su hash")
		return false
	}

	// Obtiene el hash en 16 bytes
	hashInBytes := hash.Sum(nil)[:16]
	hashCalculado := hex.EncodeToString(hashInBytes)
	// Convierte el hash a String
	if hashCalculado != hashRecibido {
		logger.Println("El hash no se pudo verificar")
		return false
	}

	logger.Println("Se verificó el hash del archivo de forma correcta")
	return true
}

func main() {

	logFile, err := os.Create("data/logClient.txt")
	logger := log.New(logFile, ">>", log.LstdFlags)
	logger.Println("Inicio")

	fmt.Println("Indica la direccion:puerto a la cual desea conectarse")
	var puerto string
	fmt.Scanln(&puerto)
	con, err := net.Dial("tcp", puerto)
	if err != nil {
		logger.Println("Hubo un error: ", err,  " al conectarse al puerto ", puerto)
		return
	}

	defer con.Close()

	fmt.Fprintln(con, "Listo")

	antes := time.Now()
	seDescargo := descargar("data/archivo.mp4", con, logger)
	despues := time.Now()

	if  seDescargo != true {
		logger.Println("Hubo un error descargando el archivo desde el servidor " + puerto)
	} else {
		seg := despues.Sub(antes).Seconds()
		fmt.Fprintln(con, "Verificado")
		logger.Println("Se descargó y verificó correctamente el archivo y tardó ", seg, " segundos")
	}
}
