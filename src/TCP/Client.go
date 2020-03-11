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
	"time"
)

const BUFFERSIZE = 512

func descargar (dstFile string, conn net.Conn) bool {

	/**
	bufferFileSize := make([]byte, 10)
	conn.Read(bufferFileSize)
	tamanio, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	// Crea el archivo para luego guardar en este lo que se descarga
	archivo, err := os.Create(dstFile)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer archivo.Close()

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
	*/

	// Crea el archivo para luego guardar en este lo que se descarga
	archivo, err := os.Create(dstFile)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer archivo.Close()

	// Copia el archivo recibido
	_, err = io.Copy(archivo, conn)
	if err != nil {
		fmt.Println("F")
		fmt.Println(err)
		return false
	}
	fmt.Fprintln(conn, "OK")

	// Obtiene el hash del servidor
	hash, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Println(hash)
	return verificarIntegridad(archivo, hash)
}

func verificarIntegridad (archivo *os.File, hashRecibido string) bool {

	// Crea una interfaz para realizar el hash
	hash := md5.New()

	// Copia el archivo a la interfaz
	_, err := io.Copy(hash, archivo)
	if err != nil {
		return false
	}

	// Obtiene el hash en 16 bytes
	hashInBytes := hash.Sum(nil)[:16]

	// Convierte el hash a String
	if hex.EncodeToString(hashInBytes) != hashRecibido {
		return false
	}

	return true
}

func main() {
	fmt.Println("Indica la direccion:puerto a la cual desea conectarse")

	var puerto string
	fmt.Scanln(&puerto)

	con, err := net.Dial("tcp", puerto)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer con.Close()

	logFile, err := os.Create("data/log.txt")
	logger := log.New(logFile, ">>", log.LstdFlags)
	logger.Println("Inicio")

	fmt.Fprintln(con, "Listo")
	fmt.Println("Listo")

	antes := time.Now()
	seDescargo := descargar("data/archivo.mp4", con)
	despues := time.Now()

	if  seDescargo != true {
		logger.Println("Hubo un error descargando el archivo desde el servidor " + puerto)
	} else {
		seg := fmt.Sprintf("%f", despues.Sub(antes).Seconds())
		fmt.Fprintln(con, "Verificado")
		logger.Println("Se descargó correctamente el archivo y tardó " + seg + " segundos")
	}
}
