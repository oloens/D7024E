package d7024e

import (
	fmt "fmt"
	"net"
	"sync"
	"time"

	pb "protobuf"
	proto "proto"
)

type Network struct {
	rt     *RoutingTable
	me     *Contact
	mtx    *sync.Mutex
	target *KademliaID
	kademlia *Kademlia
}

func NewNetwork(me *Contact, rt *RoutingTable, kademlia *Kademlia) Network {
	network := Network{}
	network.me = me
	network.rt = rt
	network.mtx = &sync.Mutex{}
	network.kademlia = kademlia
	return network

}

func (network *Network) Listen(me Contact, port int) {
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
		go handleMsg(channel, &me, network)
		channel <- bufr[:n]
		if err != nil {
			fmt.Println("Read error @Listen: ", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func (network *Network) SendPingMessage(contact *Contact) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "ping"})
	sendMsg(contact.Address, message)
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "find_node", network.me.ID.String(), contact.ID.String()})
	sendMsg(contact.Address, message)
}

func (network *Network) SendFindDataMessage(hash string, contact *Contact) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "find_val", hash})
	sendMsg(contact.Address, message)
}

func (network *Network) SendStoreMessage(value string, key *KademliaID, contact *Contact) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "store", key.String(), value})
	sendMsg(contact.Address, message)
}

//Message-delarna kan kanske flyttas till egen fil?
func handleMsg(channel chan []byte, me *Contact, network *Network) {
	data := <-channel
	message := &pb.KMessage{}
	err := proto.Unmarshal(data, message)
	if err != nil {
		fmt.Println(err)
	}
	//update RoutingTable here?
	switch message.GetMsgType() {
	case "ping":
		response := buildMsg([]string{me.ID.String(), me.Address, "pong"})
		sendMsg(message.GetSndrAddress(), response)
		fmt.Println("ping received from " + message.GetSndrAddress() + " , sending pong back")
	case "pong":
		fmt.Println("pong (ping acknowledge) received from " + message.GetSndrAddress())
	case "find_node":
		targetKey :=  message.GetKey()
		target := NewKademliaID(targetKey)
		contacts := network.rt.FindClosestContacts(target, 20)
		fmt.Println(contacts)
		me_with_dist := *me
		me_with_dist.CalcDistance(target)
		if len(contacts)<20 {
			contacts = append(contacts, me_with_dist)
		}
		//TODO Actually return the contacts
	case "find_node_response":
		//TODO
	case "find_val":
		targetKey :=  message.GetKey()
		result := network.kademlia.LookupData(targetKey)
		if result != nil {
			response := buildMsg([]string{me.ID.String(), me.Address, "find_val_response", string(result[:])})
                    sendMsg(message.GetSndrAddress(), response)

		} else {
		    target := NewKademliaID(targetKey)
                    contacts := network.rt.FindClosestContacts(target, 20)
                    fmt.Println(contacts)
                    me_with_dist := *me
                    me_with_dist.CalcDistance(target)
                    if len(contacts)<20 {
                        contacts = append(contacts, me_with_dist)
                    }
                    //TODO Actually return the contacts
		}

	case "find_val_response":
		data := message.GetData()
		if data != nil {
			fmt.Println("find_value successful!")
			fmt.Println(string(data[:]))
			fmt.Println("above is value")
		}
	case "store":
		//TODO

	default:
		fmt.Println("Error in handleMsg switch")

	}
}

func buildMsg(input []string) *pb.KMessage {
	if input[2] == "ping" || input[2] == "pong" {
		msg := &pb.KMessage{
			SndrID:      input[0],//proto.String(input[0]),
			SndrAddress: input[1],//proto.String(input[1]),
			MsgType:     input[2],//proto.String(input[2]),
		}
		return msg
	}

	if input[2] == "find_node" {
		msg := &pb.KMessage{
			SndrID:      input[0],//proto.String(input[0]),
			SndrAddress: input[1],//proto.String(input[1]),
			MsgType:     input[2],//proto.String(input[2]),
			RcvrID:      input[3],//proto.String(input[3]),
			Key:	     input[4],
		}
		return msg
	}

	if input[2] == "find_val" {
		msg := &pb.KMessage{
			SndrID:      input[0],//proto.String(input[0]),
			SndrAddress: input[1],//proto.String(input[1]),
			MsgType:     input[2],//proto.String(input[2]),
			Key:         input[3],//proto.String(input[3]),
		}
		return msg
	}

	if input[2] == "find_node_response" || input[2] == "find_val_response" {
		msg := &pb.KMessage{
			SndrID:      input[0],//proto.String(input[0]),
			SndrAddress: input[1],//proto.String(input[1]),
			MsgType:     input[2],//proto.String(input[2]),
			Data:        []byte(input[3]),
		}
		return msg
	}

	if input[2] == "store" {
		msg := &pb.KMessage{
			SndrID:      input[0],//proto.String(input[0]),
			SndrAddress: input[1],//proto.String(input[1]),
			MsgType:     input[2],//proto.String(input[2]),
			Key:         input[3],//proto.String(input[3]),
			Data:        []byte(input[4]),
		}
		return msg
	} else {
		msg := &pb.KMessage{
			SndrID:      input[0],//proto.String(input[0]),
			SndrAddress: input[1],//proto.String(input[1]),
			MsgType:      "Error, no valid message",//proto.String("Error, no valid message"),
		}
		return msg
	}
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
