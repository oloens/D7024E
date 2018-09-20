package main
import (
	"fmt"
	"d7024e"
	"net"
	pb "../protobuf"
	proto "github.com/golang/protobuf/proto"
	"os"
	)


//just testing deployment of go in a docker container, and subsequently
// deploying a cluster using docker swarm
func main() {
	fmt.Println(os.Args[1])

	test := d7024e.NewContact(d7024e.NewRandomKademliaID(), "127.0.0.1:8000")
	rt := d7024e.NewRoutingTable(test)
	go tempListen()
}
func tempListen() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8000")
	c := make(chan *pb.KMessage)
	if err != nil {
		fmt.Println("error on resolveudpaddr")
	}
	listen, listenErr := net.ListenUDP("udp", udpAddr)
	if listenErr != nil {
		fmt.Println("Error on listenUDP")
	}
	for {
		conn, connErr := listen.Accept()
		if connErr != nil {
			continue
		}
		go receiveUDP(conn, c)
	}


}
func receiveUDP(conn net.Conn, c chan *pb.KMessage) {

	defer conn.Close() //close connection upon function termination
	data := make([]byte, 2048)
	n, err := conn.Read(data)
	if err != nil {
		fmt.Println("Error reading data")
	}
	pbdata := new(pb.KMessage)
	err = proto.Unmarshal(data[0:n], pbdata)

	//temporary print
	c <- pbdata
	fmt.Println(pbdata.GetMsgType())
}
