package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Verifica se tem erros
func CheckError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}

func main() {
	// Verifica meu endereço válido de udp porta 10001
	Address, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)

	// Escuta conexões na porta encontrada
	Connection, err := net.ListenUDP("udp", Address)
	CheckError(err)
	defer Connection.Close()

	buffer := make([]byte, 1024)
	utc, _ := time.LoadLocation("America/Sao_Paulo")
	for {
		n, _, err := Connection.ReadFromUDP(buffer)
		//fmt.Println(string(buffer[0:n]))
		s := strings.Split(string(buffer[0:n]), ",")
		from, sent_at, message := s[0], s[1], s[2]
		i, _ := strconv.ParseInt(sent_at, 10, 64)
		fmt.Println("[", from, time.Unix(0, i).In(utc), "]:", message)

		if err != nil {
			fmt.Println("Erro: ", err)
		}
	}
}
