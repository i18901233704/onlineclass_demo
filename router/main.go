package main

import(
	"fmt"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"

)

var (
	c chan []byte = make(chan []byte, 10000 * 10)
	publisher,frouter,brouter *zmq.Socket
)


func main(){
	publisher, _ = zmq.NewSocket(zmq.PUB)
	publisher.Bind("tcp://*:8083")

	frouter, _ = zmq.NewSocket(zmq.ROUTER)
	frouter.Bind("tcp://*:8081")

	brouter, _ = zmq.NewSocket(zmq.ROUTER)
	brouter.Bind("tcp://*:8082")

	poller := zmq.NewPoller()
	poller.Add(frouter, zmq.POLLIN)
	poller.Add(brouter, zmq.POLLIN)

	for {
		polled, err := poller.Poll(-1)
        if err != nil {
            continue //  Context has been shut down
        }   

        for _, item := range polled {
            switch item.Socket {
            case frouter:
            	msg,_ := frouter.RecvMessage(0)
                ProcessFront(msg)
            case brouter:
            	msg,_ := brouter.RecvMessage(0)
                ProcessBackend(msg)
            }   
        } 
	}
}


func ProcessFront(msg []string){
	if len(msg) <2{
		return
	}
	fmt.Println("front recv:",msg)
	//identity := msg[0]
	select{
		case c <- []byte(msg[1]):
			fmt.Println("send to chan ok")
		default :
			fmt.Println("c is  full")
	}

}

func ProcessBackend(msg []string){

	if len(msg) <2{
		return
	}

	//fmt.Println("recv backend msg:",msg)
	identity := msg[0]
	if msg[1] == "fetch"{
		select {
			case m ,_ := <-c:
				brouter.SendMessage(identity, string(m))
				fmt.Println("worker pull msg",identity,string(m))
			default:
				brouter.SendMessage(identity, "empty")
				//fmt.Println("fetch but no new msg")
		}
	}else {
		m := map[string]string{}
		json.Unmarshal([]byte(msg[1]),&m)
		roomid := m["roomid"]
		fmt.Println("publish to roomid",roomid,string(msg[1]))
		publisher.SendMessage(roomid, string(msg[1]))

	}

}
