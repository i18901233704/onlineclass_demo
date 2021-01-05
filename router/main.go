package router

import(
	"fmt"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"

)

var (
	c chan []byte = make(chan []byte, 10000 * 10)
)


func main(){
	publisher, _ := zmq.NewSocket(zmq.PUB)
	publisher.Bind("tcp://*:8083")

	frouter, _ := zmq.NewSocket(zmq.ROUTER)
	frouter.Bind("tcp://*:8081")

	brouter, _ := zmq.NewSocket(zmq.ROUTER)
	brouter.Bind("tcp://*:8082")

	poller := zmq.NewPoller()
	poller.Add(frouter, zmq.POLLIN)
	poller.Add(brouter, zmq.POLLIN)

	for {
		polled, err := poller.Poll(tickless.Sub(time.Now()))
        if err != nil {
            continue //  Context has been shut down
        }   

        for _, item := range polled {
            switch item.Socket {
            case frouter:
            	msg,err := frouter.RecvMessage(0)
                ProcessFront(msg)
            case brouter:
            	msg,err := backend.RecvMessage(0)
                ProcessBackend(msg)
            }   
        } 
	}
}


func ProcessFront(msg string[]){
	if len(msg) <2{
		return
	}
	identity := msg[0]
	select{
		case c <- msg[1]:
		default :
			fmt.Println("c is  full")
	}

}

func ProcessBackend(msg string[]){

	if len(msg) <2{
		return
	}

	identity := msg[0]
	if msg[1] == "Fetch"{
		var m []byte
		select {
			case m ,_ = <-c
				brouter.SendMessage(identity, string(m))
			default:
				fmt.Println("fetch but no new msg")
		}
	}else {
		m := map[string]string{}
		json.UnMarshal(&m, []byte(msg[1]))
		roomid := m["roomid"]
		publisher.SendMessage(roomid, string(msg[1]))

	}

}