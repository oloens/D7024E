package d7024e
import (
	pb "protobuf"
	"sync"
	)

type MessageChannel struct {
	ID 	*KademliaID
	Channel chan *pb.KMessage
}

func NewMessageChannel(id *KademliaID) *MessageChannel {
	return &MessageChannel{id, make(chan *pb.KMessage)}
}

type MessageChannelManager struct {
	mtx 	*sync.Mutex
	MessageChannels []*MessageChannel
}

func NewMessageChannelManager() *MessageChannelManager {
	mgr := &MessageChannelManager{}
	mgr.mtx = &sync.Mutex{}
	return mgr
}
func (mgr *MessageChannelManager) Len() int {
	return len(mgr.MessageChannels)
}
func (mgr *MessageChannelManager) AddMessageChannel(msgchan *MessageChannel) {
	mgr.mtx.Lock()
	defer mgr.mtx.Unlock()
	mgr.MessageChannels = append(mgr.MessageChannels, msgchan)
}
func (mgr *MessageChannelManager) GetMessageChannel(id *KademliaID) *MessageChannel {
	//mgr.mtx.Lock()
	//defer mgr.mtx.Unlock()
	for _, msgchan := range mgr.MessageChannels {
		if msgchan.ID.String() == id.String() {
			return msgchan
		}
	}
	return nil
}
func (mgr *MessageChannelManager) RemoveMessageChannel(id *KademliaID) {
	mgr.mtx.Lock()
	defer mgr.mtx.Unlock()
	index := -1
	for i, msgchan := range mgr.MessageChannels {
		if msgchan.ID.String() == id.String() {
			index = i
		}
	}
	if index == -1 {
		return
	}

	mgr.MessageChannels[index] = mgr.MessageChannels[len(mgr.MessageChannels)-1]
	mgr.MessageChannels = mgr.MessageChannels[:len(mgr.MessageChannels)-1]
}

