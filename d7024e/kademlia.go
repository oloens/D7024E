package d7024e

import (
	"crypto/sha1"
	"encoding/hex"
)

type Kademlia struct {
	files []File
}
type File struct {
	Hash string
	Value 	[]byte
}
func (kademlia *Kademlia) AddFile(file File) {
	kademlia.files = append(kademlia.files, file)
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Kademlia) LookupData(hash string) []byte {
	for _, file := range kademlia.files {
		if hash == file.Hash {
			return file.Value
		}
	}
	return nil
}
func Hash(data []byte) string {
        hashbytes := sha1.Sum(data)
        hash := hex.EncodeToString(hashbytes[0:IDLength])
	return hash
}
func (kademlia *Kademlia) Store(data []byte) {
	hash := Hash(data)
	file := File{hash, data}
	kademlia.AddFile(file)
}
