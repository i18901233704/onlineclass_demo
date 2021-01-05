package worker

import (
	"fmt"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
)

func ProcessDealer(msg []string) ([]string, err error){

	fmt.Println("worker recv",msg[0])
	return msg,nil
}

func main(){
	dealer, _ := zmq.NewSocket(zmq.DEALER)
	dealer.Connect("tcp://127.0.0.1:8082")
	for{
		dealer.SendMessage("fetch")
		msg, err := dealer.RecvMessage(0)
		rsp ,err := ProcessDealer(msg)
		dealer.SendMessage(rsp)

	}
}