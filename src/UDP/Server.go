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
	"sync"
	"time"
)

var logger *log.Logger
// Se declara en ambos archivos ya que correran en maquinas diferentes
const BUFFERSIZE = 4048

func crearHash(archivo *os.File) string {

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

func completarString(retorno string, largo int) string {
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

func enviarArchivo(conn *net.UDPConn, add net.Addr, srcFile string, nCLiente int) bool {

	defer conn.Close()

	// Abre el archivo a enviar
	arch, err := os.Open(srcFile)
	if err != nil {
		logger.Println("Hubo un error: ", err, " al abrir el archivo para enviar al cliente ", nCLiente)
		return false
	}
	defer arch.Close()

	logger.Println("Se abrió el archivo y se va a calcular su hash")

	hash := crearHash(arch)

	if hash != "" {

		logger.Println("Se calculó el hash del archivo y es: ", hash)
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
		antes := time.Now()
		conn.WriteTo([]byte(tamanio), add)
		buffer := make([]byte, BUFFERSIZE)
		logger.Println("Empieza envío de archivo")
		i := 0
		for {
			_, err := arch.Read(buffer)
			i++
			if err == io.EOF {
				break
			}
			conn.WriteTo(buffer, add)
		}

		despues := time.Now()
		seg := fmt.Sprintf("%f", despues.Sub(antes).Seconds())
		logger.Println("Se enviaron ", i, " paquetes")
		logger.Println("Se envió el archivo al cliente ", strconv.Itoa(nCLiente), "y tardó ", seg, " segundos")

		buffIn := make([]byte, BUFFERSIZE)
		n, _, err := conn.ReadFrom(buffIn)
		msj0 := string(buffIn[0:n])
		if err != nil {
			logger.Println("Se produjo un error: ", err, "El cliente ", nCLiente, " no recibió el archivo")
			return false
		}
		if msj0 == "OK" || msj0 == "OK\n" {
			conn.WriteTo([]byte(hash), add)
		} else {
			logger.Println("El cliente ", nCLiente, " no recibe el archivo correctamente")
			return false
		}
	} else {
		logger.Println("Hubo un error al calcular el hash y enviarlo al cliente ", nCLiente)
		return false
	}

	buffIn := make([]byte, BUFFERSIZE)
	n, _, err := conn.ReadFrom(buffIn)
	msj := string(buffIn[0:n])
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

func iniciarConexiones(lAdd *net.UDPAddr, archivo string, i int, wg sync.WaitGroup) {

	defer wg.Done()

	conn, err := net.ListenUDP("udp", lAdd)
	if err != nil {
		logger.Println("Hubo un error al intentar establecer una conexión: ", err)
		return
	}

	logger.Println("Se estableció una conexión")

	defer conn.Close()

	buffer := make([]byte, BUFFERSIZE)

	n, add, err := conn.ReadFrom(buffer)
	if err != nil {
		logger.Println("Hubo un error al leer el mensaje de confirmación del cliente: ", err)
		return
	}
	logger.Println("Se recibió un mensaje de ", add)

	msj1 := string(buffer[0:n])
	fmt.Println(msj1)

	if msj1 == "Listo\n" || msj1 == "Listo" {
		enviarArchivo(conn, add, archivo, i)
	}

}
func main() {

	logFile, _ := os.Create("data/logServer.txt")
	logger = log.New(logFile, ">>", log.LstdFlags)
	logger.Println("Inicio")

	fmt.Println("Indica el puerto en  el que escuchar solicitudes")
	var port string
	fmt.Scanln(&port)
	PORT := ":" + port

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

	var wg sync.WaitGroup
	wg.Add(numero)

	lAdd, err := net.ResolveUDPAddr("udp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < numero; i++ {
		go iniciarConexiones(lAdd, archivo, i, wg)
	}
	wg.Wait()
}
