# Lab2 - Implementation of Ricart-Agrawala's with GoLang
Based on this lecture`s algorithm: https://www.coursera.org/learn/cloud-computing-2/lecture/51kBv/2-3-ricart-agrawalas-algorithm

We will use UDP protocols to comunicate between processes.

The server.go is the shared resource, run it with:
```go run server.go```
It runs on port :10001 and will print the messages as they come.

The ra.go is the standalone process, you should run it with the following parameters:
1st: this process listen port, like :10002
2nd to nth: the other processes ports.
Example for 3 process, in 3 different terminals:
``` 
go run ra.go :10003 :10004 :10005
```
``` 
go run ra.go :10004 :10003 :10005
```
``` 
go run ra.go :10005 :10003 :10004
```
Now you can send a message from any of the RA terminals. We've set a delay of 
5 seconds after receiving a message and before processing it, so you can see 
mutual exclusion in action.