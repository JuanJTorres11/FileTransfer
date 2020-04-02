package UDP

import (
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

// Se declara en ambos archivos ya que correran en maquinas diferentes
const BUFFERSIZE = 4048

func verificarIntegridad(ruta, hashRecibido string, logger *log.Logger) bool {

	// Crea una interfaz para realizar el hash
	hash := md5.New()

	archivo, err := os.Open(ruta)
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
	logger.Println("Se calculó el siguiente hash para el archivo recibido ", hashCalculado)
	// Convierte el hash a String
	if hashCalculado != hashRecibido {
		logger.Println("El hash no se pudo verificar")
		return false
	}

	logger.Println("Se verificó el hash del archivo de forma correcta")
	return true
}

func descargar(dstFile string, conn *net.UDPConn, logger *log.Logger) bool {

	bufferFileSize := make([]byte, 10)
	conn.Read(bufferFileSize)
	tamanio, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	logger.Println("Se recibirá un archivo de tamaño ", tamanio)

	// Crea el archivo para luego guardar en este lo que se descarga
	archivo, err := os.Create(dstFile)
	if err != nil {
		logger.Println("Hubo un error al crear el archivo  donde se descargará: ", err)
		return false
	}

	antes := time.Now()
	var bytes int64
	i := 0
	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second)) // Crea un límite de tiempo por si se cae la conexión

		if (tamanio - bytes) < BUFFERSIZE {
			_, err = io.CopyN(archivo, conn, tamanio-bytes)
			if err != nil {
				logger.Println("Hubo un error con la conexión y no se recibió el archivo completo")
				fmt.Fprintln(conn, "No")
				return false
			}
			i++
			break
		}
		_, err := io.CopyN(archivo, conn, BUFFERSIZE)
		if err != nil {
			logger.Println("Hubo un error con la conexión y no se recibió el archivo completo")
			fmt.Fprintln(conn, "No")
			return false
		}
		i++
		bytes += BUFFERSIZE
	}
	despues := time.Now()
	archivo.Close()

	logger.Println("Se recibieron ", i, " paquetes")
	logger.Println("Se recibió un archivo (sin verificar) y tardó ", despues.Sub(antes).Seconds(), " segundos")

	_, err = fmt.Fprintln(conn, "OK")
	if err != nil {
		logger.Println("Hubo un error al enviar el mensaje de control", err)
		return false
	}
	// Obtiene el hash del servidor
	buff := make([]byte, BUFFERSIZE)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second)) // Crea un límite de tiempo por si se cae la conexión
	tam, err := conn.Read(buff)
	if err != nil {
		logger.Println("Hubo un error al leer el hash del archivo", err)
		return false
	}

	hash := string(buff[0:tam])
	logger.Println("Se recibió el siguiente hash ", hash)
	return verificarIntegridad(dstFile, strings.TrimSuffix(hash, "\n"), logger)
}

func main() {

	logFile, err := os.Create("data/logClient.txt")
	logger := log.New(logFile, ">>", log.LstdFlags)
	logger.Println("Inicio")

	fmt.Println("Indica la direccion:puerto a la cual desea conectarse")
	var puerto string
	fmt.Scanln(&puerto)

	s, err := net.ResolveUDPAddr("udp4", puerto)
	con, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		fmt.Println(err)
		return
	}
	l := con.LocalAddr()
	logger.Println("Se estableció una conexión la dirección local es ", l, " la dirección remota es ", s)

	defer con.Close()

	con.Write([]byte("Listo"))
	//fmt.Fprintln(con, "Listo")
	logger.Println("Se envió el mensaje de listo al servidor UDP")

	antes := time.Now()
	seDescargo := descargar("data/archivo.mp4", con, logger)
	despues := time.Now()

	if seDescargo != true {
		con.Write([]byte("Error"))
		logger.Println("Hubo un error descargando el archivo desde el servidor " + puerto)
	} else {
		seg := despues.Sub(antes).Seconds()
		con.Write([]byte("Verificado"))
		logger.Println("Se descargó y verificó correctamente el archivo y tardó ", seg, " segundos")
	}
}
