package main

import(
	zmq "github.com/pebbe/zmq4"
	"sync"
	"fmt"
	"time"
	"os"
	"encoding/json"
)


var(

	roomMsgChan chan []byte
	InnerPubAddr string = "inproc://"
	roomManager *RoomManager
)


type Room struct {

	Roomid string
	lock sync.RWMutex 
	userlist map[string]*Client

	insubscriber *zmq.Socket
	subscriber *zmq.Socket
	dealer     *zmq.Socket
	poller   *zmq.Poller
}

type RoomManager struct{
	lock sync.RWMutex
	roomlist map[string]*Room
}

func init(){
	roomMsgChan  = make(chan []byte,10000 * 10)
	roomManager = &RoomManager{roomlist: make(map[string]*Room)}

	go roomManager.RecvFromChanLoop()
}

func (m *RoomManager) AddUerToRoom(roomid ,userid string, c *Client){
	m.lock.Lock()
	r ,ok := m.roomlist[roomid]
	m.lock.Unlock()
	if ok  && r !=nil{
		r.AddUser(userid,c)
	}
}

func (m *RoomManager) RecvFromChanLoop(){
	publisher, err := zmq.NewSocket(zmq.PUB)
	defer publisher.Close()
	InnerPubAddr = fmt.Sprintf("%v%v",InnerPubAddr,time.Now().Format("2006-01-02:15:04:05"))
	fmt.Println("pub addr",InnerPubAddr)
	err = publisher.Bind(InnerPubAddr)
	if err != nil{
		fmt.Println("publisher bind:",err)
		os.Exit(1)
	}
	
	for msg := range roomMsgChan {
		fmt.Println("RecvFromChanLoop", string(msg))
		m := map[string]string{}
		json.Unmarshal(msg,&m)
		roomid := m["roomid"]
		publisher.SendMessage(roomid,msg)
	}

	os.Exit(1)

}

func (m *RoomManager) CreateRoom (roomid string){
	m.lock.Lock()
	_, ok := m.roomlist[roomid]
	if ok {
		m.lock.Unlock()
		return
	}
	 
	room := &Room{Roomid: roomid, userlist: make(map[string]*Client)}
	m.roomlist[roomid] =room
	m.lock.Unlock()
	
	c := make(chan struct{})
	go room.SendAndRecvLoop(c)
	<- c
}

func (m *RoomManager) DeleteRoom(roomid string) {
	m.lock.Lock()
	room ,ok := m.roomlist[roomid]
	delete(m.roomlist,roomid)
	m.lock.Unlock()
	if ok && room != nil{
		room.Destroy()
	}
}

func (r *Room) AddUser(userid string, c *Client) {
	r.lock.Lock()
	u, ok := r.userlist[userid]
	r.userlist[userid] = c
	r.lock.Unlock()

	if ok && u !=nil {
		u.SelfKickoff()	
	}
}

func (r * Room) Destroy(){
	r.lock.Lock()
	ulist := []*Client{}
	for _ ,c := range r.userlist{
		ulist= append(ulist,c)
	}
	r.lock.Unlock()

	for _,u := range ulist{
		u.Destroy()
	}
}

func (r * Room) DelUser(userid string){
	r.lock.Lock()
	u,ok := r.userlist[userid]
	delete(r.userlist, userid)
	r.lock.Unlock()

	if ok && u != nil{
		u.Destroy()
	}

}

func (r *Room) ConnectInPub(){
	r.insubscriber, _ = zmq.NewSocket(zmq.SUB)
	r.insubscriber.Connect(InnerPubAddr)
	fmt.Println("sub:",InnerPubAddr, r.Roomid)
	r.insubscriber.SetSubscribe(r.Roomid)
	r.poller.Add(r.insubscriber, zmq.POLLIN)
}

func (r *Room) ConnectPub(){
	r.subscriber, _ = zmq.NewSocket(zmq.SUB)
	r.subscriber.Connect("tcp://127.0.0.1:8083")
	r.subscriber.SetSubscribe(r.Roomid)
	r.poller.Add(r.subscriber, zmq.POLLIN)
}

func (r *Room) ConnectRouter(){
	r.dealer, _ = zmq.NewSocket(zmq.DEALER)
	r.dealer.Connect("tcp://127.0.0.1:8081")
	fmt.Println("Connect tcp://127.0.0.1:8081")
	r.poller.Add(r.dealer, zmq.POLLIN)
}

func (r *Room) SendAndRecvLoop(c chan struct{}){

	fmt.Println("run SendAndRecvLoop")
	r.poller = zmq.NewPoller()
	r.ConnectPub()
	r.ConnectRouter()
	r.ConnectInPub()

	c <- struct{}{}

	for {
        sockets, _ := r.poller.Poll(-1)
        for _, socket := range sockets {
            switch s := socket.Socket; s {
            case r.insubscriber:
                msg, _ := s.RecvMessage(0)
                r.ProcessInSub(msg)
            case r.subscriber:
                msg, _ := s.RecvMessage(0)
                r.ProcessSub(msg)

            case r.dealer:
                msg, _ := s.RecvMessage(0)
                r.ProcessDealer(msg)
            
        	}
    	}
	}
}


func (r *Room) ProcessInSub(msg []string){
	if len(msg) <2 {
		return
	}

	fmt.Println("ProcessInSub recv:",msg)
	/*
	m := map[string]string{}
	json.Unmarshal([]byte(msg[1],&m))
	roomid := m["roomid"]
	userid := m["userid"]
	cmd := m["cmd"]
	if cmd == "enter"{
		
	}	
	*/
	_,err := r.dealer.SendMessage(msg[1])
	if err != nil{
		fmt.Println("send to router error",err)
		return
	}
	
	fmt.Println("send to router ok")
}

func (r *Room) ProcessSub(msg []string){
	
	if len(msg) <2 {
		return
	}

	fmt.Println("ProcessSub recv:",msg[1])
	for _, c := range r.userlist {
		if c != nil{
			c.SendToClient([]byte(msg[1]))
		}
	}

}

func (r *Room) ProcessDealer(msg []string){
	
	if len(msg) ==0 {
		return
	}

	m:= map[string]string{}
	json.Unmarshal([]byte(msg[0]),&m)

	userid ,_:= m["userid"]

	c ,ok := r.userlist[userid]
	if ok && c != nil{
		c.SendToClient([]byte(msg[0]))
	}
}
