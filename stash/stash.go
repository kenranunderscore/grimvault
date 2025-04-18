package stash

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
)

const TableLength = 256

type Decoder struct {
	data     *[]byte
	cursor   uint
	key      uint32
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

func (d *Decoder) DecodeEx(encoded uint32, updateKey bool) uint32 {
	n := encoded ^ d.key
	if updateKey {
		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, encoded)
		for _, b := range bytes {
			d.key ^= d.keyTable[b]
		}
	}
	return n
}

func (d *Decoder) Decode(encoded uint32) uint32 {
	return d.DecodeEx(encoded, true)
}

func (d *Decoder) ReadUIntEx(updateKey bool) uint32 {
	bytes := (*d.data)[d.cursor : d.cursor+4]
	d.cursor += 4
	encoded := binary.LittleEndian.Uint32(bytes)
	return d.DecodeEx(encoded, updateKey)
}

func (d *Decoder) ReadUInt() uint32 {
	return d.ReadUIntEx(true)
}

func (d *Decoder) ReadBool() bool {
	b := (*d.data)[d.cursor : d.cursor+1][0]
	d.cursor += 1
	// FIXME: consolidate with `DecodeEx`
	n := byte(uint32(b) ^ d.key)
	fmt.Printf("n == %d\n", n)
	d.key ^= d.keyTable[b]
	fmt.Printf("new key == %d\n", d.key)
	return n == 1
}

type Block struct {
	result uint32
	length uint32
	end    uint
}

func (d *Decoder) ReadBlock() Block {
	result := d.ReadUInt()
	length := d.ReadUIntEx(false)
	end := d.cursor + uint(length)
	return Block{result, length, end}
}

func (d *Decoder) ReadBlockEnd(block Block) error {
	if block.end != d.cursor {
		return errors.New("unexpected cursor position when reading block end")
	}

	res := d.ReadUIntEx(false)
	if res > 0 {
		return errors.New("block end > 0: " + strconv.FormatUint(uint64(res), 10))
	}
	return nil
}

func (d *Decoder) ReadString() (error, string) {
	length := d.ReadUInt()
	if length == 0 {
		return nil, ""
	}

	if d.cursor+uint(length) > uint(len(*d.data)) {
		return errors.New("too little data"), ""
	}

	return nil, "FIXME: string parsing not fully implemented yet"
}
