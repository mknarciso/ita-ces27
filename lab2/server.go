package main
 
import (
    "fmt"
    "net"
    "os"
)
 
// Verifica se tem erros 
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Erro: " , err)
        os.Exit(0)
    }
}
 
func main() {
    // Verifica meu endereço válido de udp porta 10001
    Address,err := net.ResolveUDPAddr("udp",":10001")
    CheckError(err)
 
    // Escuta conexões na porta encontrada
    Connection, err := net.ListenUDP("udp", Address)
    CheckError(err)
    defer Connection.Close()
 
    buffer := make([]byte, 1024)
 
    for {
        n,addr,err := Connection.ReadFromUDP(buffer)
        fmt.Println("Recebido ",string(buffer[0:n]), " de ",addr)
 
        if err != nil {
            fmt.Println("Erro: ",err)
        } 
    }
}