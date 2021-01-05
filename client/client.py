from __future__ import print_function
import websocket
import json
def Enter(ws):
    msg ={"cmd":"enter","roomid":"51talk","userid":"123","token":"123","cli_timestamp":"16092989680    00_1"}
    d=json.dumps(msg)
    ws.send(d)
    print("Sent",d)
    result = ws.recv()
    print("Received '%s'" % result)

def Chat(ws):
    msg ={"cmd":"chat","roomid":"51talk","userid":"123","token":"123","cli_timestamp":"16092989680    00_1","data":"hello world"}
    d=json.dumps(msg)
    ws.send(d)
    print("Sent",d)
    result = ws.recv()
    print("Received '%s'" % result)

if __name__ == "__main__":
    #websocket.enableTrace(True)
    ws = websocket.create_connection("ws://172.16.72.60:6060/ws")
    Enter(ws)
    Chat(ws)
    ws.close()
