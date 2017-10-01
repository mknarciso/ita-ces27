package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var state, myPort string

// Verifica se tem erros
func CheckError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}

func doServerJob(ServConn *net.UDPConn, buffer []byte, myPort string) {
	utc, _ := time.LoadLocation("America/Sao_Paulo")
	n, _, err := ServConn.ReadFromUDP(buffer)
	s := strings.Split(string(buffer[0:n]), ",")
	from, sent_at, message := s[0], s[1], s[2]
	i, _ := strconv.ParseInt(sent_at, 10, 64)
	fmt.Println("[", myPort, "]", from, "=>", message, "às", time.Unix(0, i).In(utc), state)
	PrintError(err)
}

func doClientJob(CliConn *net.UDPConn, msg string) {
	buf := []byte(msg)
	_, err := CliConn.Write(buf)
	PrintError(err)
}

func doSharedJob(SharedConn *net.UDPConn, msg string) {
	// Msg inicio
	msg_i := makeMsg("[Iniciando acesso ao recurso]")
	buf := []byte(msg_i)
	_, err := SharedConn.Write(buf)
	PrintError(err)
	// Msg a enviar
	buf = []byte(msg)
	_, err = SharedConn.Write(buf)
	PrintError(err)
	// Msg encerrando
	msg_f := makeMsg("[Finalizando acesso ao recurso]")
	buf = []byte(msg_f)
	_, err = SharedConn.Write(buf)
	PrintError(err)
}

func readInput(ch chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func makeMsg(x string) string {
	m := myPort
	m += ","
	m += strconv.FormatInt(time.Now().UnixNano(), 10)
	m += ","
	m += x
	return m
}

func main() {
	myPort = os.Args[1]
	nServers := len(os.Args) - 2
	otherAddr := make([]string, nServers)
	otherServerAddr := make([]*net.UDPAddr, nServers)
	CliConn := make([]*net.UDPConn, nServers)

	/// Inicializa servidor
	// Verifica meu endereço serv
	myServerAddr, err := net.ResolveUDPAddr("udp", myPort)
	CheckError(err)
	// Escuta conexões na porta encontrada
	ServConn, err := net.ListenUDP("udp", myServerAddr)
	CheckError(err)
	defer ServConn.Close()

	/// Inicializa conexões como cliente
	myClientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)
	for i := 0; i < nServers; i++ {
		otherAddr[i] = "127.0.0.1"
		otherAddr[i] += os.Args[i+2]
		// Verifica o endereço serv do outro
		otherServerAddr[i], err = net.ResolveUDPAddr("udp", otherAddr[i])
		CheckError(err)
		CliConn[i], err = net.DialUDP("udp", myClientAddr, otherServerAddr[i])
		CheckError(err)
		defer CliConn[i].Close()
	}

	/// Inicializa o acesso de escrita ao recurso compartilhado
	sharedAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	CheckError(err)
	SharedConn, err := net.DialUDP("udp", myClientAddr, sharedAddr)
	CheckError(err)
	defer SharedConn.Close()

	ch := make(chan string)
	go readInput(ch)
	buffer := make([]byte, 1024)
	state = "released"
	for {
		//Server
		go doServerJob(ServConn, buffer, myPort)
		//Client
		select {
		case x, ok := <-ch:
			if ok {

				//for j := 0; j < nServers; j++ {
				//	mensagem := makeMsg(x)
				//	go doClientJob(CliConn[j], mensagem)
				//}
				doSharedJob(SharedConn, x)
				fmt.Printf("[Mensagem] %s | Read and Sent\n", x)
			} else {
				fmt.Println("Channel closed!")
			}
		default:
			// Do nothing
			time.Sleep(time.Second * 1)
			//fmt.Println("No value ready, moving on.")
		}

		// Wait a while
		time.Sleep(time.Second * 1)
	}

}
