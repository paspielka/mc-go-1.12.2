package packet

import (
	"bytes"
	"compress/zlib"
	"fmt"
	. "github.com/edouard127/mc-go-1.12.2/data"
	"github.com/edouard127/mc-go-1.12.2/maths"
	"io"
	"math"
)

// Packet define a net data package
type Packet struct {
	ID   byte
	Data []byte
}

// Pack 打包一个数据包
func (p *Packet) Pack(threshold int) (pack []byte) {
	data := []byte{p.ID}           //data
	data = append(data, p.Data...) //data

	if threshold > 0 { //是否启用了压缩
		if len(data) > threshold { //是否需要压缩
			Len := len(data)
			VarLen := PackVarInt(int32(Len))
			data = Compress(data)

			pack = append(pack, PackVarInt(int32(len(VarLen)+len(data)))...)
			pack = append(pack, VarLen...)
			pack = append(pack, data...)
		} else {
			pack = append(pack, PackVarInt(int32(len(data)+1))...)
			pack = append(pack, 0x00)
			pack = append(pack, data...)
		}
	} else {
		pack = append(pack, PackVarInt(int32(len(data)))...) //len
		pack = append(pack, data...)
	}

	return
}

// PackString 打包一个字符串
func PackString(s string) (p []byte) {
	byteString := []byte(s)
	p = append(p, PackVarInt(int32(len(byteString)))...) //len
	p = append(p, byteString...)                         //data
	return
}

// PackUint16 打包一个16位无符号整数
func PackUint16(n uint16) (p []byte) {
	p = []byte{
		byte(n >> 8),
		byte(n),
	}
	return
}

// PackUint32 打包一个32位无符号整数
func PackUint32(n uint32) (p []byte) {
	p = []byte{
		byte(n >> 24), byte(n >> 16),
		byte(n >> 8), byte(n),
	}
	return
}

// PackUint64 打包一个64位无符号整数
func PackUint64(n uint64) (p []byte) {
	p = []byte{
		byte(n >> 56), byte(n >> 48), byte(n >> 40), byte(n >> 32),
		byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n),
	}
	return
}

// PackVarInt 打包一个VarInt
func PackVarInt(n int32) (VarInt []byte) {
	num := uint32(n)
	for {
		b := num & 0x7F
		num >>= 7
		if num != 0 {
			b |= 0x80
		}
		VarInt = append(VarInt, byte(b))
		if num == 0 {
			break
		}
	}
	return
}

func PackPosition(v3 maths.Vector3) (p []byte) {
	p = append(p, PackDouble(v3.X)...)
	p = append(p, PackDouble(v3.Y)...)
	p = append(p, PackDouble(v3.Z)...)
	return p
}

func PackRotation(v2 maths.Vector2) (p []byte) {
	p = append(p, PackFloat(float32(v2.X))...)
	p = append(p, PackFloat(float32(v2.Y))...)
	return p
}

// PackFloat 打包一个32位浮点数
func PackFloat(f float32) (p []byte) {
	return PackUint32(math.Float32bits(f))
}

// PackDouble 打包一个64位浮点数
func PackDouble(d float64) (p []byte) {
	return PackUint64(math.Float64bits(d))
}

// PackBoolean 打包一个布尔值
func PackBoolean(b bool) byte {
	if b {
		return 0x01
	}
	return 0x00
}

func UnpackBoolean(b io.ByteReader) (bool, error) {
	bs, err := ReadNBytes(b, 1)
	if err != nil {
		return false, err
	}
	return bs[0] == 0x01, nil
}

// ReadNBytes read N bytes from bytes.Reader
func ReadNBytes(b io.ByteReader, n int) (bs []byte, err error) {
	bs = make([]byte, n)
	for i := 0; i < n; i++ {
		bs[i], err = b.ReadByte()
		if err != nil {
			return
		}
	}
	return
}

// UnpackString 读取一个字符串
func UnpackString(b io.ByteReader) (s string, err error) {
	l, err := UnpackVarInt(b)
	if err != nil {
		return "", err
	}

	bs, err := ReadNBytes(b, int(l))

	return string(bs), err
}

// UnpackVarInt 读取一个VarInt
func UnpackVarInt(b io.ByteReader) (int32, error) {
	var n uint
	for i := 0; i < 5; i++ { //读数据前的长度标记
		sec, err := b.ReadByte()
		if err != nil {
			return 0, err
		}

		n |= uint(sec&0x7F) << uint(7*i)

		if sec&0x80 == 0 {
			break
		}
	}
	return int32(n), nil //这里要把超过int32的负数溢出
}

func UnpackAngle(b io.ByteReader) (float32, error) {
	// 0x00 = 0 degrees, 0x80 = 90 degrees, 0x100 = 180 degrees, etc.
	angle, err := b.ReadByte()
	if err != nil {
		return 0, err
	}
	return float32(angle) * 360 / 256, nil
}

func UnpackByte(b io.ByteReader) (byte, error) {
	return b.ReadByte()
}

func UnpackInt8(b io.ByteReader) (int8, error) {
	bs, err := ReadNBytes(b, 1)
	if err != nil {
		return 0, err
	}
	return int8(bs[0]), nil
}

// UnpackInt16 读取一个16位有符号整数
func UnpackInt16(b io.ByteReader) (int16, error) {
	bs, err := ReadNBytes(b, 2)
	if err != nil {
		return 0, err
	}
	return int16(bs[0])<<8 | int16(bs[1]), nil
}

// UnpackInt32 读取一个32位有符号整数
func UnpackInt32(b io.ByteReader) (int32, error) {
	bs, err := ReadNBytes(b, 4)
	if err != nil {
		return 0, err
	}
	return int32(bs[0])<<24 | int32(bs[1])<<16 | int32(bs[2])<<8 | int32(bs[3]), nil
}

// UnpackInt64 读取一个64位有符号整数
func UnpackInt64(b io.ByteReader) (int64, error) {
	bs, err := ReadNBytes(b, 8)
	if err != nil {
		return 0, err
	}
	return int64(bs[0])<<56 | int64(bs[1])<<48 | int64(bs[2])<<40 | int64(bs[3])<<32 |
		int64(bs[4])<<24 | int64(bs[5])<<16 | int64(bs[6])<<8 | int64(bs[7]), nil
}

func UnpackPosition(b io.ByteReader) (v maths.Vector3, err error) {
	position, err := UnpackInt64(b)

	v = maths.Vector3{
		X: float64(position >> 38),
		Y: float64((position >> 26) & 0xFFF),
		Z: float64(position << 38 >> 38),
	}

	// Handling negative numbers
	if v.X >= 1<<25 {
		v.X -= 1 << 26
	}
	if v.Y >= 1<<11 {
		v.Y -= 1 << 12
	}
	if v.Z >= 1<<25 {
		v.Z -= 1 << 26
	}
	return
}

func UnpackFloat(b io.ByteReader) (float32, error) {
	n, err := UnpackInt32(b)
	return math.Float32frombits(uint32(n)), err
}

func UnpackDouble(b io.ByteReader) (float64, error) {
	n, err := UnpackInt64(b)
	return math.Float64frombits(uint64(n)), err
}

type Metadata struct {
	Index byte
	Type  byte
	Value byte
}

func UnpackMetadata(b io.ByteReader) (metadata []Metadata, err error) {
	var index byte
	for {
		index, err = b.ReadByte()
		if err != nil {
			return
		}
		if index == 0xFF {
			break
		}
		metadata = append(metadata, Metadata{
			Index: index,
			Type:  index >> 5,
			Value: index & 0x1F,
		})
	}
	return
}

func UnpackSlot(b *bytes.Reader) (Slot, error) {
	index := 0
	p, err := b.ReadByte()
	if err != nil {
		return Slot{}, err
	}
	Present := p != 0x00
	index++
	if Present {
		itemID, err := UnpackVarInt(b)
		if err != nil {
			return Slot{}, err
		}
		count, err := b.ReadByte()
		if err != nil {
			return Slot{}, err
		}
		index++

		//nbt.Unmarshal(nbt.Uncompressed)

		return Slot{
			ID:    int(itemID),
			Count: count,
		}, nil
	}
	return Slot{}, nil
}

// RecvPacket receive a packet from server
func RecvPacket(r io.ByteReader, useZlib bool) (*Packet, error) {
	var len int
	for i := 0; i < 5; i++ { //读数据前的长度标记
		b, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read len of packet fail: %v", err)
		}
		len |= (int(b&0x7F) << uint(7*i))
		if b&0x80 == 0 {
			break
		}
	}

	if len < 1 {
		return nil, fmt.Errorf("packet length too short")
	}

	data := make([]byte, len) //读包内容
	var err error
	for i := 0; i < len; i++ {
		data[i], err = r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read content of packet fail: %v", err)
		}
	}

	// Decompress data
	if useZlib {
		return UnCompress(data)
	}

	return &Packet{
		ID:   data[0],
		Data: data[1:],
	}, nil
}

func UnCompress(data []byte) (*Packet, error) {
	reader := bytes.NewReader(data)
	sizeUncompressed, err := UnpackVarInt(reader)
	if err != nil {
		return nil, err
	}

	uncompressData := make([]byte, sizeUncompressed)
	if sizeUncompressed != 0 { // != 0 means compressed, let's decompress
		r, err := zlib.NewReader(reader)

		if err != nil {
			return nil, fmt.Errorf("decompress fail: %v", err)
		}
		_, err = io.ReadFull(r, uncompressData)
		if err != nil {
			return nil, fmt.Errorf("decompress fail: %v", err)
		}
		r.Close()
	} else {
		uncompressData = data[1:]
	}
	return &Packet{
		ID:   uncompressData[0],
		Data: uncompressData[1:],
	}, nil
}

func Compress(data []byte) []byte {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(data)
	w.Close()
	return b.Bytes()
}
