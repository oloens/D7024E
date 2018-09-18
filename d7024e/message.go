package d7024e

import (
	"sync"
)

type Message struct {
	network *Network //network
	mutex   *sync.Mutex
}

func (this *Message) handleMsg(channel chan []byte, me *Contact, network *Network) {

}
