package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type msgStruct struct {
	from    string
	sent_at time.Time
	msg     string
}

var state, myPort string
var CliConn []*net.UDPConn
var nServers int
var otherAddr []string
var msgStack []msgStruct
var m map[string]*net.UDPConn
var waiting map[string]bool
var myTime time.Time

// In order to sort the array by sent_at
type ByTime []msgStruct

func (s ByTime) Len() int {
	return len(s)
}
func (s ByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByTime) Less(i, j int) bool {
	return s[i].sent_at.Before(s[j].sent_at)
}

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
	if message == "request" {
		if (state == "held") || (state == "wanted" && myTime.Before(time.Unix(0, i))) {
			msgStack = append(msgStack, msgStruct{from, time.Unix(0, i).In(utc), message})
			sort.Sort(ByTime(msgStack))
		} else {
			m[from].Write([]byte(makeMsg("reply")))
		}
	}
	if message == "reply" {
		delete(waiting, from)
	}

	fmt.Println(msgStack)
	PrintError(err)
}

func doClientJob(Conn *net.UDPConn, msg string) {
	buf := []byte(msg)
	_, err := Conn.Write(buf)
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
	time_at := time.Now()
	if x == "request" {
		myTime = time_at
	}
	m += strconv.FormatInt(time_at.UnixNano(), 10)
	m += ","
	m += x
	return m
}

func enter() {
	state = "wanted"
	multicast("request")
	for len(waiting) > 0 {

	}
	state = "held"

}

func multicast(x string) {
	mensagem := makeMsg(x)
	// So we can try the mutual exclusion
	time.Sleep(time.Second * 5)
	for j := 0; j < nServers; j++ {
		go doClientJob(CliConn[j], mensagem)
		waiting[otherAddr[j]] = true
	}
}

func exit() {
	state = "released"
	for len(msgStack) > 0 {
		// Top (just get next element, don't remove it)
		msg := msgStack[0]
		// Discard top element
		msgStack = msgStack[1:]
		// Reply msg
		m[msg.from].Write([]byte(makeMsg("reply")))
	}
}

func main() {
	m = make(map[string]*net.UDPConn)
	myPort = os.Args[1]
	nServers = len(os.Args) - 2
	waiting = make(map[string]bool)
	otherAddr = make([]string, nServers)
	otherServerAddr := make([]*net.UDPAddr, nServers)
	CliConn = make([]*net.UDPConn, nServers)

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
		//otherAddr[i] =
		otherAddr[i] = os.Args[i+2]
		// Verifica o endereço serv do outro
		otherServerAddr[i], err = net.ResolveUDPAddr("udp", "127.0.0.1"+otherAddr[i])
		CheckError(err)
		CliConn[i], err = net.DialUDP("udp", myClientAddr, otherServerAddr[i])
		CheckError(err)
		m[os.Args[i+2]] = CliConn[i]
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
				enter()
				doSharedJob(SharedConn, x)
				fmt.Printf("[Mensagem] %s | Read and Sent - Out of shared \n", x)
				exit()
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
