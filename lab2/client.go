package main
 
import (
    "fmt"
    "net"
    "time"
    "strconv"
)
 
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Erro: " , err)
    }
}
 
func main() {
    Server,err := net.ResolveUDPAddr("udp","127.0.0.1:10001")
    CheckError(err)
 
    Local, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    CheckError(err)
 
    Connection, err := net.DialUDP("udp", Local, Server)
    CheckError(err)
 
    defer Connection.Close()
    i := 0
    for {
        msg := strconv.Itoa(i)
        i++
        buf := []byte(msg)
        _,err := Connection.Write(buf)
        if err != nil {
            fmt.Println(msg, err)
        }
        time.Sleep(time.Second * 1)
    }
}