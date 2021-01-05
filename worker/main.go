package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"time"
)

func ProcessDealer(msg []string) (rsp []string, err error){

	fmt.Println("worker recv",msg)
	return msg,nil
}

func main(){
	dealer, _ := zmq.NewSocket(zmq.DEALER)
	dealer.Connect("tcp://127.0.0.1:8082")
	for{
		dealer.SendMessage("fetch")
		//fmt.Println("fetch message")
		msg, _:= dealer.RecvMessage(0)
		if msg[0] == "empty"{
			//fmt.Println("fetch empty message sleep 1s")
			time.Sleep(10 * time.Millisecond)
			continue
		}
		rsp ,_:= ProcessDealer(msg)
		_,err := dealer.SendMessage(rsp)
		if err == nil{
			fmt.Println("rsp send ok")
		}
		
	}
}
