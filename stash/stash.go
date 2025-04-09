package stash

import (
	"encoding/binary"
	"os"
)

const TableLength = 256

type Decoder struct {
	data *[]byte
	cursor uint
	key uint32
	keyTable *[TableLength]uint32
}

func DecodeKey(stash []byte) uint32 {
	const XorKey uint32 = 1431655765
	res := uint32(stash[0])
	res |= uint32(stash[1]) << 8
	res |= uint32(stash[2]) << 0x10
	res |= uint32(stash[3]) << 0x18
	return res ^ XorKey
}

func ReadKeyTable(stash []byte) (uint32, [TableLength]uint32) {
	const Prime uint32 = 39916801
	var res [TableLength]uint32
	key := DecodeKey(stash)
	x := key
	for i := range TableLength {
		x = x>>1 | x<<31
		x *= Prime
		res[i] = x
	}
	return key, res
}

func NewDecoder(path string) (*Decoder, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	key, keyTable := ReadKeyTable(bytes)
	return &Decoder{&bytes, 4, key, &keyTable}, nil
}

func (d *Decoder) Decode(encoded uint32) uint32 {
	n := encoded ^ d.key
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, encoded)
	for b := range bytes {
		d.key ^= d.keyTable[b]
	}
	return n
}

func (d *Decoder) ReadUInt() uint32 {
	bytes := (*d.data)[d.cursor : d.cursor+4]
	d.cursor += 4
	encoded := binary.LittleEndian.Uint32(bytes)
	return d.Decode(encoded)
}
