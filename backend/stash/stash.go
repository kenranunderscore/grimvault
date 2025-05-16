package stash

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/kenranunderscore/grimvault/backend/rawreader"
)

const TableLength = 256

type decoder struct {
	reader   *rawreader.T
	key      uint32
	keyTable *[TableLength]uint32
}

func (d *decoder) cursor() uint32 {
	return d.reader.Cursor
}

func decodeKey(r *rawreader.T) uint32 {
	const XorKey uint32 = 1431655765
	res := uint32(r.Byte())
	res |= uint32(r.Byte()) << 8
	res |= uint32(r.Byte()) << 0x10
	res |= uint32(r.Byte()) << 0x18
	return res ^ XorKey
}

func readKeyTable(r *rawreader.T) (uint32, [TableLength]uint32) {
	const Prime uint32 = 39916801
	var res [TableLength]uint32
	key := decodeKey(r)
	x := key
	for i := range TableLength {
		x = x>>1 | x<<31
		x *= Prime
		res[i] = x
	}
	return key, res
}

func newDecoder(file string) (*decoder, error) {
	reader, err := rawreader.FromFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot create stash decoder: %w", err)
	}

	key, keyTable := readKeyTable(reader)
	return &decoder{reader, key, &keyTable}, nil
}

func (d *decoder) decodeEx(encoded uint32, updateKey bool) uint32 {
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

func (d *decoder) decode(encoded uint32) uint32 {
	return d.decodeEx(encoded, true)
}

func (d *decoder) readUintEx(updateKey bool) uint32 {
	bytes := d.reader.Bytes(4)
	encoded := binary.LittleEndian.Uint32(bytes)
	return d.decodeEx(encoded, updateKey)
}

func (d *decoder) readUint() uint32 {
	return d.readUintEx(true)
}

func (d *decoder) readBool() bool {
	b := d.reader.Byte()
	// FIXME: consolidate with `DecodeEx`
	n := byte(uint32(b) ^ d.key)
	d.key ^= d.keyTable[b]
	return n == 1
}

type block struct {
	result uint32
	length uint32
	end    uint32
}

func (d *decoder) readBlock() block {
	result := d.readUint()
	length := d.readUintEx(false)
	end := d.cursor() + length
	return block{result, length, end}
}

func (d *decoder) readBlockEnd(block block) error {
	if block.end != d.cursor() {
		return errors.New("unexpected cursor position when reading block end")
	}

	res := d.readUintEx(false)
	if res > 0 {
		return errors.New("block end > 0: " + strconv.FormatUint(uint64(res), 10))
	}
	return nil
}

func (d *decoder) readString() (error, string) {
	length := d.readUint()
	if length == 0 {
		return nil, ""
	}

	if d.cursor()+length > uint32(len(d.reader.Data)) {
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

type Item struct {
	Base                 string
	Prefix               string
	Suffix               string
	Modifier             string
	Transmute            string
	Material             string
	RelicCompletionBonus string
	Enchantment          string
	Seed                 uint32
	RelicSeed            uint32
	EnchantmentSeed      uint32
	MaterialCombines     uint32
	StackSize            uint32
	X                    uint32
	Y                    uint32
}

func (d *decoder) readItem() (Item, error) {
	err, base := d.readString()
	err, prefix := d.readString()
	err, suffix := d.readString()
	err, modifier := d.readString()
	err, transmute := d.readString()
	seed := d.readUint()
	err, material := d.readString()
	err, relicCompletionBonus := d.readString()
	relicSeed := d.readUint()
	err, enchantment := d.readString()
	_ = d.readUint()
	enchantmentSeed := d.readUint()
	materialCombines := d.readUint()
	stackSize := d.readUint()
	xpos := d.readUint()
	ypos := d.readUint()

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

type StashTab struct {
	Items  []Item
	Width  uint32
	Height uint32
	Block  block
}

func (d *decoder) readStashTab() (StashTab, error) {
	block := d.readBlock()
	width := d.readUint()
	height := d.readUint()
	itemCount := d.readUint()
	items := make([]Item, 0, itemCount)
	for range itemCount {
		item, err := d.readItem()
		if err != nil {
			return StashTab{}, fmt.Errorf("failed to read item: %v", err)
		}
		items = append(items, item)
	}
	d.readBlockEnd(block)
	return StashTab{items, width, height, block}, nil
}

type Stash struct {
	Tabs []StashTab
}

func ReadStash(file string) (*Stash, error) {
	d, err := newDecoder(file)
	if err != nil {
		return nil, fmt.Errorf("could not open stash file '%s': %w", file, err)
	}

	if x := d.readUint(); x != 2 {
		return nil, fmt.Errorf("expected literal 2, got %d", x)
	}

	mainBlock := d.readBlock()
	if mainBlock.result != 18 {
		return nil, fmt.Errorf("expected main block to start with literal 18, got %d", mainBlock.result)
	}

	version := d.readUint()
	if zero := d.readUintEx(false); zero != 0 {
		return nil, fmt.Errorf("expected literal 0, got %d", zero)
	}

	d.readString()

	if version >= 5 {
		// is it an expansion stash file?
		d.readBool()
	}

	tabCount := d.readUint()
	stash := Stash{make([]StashTab, 0, tabCount)}
	for i := range tabCount {
		tab, err := d.readStashTab()
		if err != nil {
			return &stash, fmt.Errorf("failed to read tab %d", i)
		}

		stash.Tabs = append(stash.Tabs, tab)
	}

	err = d.readBlockEnd(mainBlock)
	if err != nil {
		return &stash, fmt.Errorf("failed to read main block end: %w", err)
	}

	return &stash, nil
}
