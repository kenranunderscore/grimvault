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
	cursor   uint32
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

func (d *Decoder) ReadUintEx(updateKey bool) uint32 {
	bytes := d.getBytes(4)
	encoded := binary.LittleEndian.Uint32(bytes)
	return d.DecodeEx(encoded, updateKey)
}

func (d *Decoder) ReadUint() uint32 {
	return d.ReadUintEx(true)
}

func (d *Decoder) ReadBool() bool {
	b := d.getBytes(1)[0]
	// FIXME: consolidate with `DecodeEx`
	n := byte(uint32(b) ^ d.key)
	d.key ^= d.keyTable[b]
	return n == 1
}

type Block struct {
	result uint32
	length uint32
	end    uint32
}

func (d *Decoder) ReadBlock() Block {
	result := d.ReadUint()
	length := d.ReadUintEx(false)
	end := d.cursor + length
	return Block{result, length, end}
}

func (d *Decoder) ReadBlockEnd(block Block) error {
	if block.end != d.cursor {
		return errors.New("unexpected cursor position when reading block end")
	}

	res := d.ReadUintEx(false)
	if res > 0 {
		return errors.New("block end > 0: " + strconv.FormatUint(uint64(res), 10))
	}
	return nil
}

func (d *Decoder) getBytes(count uint32) []byte {
	res := (*d.data)[d.cursor : d.cursor+count]
	d.cursor += count
	return res
}

func (d *Decoder) ReadString() (error, string) {
	length := d.ReadUint()
	if length == 0 {
		return nil, ""
	}

	if d.cursor+length > uint32(len(*d.data)) {
		return errors.New("too little data"), ""
	}

	// FIXME: consolidate
	// FIXME: decodeBytes instead?
	bytes := d.getBytes(length)
	for i := range length {
		b := bytes[i]
		decoded := byte(uint32(b) ^ d.key)
		d.key ^= d.keyTable[b]
		bytes[i] = decoded
	}

	return nil, string(bytes)
}

type StashTab struct {
	items         uint32
	width, height uint32
	block         Block
}

func (d *Decoder) ReadStashTab() (error, *StashTab) {
	fmt.Printf("   starting to read stash tab; cursor %d\n", d.cursor)
	block := d.ReadBlock()
	width := d.ReadUint()
	height := d.ReadUint()
	itemCount := d.ReadUint()
	fmt.Printf("   got stash tab block %d with %d items, cursor %d\n", block, itemCount, d.cursor)
	fmt.Printf("       width %d,  height %d\n", width, height)
	for range itemCount {
		err := d.ReadItem()
		if err != nil {
			return err, nil
		}
	}
	d.ReadBlockEnd(block)
	return nil, &StashTab{itemCount, width, height, block}
}

func (d *Decoder) ReadItem() error {
	err, base := d.ReadString()
	fmt.Printf("  base: %s\n", base)
	err, prefix := d.ReadString()
	fmt.Printf("  prefix: %s\n", prefix)
	err, suffix := d.ReadString()
	fmt.Printf("  suffix: %s\n", suffix)
	err, modifier := d.ReadString()
	fmt.Printf("  modifier: %s\n", modifier)
	err, transmute := d.ReadString()
	fmt.Printf("  transmute: %s\n", transmute)
	seed := d.ReadUint()
	fmt.Printf("  seed: %d\n", seed)
	err, material := d.ReadString()
	fmt.Printf("  material: %s\n", material)
	err, relicCompletionBonus := d.ReadString()
	fmt.Printf("  completion bonus: %s\n", relicCompletionBonus)
	relicSeed := d.ReadUint()
	fmt.Printf("  relic seed: %d\n", relicSeed)
	err, enchantment := d.ReadString()
	fmt.Printf("  enchantment: %s\n", enchantment)
	_ = d.ReadUint()
	enchantmentSeed := d.ReadUint()
	fmt.Printf("  enchantment seed: %d\n", enchantmentSeed)
	materialCombines := d.ReadUint()
	fmt.Printf("  material combines: %d\n", materialCombines)
	stackSize := d.ReadUint()
	fmt.Printf("  stack size: %d\n", stackSize)
	xpos := d.ReadUint()
	ypos := d.ReadUint()
	fmt.Printf("  pos: (%d, %d)\n", xpos, ypos)

	if err != nil {
		return err
	}
	return nil
}
