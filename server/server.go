package main
import (
	"github.com/gorilla/mux"
	"fmt"
	"net/http"
	"log"
	"net"
	"bytes"
	"io"
	"strings"
	"encoding/json"
	pb "protobuf"
	proto "proto"
	"time"
	"d7024e"
	"math/rand"
)

type ResponseData struct {
	Value        []byte   `json:"value,omitempty"`
	Key          string   `json:"key,omitempty"`
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	router := mux.NewRouter()
	router.HandleFunc("/store", Store)
	router.HandleFunc("/cat", Cat)
	router.HandleFunc("/pin", Pin)
	router.HandleFunc("/unpin", Unpin)
	addr = getIP()
	addr = addr + ":8000"
	for {
		addrs, err := net.LookupIP("baseNode")
		if err != nil {
			fmt.Println("could not find baseNode, trying again in 2 seconds")
			time.Sleep(2 * time.Second)
		} else {
			fmt.Println("baseNode found")
			target = addrs[0].String() + ":8000"
			break
		}
	}

	mgr = d7024e.NewMessageChannelManager()
	go Listen()
	log.Fatal(http.ListenAndServe(":8001", router))
}

var target string
var addr string
var mgr *d7024e.MessageChannelManager
const ttl = 10 * time.Second

func getIP() string {
	iface, _ := net.InterfaceByName("eth0")
        addrs, _ := iface.Addrs()
        for _, addr := range addrs {
                var ip net.IP
                switch v := addr.(type) {
                        case *net.IPNet:
                                ip = v.IP
                        case *net.IPAddr:
                                ip = v.IP
                        }
                return ip.String()
        }
	return "error"
}

func Store(w http.ResponseWriter, r *http.Request) {
	var Buf bytes.Buffer
	file, header, err := r.FormFile("file")
	if err != nil {
		//todo fix response
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")
	fmt.Printf("File name %s\n", name[0])
	io.Copy(&Buf, file)
	contents := Buf.String()
	Buf.Reset()
	id := d7024e.NewRandomKademliaID()
	msgchan := d7024e.NewMessageChannel(id)
	mgr.AddMessageChannel(msgchan)
	message := buildMsg([]string{"SERVER", addr, "store", "", contents, id.String()})
	sendMsg(target, message)
	select {
	case response := <-msgchan.Channel:
		key := response.GetKey()
		response_object := &ResponseData{}
		response_object.Key = key
		json.NewEncoder(w).Encode(response_object)
	case <-time.After(ttl):
		fmt.Println("request timed out")
		response_object := &ResponseData{}
		response_object.Key = "ERROR"
	}
	close(msgchan.Channel)
	mgr.RemoveMessageChannel(id)

	return
}
func Cat(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.FormValue("key")
	if len(key) != 40 {
		fmt.Println("length of key is off, aborting")
		return
	}
	id := d7024e.NewRandomKademliaID()
	msgchan := d7024e.NewMessageChannel(id)
	mgr.AddMessageChannel(msgchan)
	message := buildMsg([]string{"SERVER", addr, "find_val", key, "", id.String()})
	sendMsg(target, message)
	select {
	case response := <-msgchan.Channel:
		data := response.GetData()
		response_object := &ResponseData{}
		response_object.Value = data
		json.NewEncoder(w).Encode(response_object)
	case <-time.After(ttl):
		response_object := &ResponseData{}
		response_object.Value = []byte("ERROR")
		json.NewEncoder(w).Encode(response_object)
		fmt.Println("request timed out 2")
	}
	close(msgchan.Channel)
	mgr.RemoveMessageChannel(id)

	return
}

func Pin(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
        key := r.FormValue("key")
        if len(key) != 40 {
                fmt.Println("length of key is off, aborting")
                return
        }
        id := d7024e.NewRandomKademliaID()
        msgchan := d7024e.NewMessageChannel(id)
        mgr.AddMessageChannel(msgchan)
        message := buildMsg([]string{"SERVER", addr, "pin", key, "", id.String()})
        sendMsg(target, message)
        select {
        case response := <-msgchan.Channel:
                key := response.GetKey()
                response_object := &ResponseData{}
                response_object.Key = key
                json.NewEncoder(w).Encode(response_object)
        case <-time.After(ttl):
                response_object := &ResponseData{}
                response_object.Key = "ERROR"
                json.NewEncoder(w).Encode(response_object)
                fmt.Println("request timed out 2")
        }
        close(msgchan.Channel)
        mgr.RemoveMessageChannel(id)

        return
	

}
func Unpin(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
        key := r.FormValue("key")
        if len(key) != 40 {
                fmt.Println("length of key is off, aborting")
                return
        }
        id := d7024e.NewRandomKademliaID()
        msgchan := d7024e.NewMessageChannel(id)
        mgr.AddMessageChannel(msgchan)
        message := buildMsg([]string{"SERVER", addr, "unpin", key, "", id.String()})
        sendMsg(target, message)
        select {
        case response := <-msgchan.Channel:
                key := response.GetKey()
                response_object := &ResponseData{}
                response_object.Key = key
                json.NewEncoder(w).Encode(response_object)
        case <-time.After(ttl):
                response_object := &ResponseData{}
                response_object.Key = "ERROR"
                json.NewEncoder(w).Encode(response_object)
                fmt.Println("request timed out 2")
        }
        close(msgchan.Channel)
        mgr.RemoveMessageChannel(id)

        return

}

func buildMsg(input []string) *pb.KMessage {
	msg := &pb.KMessage{
		SndrID:		input[0],
		SndrAddress: 	input[1],
		MsgType: 	input[2],
		Key: 		input[3],
		Data: 		[]byte(input[4]),
		RpcID:		input[5],
	}
	return msg
}

func sendMsg(address string, msg *pb.KMessage) {
	data, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println("Marshalling error: ", err, " in sendMsg")
	}
	Conn, err := net.Dial("udp", address)
	if err != nil {
		fmt.Println("UDP error: ", err, " in sendMsg")
	}
	defer Conn.Close()
	_, err = Conn.Write(data)
	if err != nil {
		fmt.Println("Writing error: ", err, " in sendMsg")
	}

}

func Listen() {
	address, err1 := net.ResolveUDPAddr("udp", ":8000")//me.Address)
	Conn, err2 := net.ListenUDP("udp", address)
	if (err1 != nil) || (err2 != nil) {
		fmt.Println("Error listener: ", err1, " .... and : ", err2)
	}
	defer Conn.Close()
	channel := make(chan []byte)
	bufr := make([]byte, 4096)
	fmt.Println("Listening...")
	for {
		time.Sleep(5 * time.Millisecond)
		n, _, err := Conn.ReadFromUDP(bufr)
		go handleMsg(channel)
		channel <- bufr[:n]
		if err != nil {
			fmt.Println("Read error @Listen: ", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
func handleMsg(channel chan[]byte) {
	data := <-channel
	message := &pb.KMessage{}
	err := proto.Unmarshal(data, message)
	if err != nil {
		fmt.Println(err)
	}
	msgchan := mgr.GetMessageChannel(d7024e.NewKademliaID(message.GetRpcID()))
	msgchan.Channel <- message
	
}
