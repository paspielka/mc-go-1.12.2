package gomcbot

import (
	"bytes"
	"fmt"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	"io"
)

func unpackChunkDataPacket(p *pk.Packet, hasSkyLight bool) (c *Chunk, x, y int, err error) {
	reader := bytes.NewReader(p.Data)
	// Block coordinates
	X, err := pk.UnpackInt32(reader)
	if err != nil {
		return nil, 0, 0, err
	}
	Y, err := pk.UnpackInt32(reader)
	if err != nil {
		return nil, 0, 0, err
	}
	// fmt.Println("Chunk: (", X, ", ", Y, ")") //Debug: Show Chunk loc
	fc, err := reader.ReadByte()
	if err != nil {
		return nil, 0, 0, err
	}
	FullChunk := fc != 0x00

	//Master Mask
	PrimaryBitMask, err := pk.UnpackVarInt(reader)
	if err != nil {
		return nil, 0, 0, err
	}

	// Block Data
	Size, err := pk.UnpackVarInt(reader)
	if err != nil {
		return nil, 0, 0, err
	}
	Data := make([]byte, Size)
	_, err = io.ReadAtLeast(reader, Data, int(Size))
	if err != nil {
		return nil, 0, 0, err
	}

	// Entity Information
	// NumberofBlockEntities, len := pk.UnpackVarInt(p.Data[index:])
	// index += len

	// Parsing Block Data
	cc, err := readChunkColumn(FullChunk, PrimaryBitMask, bytes.NewReader(Data), hasSkyLight)
	if err != nil {
		panic(err)
	}
	return cc, int(X), int(Y), err
}

func readChunkColumn(isFull bool, mask int32, data *bytes.Reader, hasSkyLight bool) (*Chunk, error) {
	var c Chunk
	for sectionY := 0; sectionY < 16; sectionY++ {
		if (mask & (1 << uint(sectionY))) != 0 { // Is the given bit set in the mask?
			BitsPerBlock, err := data.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("read BitsPerBlock fail: %v", err)
			}
			//Read Palette
			var palette []uint
			if BitsPerBlock < 9 {
				length, err := pk.UnpackVarInt(data)
				if err != nil {
					return nil, fmt.Errorf("read palette (id len) fail: %v", err)
				}
				palette = make([]uint, length)

				for id := uint(0); id < uint(length); id++ {
					stateID, err := pk.UnpackVarInt(data)
					if err != nil {
						return nil, fmt.Errorf("read palette (id) fail: %v", err)
					}

					palette[id] = uint(stateID)
				}
			}

			//Section data
			DataArrayLength, err := pk.UnpackVarInt(data)
			if err != nil {
				return nil, fmt.Errorf("read DataArrayLength fail: %v", err)
			}

			DataArray := make([]int64, DataArrayLength)
			for i := 0; i < int(DataArrayLength); i++ {
				DataArray[i], err = pk.UnpackInt64(data)
				if err != nil {
					return nil, fmt.Errorf("read DataArray fail: %v", err)
				}
			}
			//Populate blocks with data
			fillSection(&c.sections[sectionY], perBits(BitsPerBlock), DataArray, palette)

			//throw BlockLight data
			_, err = pk.ReadNBytes(data, 2048)
			if err != nil {
				return nil, fmt.Errorf("read BlockLight fail: %v", err)
			}

			if hasSkyLight {
				//throw SkyLight data
				_, err = pk.ReadNBytes(data, 2048)
				if err != nil {
					return nil, fmt.Errorf("read SkyLight fail: %v", err)
				}
			}
		}
	}
	if isFull { // Read Biomes Data
		_, err := pk.ReadNBytes(data, 256*4)
		if err != nil {
			return nil, fmt.Errorf("read Biomes fail: %v", err)
		}
	}

	// fmt.Println(c)
	return &c, nil
}

const defaultBitsPerBlock = 14

func perBits(BitsPerBlock byte) uint {
	switch {
	case BitsPerBlock <= 4:
		return 4
	case BitsPerBlock < 9:
		return uint(BitsPerBlock)
	default:
		return defaultBitsPerBlock
	}
}

func fillSection(s *Section, bpb uint, DataArray []int64, palette []uint) {
	mask := uint(1<<bpb - 1)
	for n := 0; n < 16*16*16; n++ {
		offset := uint(n * int(bpb))
		data := uint(DataArray[offset/64])
		data >>= offset % 64
		if offset%64 > 64-bpb {
			l := bpb + offset%64 - 64
			data &= uint(DataArray[offset/64+1] << l)
		}
		data &= mask

		if bpb < 9 {
			s.blocks[n%16][n/(16*16)][n%(16*16)/16].id = palette[data]
		} else {
			s.blocks[n%16][n/(16*16)][n%(16*16)/16].id = data
		}
	}
}
