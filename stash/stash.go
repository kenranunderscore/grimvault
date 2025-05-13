package stash

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/kenranunderscore/grimvault/rawreader"
)

const TableLength = 256

type Decoder struct {
	reader   *rawreader.T
	key      uint32
	keyTable *[TableLength]uint32
}

func (d *Decoder) Cursor() uint32 {
	return d.reader.Cursor
}

func DecodeKey(r *rawreader.T) uint32 {
	const XorKey uint32 = 1431655765
	res := uint32(r.Byte())
	res |= uint32(r.Byte()) << 8
	res |= uint32(r.Byte()) << 0x10
	res |= uint32(r.Byte()) << 0x18
	return res ^ XorKey
}

func ReadKeyTable(r *rawreader.T) (uint32, [TableLength]uint32) {
	const Prime uint32 = 39916801
	var res [TableLength]uint32
	key := DecodeKey(r)
	x := key
	for i := range TableLength {
		x = x>>1 | x<<31
		x *= Prime
		res[i] = x
	}
	return key, res
}

func NewDecoder(file string) (*Decoder, error) {
	reader, err := rawreader.FromFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot create stash decoder: %w", err)
	}

	key, keyTable := ReadKeyTable(reader)
	return &Decoder{reader, key, &keyTable}, nil
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
	bytes := d.reader.Bytes(4)
	encoded := binary.LittleEndian.Uint32(bytes)
	return d.DecodeEx(encoded, updateKey)
}

func (d *Decoder) ReadUint() uint32 {
	return d.ReadUintEx(true)
}

func (d *Decoder) ReadBool() bool {
	b := d.reader.Byte()
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
	end := d.Cursor() + length
	return Block{result, length, end}
}

func (d *Decoder) ReadBlockEnd(block Block) error {
	if block.end != d.Cursor() {
		return errors.New("unexpected cursor position when reading block end")
	}

	res := d.ReadUintEx(false)
	if res > 0 {
		return errors.New("block end > 0: " + strconv.FormatUint(uint64(res), 10))
	}
	return nil
}

func (d *Decoder) ReadString() (error, string) {
	length := d.ReadUint()
	if length == 0 {
		return nil, ""
	}

	if d.Cursor()+length > uint32(len(d.reader.Data)) {
		return errors.New("too little data"), ""
	}

	// FIXME: consolidate
	// FIXME: decodeBytes instead?
	bytes := d.reader.Bytes(length)
	for i := range length {
		b := bytes[i]
		decoded := byte(uint32(b) ^ d.key)
		d.key ^= d.keyTable[b]
		bytes[i] = decoded
	}

	return nil, string(bytes)
}

type StashTab struct {
	items         []Item
	width, height uint32
	block         Block
}

type Item struct {
	base                 string
	prefix               string
	suffix               string
	modifier             string
	transmute            string
	material             string
	relicCompletionBonus string
	enchantment          string
	seed                 uint32
	relicSeed            uint32
	enchantmentSeed      uint32
	materialCombines     uint32
	stackSize            uint32
	x                    uint32
	y                    uint32
}

func (d *Decoder) ReadStashTab() (error, *StashTab) {
	block := d.ReadBlock()
	width := d.ReadUint()
	height := d.ReadUint()
	itemCount := d.ReadUint()
	items := make([]Item, 0, itemCount)
	for range itemCount {
		item, err := d.ReadItem()
		if err != nil {
			return err, nil
		}
		items = append(items, item)
	}
	d.ReadBlockEnd(block)
	return nil, &StashTab{items, width, height, block}
}

func (d *Decoder) ReadItem() (Item, error) {
	err, base := d.ReadString()
	err, prefix := d.ReadString()
	err, suffix := d.ReadString()
	err, modifier := d.ReadString()
	err, transmute := d.ReadString()
	seed := d.ReadUint()
	err, material := d.ReadString()
	err, relicCompletionBonus := d.ReadString()
	relicSeed := d.ReadUint()
	err, enchantment := d.ReadString()
	_ = d.ReadUint()
	enchantmentSeed := d.ReadUint()
	materialCombines := d.ReadUint()
	stackSize := d.ReadUint()
	xpos := d.ReadUint()
	ypos := d.ReadUint()

	if err != nil {
		return Item{}, err
	}
	return Item{
		base,
		prefix,
		suffix,
		modifier,
		transmute,
		material,
		relicCompletionBonus,
		enchantment,
		seed,
		relicSeed,
		enchantmentSeed,
		materialCombines,
		stackSize,
		xpos,
		ypos,
	}, nil
}
