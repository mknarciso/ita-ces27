package main

import "time"
import "fmt"

func main() {
	utc, _ := time.LoadLocation("America/Sao_Paulo")
	t1 := time.Now()
	t2 := time.Now()
	fmt.Println(t1.In(utc), t2)
}
