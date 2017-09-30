package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

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
	n, addr, err := ServConn.ReadFromUDP(buffer)
	fmt.Println("Recebido", myPort, string(buffer[0:n]), " de ", addr)
	PrintError(err)
}

func doClientJob(CliConn *net.UDPConn, msg string) {
	buf := []byte(msg)
	_, err := CliConn.Write(buf)
	PrintError(err)
}

func readInput(ch chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		ch <- text
	}
}

func main() {
	myPort := os.Args[1]
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
	ch := make(chan string)
	go readInput(ch)
	buffer := make([]byte, 1024)
	i := 0
	state := "released"
	for {
		//Server
		go doServerJob(ServConn, buffer, myPort)
		//Client
		select {
		case x, ok := <-ch:
			if ok {
				for j := 0; j < nServers; j++ {
					go doClientJob(CliConn[j], x)
				}
				fmt.Printf("Value %s was read and sent\n", x)
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
		i++
	}

}
