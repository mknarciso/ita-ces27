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

/// In order to sort the array by sent_at
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

// Check and break if has errors
func CheckError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}

// Check and print errors
func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}

// Routine that handles the listening port, receiving other RA messages
// and processing them
func doServerJob(ServConn *net.UDPConn, buffer []byte, myPort string) {
	utc, _ := time.LoadLocation("America/Sao_Paulo")
	n, _, err := ServConn.ReadFromUDP(buffer)
	PrintError(err)
	s := strings.Split(string(buffer[0:n]), ",")
	from, sent_at, message := s[0], s[1], s[2]
	i, _ := strconv.ParseInt(sent_at, 10, 64)
	// Add 5 seconds delay so we can try the mutual exclusion
	time.Sleep(time.Second * 5)

	// Process the Mutual exclusion as the RA algorithm
	if message == "request" {
		fmt.Println("Received request from", from)
		if (state == "held") || (state == "wanted" && myTime.Before(time.Unix(0, i))) {
			msgStack = append(msgStack, msgStruct{from, time.Unix(0, i).In(utc), message})
			sort.Sort(ByTime(msgStack))
			fmt.Println("Put", from, "in stack")
		} else {
			m[from].Write([]byte(makeMsg("reply")))
			fmt.Println("Sent reply to", from)

		}
	}
	if message == "reply" {
		fmt.Println("Received reply from ", from)
		delete(waiting, from)
	}
}

func doClientJob(Conn *net.UDPConn, msg string) {
	buf := []byte(msg)
	_, err := Conn.Write(buf)
	PrintError(err)
}

func doSharedJob(SharedConn *net.UDPConn, msg string) {
	// Do whatever you have to do inside the shared zone
	_, err := SharedConn.Write([]byte(makeMsg(msg)))
	PrintError(err)
}

func readInput(ch chan string) {
	// Non-blocking async routine to listen for terminal
	// input as a request for shared resource
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func makeMsg(x string) string {
	// Put a message and important info in a single string
	// comma separeted
	time_at := time.Now()
	if x == "request" {
		myTime = time_at
	}
	return myPort + "," + strconv.FormatInt(time_at.UnixNano(), 10) + "," + x
}

func enter() {
	// Request to enter the shared zone
	state = "wanted"
	fmt.Println("State is wanted")
	multicast("request")
	// wait for all other processes to reply
	if len(waiting) > 0 {
		fmt.Println("Waiting all replies")
	}
	for len(waiting) > 0 {
	}
	// Gained access to the shared zone
	state = "held"
	fmt.Println("State is held")
}

func multicast(x string) {
	// Send RA message to all other Processes
	mensagem := makeMsg(x)
	for j := 0; j < nServers; j++ {
		go doClientJob(CliConn[j], mensagem)
		waiting[otherAddr[j]] = true
	}
}

func exit() {
	// Exit the shared zone
	state = "released"
	fmt.Println("State is released")
	for len(msgStack) > 0 {
		// Top (just get next element, don't remove it)
		msg := msgStack[0]
		// Discard top element
		msgStack = msgStack[1:]
		// Reply msg
		m[msg.from].Write([]byte(makeMsg("reply")))
		fmt.Println("Sent reply to", msg.from)
	}
}

func main() {
	// Initialize variables and get terminal inputs
	m = make(map[string]*net.UDPConn)
	myPort = os.Args[1]
	nServers = len(os.Args) - 2
	waiting = make(map[string]bool)
	// Other processes server ports
	otherAddr = make([]string, nServers)
	otherServerAddr := make([]*net.UDPAddr, nServers)
	CliConn = make([]*net.UDPConn, nServers)

	/// Start server services
	myServerAddr, err := net.ResolveUDPAddr("udp", myPort)
	CheckError(err)
	ServConn, err := net.ListenUDP("udp", myServerAddr)
	CheckError(err)
	defer ServConn.Close()

	/// Start client service connection with each of the
	/// processes
	myClientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)
	for i := 0; i < nServers; i++ {
		otherAddr[i] = os.Args[i+2]
		otherServerAddr[i], err = net.ResolveUDPAddr("udp", "127.0.0.1"+otherAddr[i])
		CheckError(err)
		CliConn[i], err = net.DialUDP("udp", myClientAddr, otherServerAddr[i])
		CheckError(err)
		m[os.Args[i+2]] = CliConn[i]
		defer CliConn[i].Close()
	}

	/// Starts the connection to the shared resource
	sharedAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	CheckError(err)
	SharedConn, err := net.DialUDP("udp", myClientAddr, sharedAddr)
	CheckError(err)
	defer SharedConn.Close()

	// Starts listening to inputs (requests for shared resource)
	ch := make(chan string)
	go readInput(ch)
	buffer := make([]byte, 1024)
	state = "released"
	for {
		// Async listen to my server port
		go doServerJob(ServConn, buffer, myPort)
		// When there is a request (from stdin) ask for the
		// shared resource, to send it
		select {
		case x, ok := <-ch:
			if ok {
				enter()
				fmt.Printf("[SharedZone] Write on server: %s \n", x)
				doSharedJob(SharedConn, x)
				fmt.Println("[SharedZone] Leaving...")
				exit()
			} else {
				fmt.Println("Channel closed!")
			}
		default:
			// Do nothing
			time.Sleep(time.Second * 1)
		}

		// Wait a while
		time.Sleep(time.Second * 1)
	}

}
