package main
import (
	"time"
	"fmt"
	"d7024e"
	"net"
	pb "protobuf"
	proto "proto"
	//"os"
)


//just testing deployment of go in a docker container, and subsequently
// deploying a cluster using docker swarm
func main() {
	rawip := getIP()
	ip := rawip+":8000"
	me := d7024e.NewContact(d7024e.NewRandomKademliaID(), ip)
	rt := d7024e.NewRoutingTable(me)

	net := d7024e.NewNetwork(&me, rt)
	go net.Listen(me, 8000)
	tarip := "172.17.0.2:8000"
	if ip == tarip {
		fmt.Println("not pinging myself")
		for {
			time.Sleep(1000 * time.Millisecond)
		}
	} else {
	tar := d7024e.NewContact(d7024e.NewRandomKademliaID(), tarip)
	for {
		time.Sleep(1000 * time.Millisecond)
		net.SendPingMessage(&tar)
		//fmt.Println("sent ping msg, sleeping...")
	}
	}
	/*
	//fmt.Println(os.Args[1])

//	test := d7024e.NewContact(d7024e.NewRandomKademliaID(), "127.0.0.1:8000")
//	rt := d7024e.NewRoutingTable(test)
	c := make(chan *pb.KMessage)
	go tempListen(c)
	msg := buildMessage()
	target := d7024e.NewContact(d7024e.NewRandomKademliaID(), "172.17.0.2:8000")
	go sendLoop(msg, &target)
	for {
		d := <-c
		fmt.Println(d.GetSndrAddress())
	}
}
func sendLoop(msg *pb.KMessage, target *d7024e.Contact) {
	for{
		
		time.Sleep(1000 * time.Millisecond)
		//fmt.Println("trying to send message")
		sendUDP(msg, target)
	}
*/
}
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
func buildMessage() *pb.KMessage {
	t1 := "testaddress"
	msg := &pb.KMessage{
		SndrAddress: t1,//proto.String(t1),
	}
	return msg
}
func tempListen(c chan *pb.KMessage) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":8000")
	if err != nil {
		fmt.Println("error on resolveudpaddr")
	}
	udpconn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error on listenUDP")
	}
	fmt.Println("listening")
	for {
		//time.Sleep(500 * time.Millisecond)
		 receiveUDP(udpconn, c)
	}


}
func receiveUDP(conn net.Conn, c chan *pb.KMessage) {
	//fmt.Println("here")
	//defer conn.Close() //close connection upon function termination
	data := make([]byte, 2048)
	n, err := conn.Read(data)
	//fmt.Println("here2")
	if err != nil {
		fmt.Println("Error reading data")
	}
	pbdata := new(pb.KMessage)
	//fmt.Println("here3")
	proto.Unmarshal(data[0:n], pbdata)
	//fmt.Println("msg received:", pbdata.GetSndrAddress())
	//temporary print
	c <- pbdata
	fmt.Println("msg received:", pbdata.GetSndrAddress())
}
func sendUDP(msg *pb.KMessage, target *d7024e.Contact) {
	ip := target.Address
	//fmt.Println(ip)
	conn, err := net.Dial("udp", ip)
	if err != nil {
		fmt.Println("error dialing for send")
		return
	}
	//TODO marshal msg into data
	data, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println("error marshaling msg")
		return
	}
	conn.Write(data)
	///conn.Close()

}
