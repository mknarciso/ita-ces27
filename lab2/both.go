package main
 
import (
    "fmt"
    "net"
    "os"
    "time"
    "strconv"
)
 
// Verifica se tem erros 
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Erro: " , err)
        os.Exit(0)
    }
}

func PrintError(err error) {
    if err  != nil {
        fmt.Println("Erro: " , err)
    }
}

func doServerJob(ServConn *net.UDPConn, buffer []byte, myPort string) {
    n,addr,err := ServConn.ReadFromUDP(buffer)
    fmt.Println("Recebido",myPort,string(buffer[0:n]), " de ",addr)
    PrintError(err)
}

func doClientJob(CliConn *net.UDPConn, i int) {
    msg := strconv.Itoa(i)
    buf := []byte(msg)
    _,err := CliConn.Write(buf)
    PrintError(err)
}
 
func main() {
    myPort:=os.Args[1]
    nServers := len(os.Args) - 2
    otherAddr := make([]string, nServers)
    otherServerAddr := make([]*net.UDPAddr, nServers)
    CliConn := make([]*net.UDPConn, nServers)
    
    /// Inicializa servidor
    // Verifica meu endereço serv
    myServerAddr,err := net.ResolveUDPAddr("udp",myPort)
    CheckError(err)
     // Escuta conexões na porta encontrada
    ServConn, err := net.ListenUDP("udp", myServerAddr)
    CheckError(err)
    defer ServConn.Close()
    
    /// Inicializa conexões como cliente
    myClientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    CheckError(err)
    for i := 0; i < nServers; i++ {
        otherAddr[i]="127.0.0.1"
        otherAddr[i]+=os.Args[i+2]
        // Verifica o endereço serv do outro
        otherServerAddr[i],err = net.ResolveUDPAddr("udp",otherAddr[i])
        CheckError(err)
        
        CliConn[i], err = net.DialUDP("udp", myClientAddr, otherServerAddr[i])
        CheckError(err)
        defer CliConn[i].Close()
	}
 
    buffer := make([]byte, 1024)
    i := 0
    for {
        //Server
        go doServerJob(ServConn,buffer,myPort)
        //Client
        for j := 0; j < nServers; j++ {
            go doClientJob(CliConn[j],i)
    	}
        // Wait a while
        time.Sleep(time.Second * 1)
        i++
    }
    
}
