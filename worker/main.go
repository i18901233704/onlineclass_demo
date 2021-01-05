package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
)

func ProcessDealer(msg []string) (rsp []string, err error){

	fmt.Println("worker recv",msg[0])
	return msg,nil
}

func main(){
	dealer, _ := zmq.NewSocket(zmq.DEALER)
	dealer.Connect("tcp://127.0.0.1:8082")
	for{
		dealer.SendMessage("fetch")
		msg, _:= dealer.RecvMessage(0)
		rsp ,_:= ProcessDealer(msg)
		dealer.SendMessage(rsp)

	}
}
