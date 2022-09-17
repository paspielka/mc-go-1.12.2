package _struct

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/edouard127/mc-go-1.12.2/CFB8"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

// Auth includes a account
type Auth struct {
	Name string
	UUID string
	AsTk string
}

// JoinServer connect a Minecraft server.
// Return a JSON string about server status.
// see JSON format at https://wiki.vg/Server_List_Ping#Response
func (p *Auth) JoinServer(addr string, port int) (g *Game, err error) {
	// Connection
	g = new(Game)
	g.Conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		err = fmt.Errorf("cannot connect the server %q: %v", addr, err)
		return
	}

	//init Game
	g.Settings = DefaultSettings //默认设置
	g.Receiver = bufio.NewReader(g.Conn)
	g.Sender = g.Conn
	g.World.Entities = make(map[int32]*LivingEntity)
	g.World.Chunks = make(map[ChunkLoc]*Chunk)
	g.Events = make(chan Event)
	g.Motion = make(chan func())
	g.Server = Server{Addr: addr, Port: port}

	// Handshake
	hsPacket := NewHandshakePacket(340, addr, port, 2) // Constructing handshake packets
	err = g.SendPacket(hsPacket)
	if err != nil {
		err = fmt.Errorf("send handshake packect fail: %v", err)
		return
	}

	// Login
	lsPacket := newLoginStartPacket(p.Name)
	err = g.SendPacket(lsPacket) //LoginStart
	if err != nil {
		err = fmt.Errorf("send login start packect fail: %v", err)
		return
	}
	for {
		//Receive Packet
		var pack *pk.Packet
		pack, err = g.recvPacket()
		if err != nil {
			err = fmt.Errorf("recv packet at state Login fail: %v", err)
			return
		}

		//Handle Packet
		switch pack.ID {
		case 0x00: //Disconnect
			s, _ := pk.UnpackString(bytes.NewReader(pack.Data))
			err = fmt.Errorf("connect disconnected by server because: %s", s)
			return
		case 0x01: //Encryption Request
			HandleEncryptionRequest(g, pack, p)
		case 0x02: //Login Success
			// uuid, l := pk.UnpackString(pack.Data)
			// name, _ := unpackString(pack.Data[l:])
			return //switches the connection state to PLAY.
		case 0x03: //Set Compression
			threshold, _ := pk.UnpackVarInt(bytes.NewReader(pack.Data))
			g.threshold = int(threshold)
		case 0x04: //Login Plugin Request
			fmt.Println("Waring Login Plugin Request")
		}
	}
}

type encryptionRequest struct {
	ServerID    string
	PublicKey   []byte
	VerifyToken []byte
}

func UnpackEncryptionRequest(p pk.Packet) (*encryptionRequest, error) {
	r := bytes.NewReader(p.Data)
	serverID, err := pk.UnpackString(r)
	if err != nil {
		return nil, err
	}
	publicKeyLength, err := pk.UnpackVarInt(r)
	if err != nil {
		return nil, err
	}
	publicKey, err := pk.ReadNBytes(r, int(publicKeyLength))
	if err != nil {
		return nil, err
	}
	verifyTokenLength, err := pk.UnpackVarInt(r)
	if err != nil {
		return nil, err
	}
	verifyToken, err := pk.ReadNBytes(r, int(verifyTokenLength))
	if err != nil {
		return nil, err
	}

	er := encryptionRequest{
		ServerID:    serverID,
		PublicKey:   publicKey,
		VerifyToken: verifyToken,
	}
	return &er, nil
}

// authDigest computes a special SHA-1 digest required for Minecraft web
// authentication on Premium servers (online-mode=true).
// Source: http://wiki.vg/Protocol_Encryption#Server
//
// Also many, many thanks to SirCmpwn and his wonderful gist (C#):
// https://gist.github.com/SirCmpwn/404223052379e82f91e6
func authDigest(serverID string, sharedSecret, publicKey []byte) string {
	h := sha1.New()
	h.Write([]byte(serverID))
	h.Write(sharedSecret)
	h.Write(publicKey)
	hash := h.Sum(nil)

	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosComplement(hash)
	}

	// Trim away zeroes
	res := strings.TrimLeft(fmt.Sprintf("%x", hash), "0")
	if negative {
		res = "-" + res
	}

	return res
}

// little endian
func twosComplement(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = byte(^p[i])
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}

type profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type request struct {
	AccessToken     string  `json:"accessToken"`
	SelectedProfile profile `json:"selectedProfile"`
	ServerID        string  `json:"serverId"`
}

func LoginAuth(AsTk, name, UUID string, shareSecret []byte, er encryptionRequest) error {
	digest := authDigest(er.ServerID, shareSecret, er.PublicKey)

	client := http.Client{}
	requestPacket, err := json.Marshal(
		request{
			AccessToken: AsTk,
			SelectedProfile: profile{
				ID:   UUID,
				Name: name,
			},
			ServerID: digest,
		},
	)
	if err != nil {
		return fmt.Errorf("create request packet to authenticate faile: %v", err)
	}

	PostRequest, err := http.NewRequest(http.MethodPost, "https://sessionserver.mojang.com/session/minecraft/join",
		bytes.NewReader(requestPacket))
	if err != nil {
		return fmt.Errorf("make request error: %v", err)
	}
	PostRequest.Header.Set("User-Agent", "gomcbot")
	PostRequest.Header.Set("Connection", "keep-alive")
	resp, err := client.Do(PostRequest)
	if err != nil {
		return fmt.Errorf("post fail: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.Status != "204 No Content" {
		return fmt.Errorf("auth fail: %s", string(body))
	}
	return nil
}

// AES/CFB8 with random key
func NewSymmetricEncryption() (key []byte, encoStream, decoStream cipher.Stream) {
	key = make([]byte, 16)
	rand.Read(key) //生成密钥

	b, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	decoStream = CFB8.NewCFB8Decrypt(b, key)
	encoStream = CFB8.NewCFB8Encrypt(b, key)
	return
}

func GenEncryptionKeyResponse(shareSecret, publicKey, verifyToken []byte) (erp *pk.Packet, err error) {

	iPK, err := x509.ParsePKIXPublicKey(publicKey) // Decode Public Key
	if err != nil {
		err = fmt.Errorf("decode public key fail: %v", err)
		return
	}
	rsaKey := iPK.(*rsa.PublicKey)
	cryptPK, err := rsa.EncryptPKCS1v15(rand.Reader, rsaKey, shareSecret)
	if err != nil {
		err = fmt.Errorf("encryption share secret fail: %v", err)
		return
	}
	verifyT, err := rsa.EncryptPKCS1v15(rand.Reader, rsaKey, verifyToken)
	if err != nil {
		err = fmt.Errorf("encryption verfy tokenfail: %v", err)
		return
	}
	var data []byte
	data = append(data, pk.PackVarInt(int32(len(cryptPK)))...)
	data = append(data, cryptPK...)
	data = append(data, pk.PackVarInt(int32(len(verifyT)))...)
	data = append(data, verifyT...)
	erp = &pk.Packet{
		ID:   0x01,
		Data: data,
	}
	return
}

// NewHandshakePacket 构造了一个Handshake包
func NewHandshakePacket(protocolVersion int, addr string, port int, nextState byte) *pk.Packet {
	var hsData []byte                                                 //Handshake packet data
	hsData = append(hsData, pk.PackVarInt(int32(protocolVersion))...) //Protocol Version
	hsData = append(hsData, pk.PackString(addr)...)
	hsData = append(hsData, pk.PackUint16(uint16(port))...)
	hsData = append(hsData, nextState)
	pack := pk.Packet{
		ID:   0,
		Data: hsData,
	}
	return &pack
}

// newLoginStartPakcket 构造一个LoginStart包
func newLoginStartPacket(userName string) *pk.Packet {
	var data []byte
	data = append(data, pk.PackString(userName)...)
	pack := pk.Packet{
		ID:   0,
		Data: data,
	}
	return &pack
}
func HandleEncryptionRequest(g *Game, pack *pk.Packet, auth *Auth) error {
	//Create AES symmetric encryption key
	key, encoStream, decoStream := NewSymmetricEncryption()

	//解析EncryptionRequest包
	er, err := UnpackEncryptionRequest(*pack)
	if err != nil {
		return err
	}
	err = LoginAuth(auth.AsTk, auth.Name, auth.UUID, key, *er) // Verify to Mojang
	if err != nil {
		return fmt.Errorf("login fail: %v", err)
	}

	// Respond to encryption requests
	var p *pk.Packet // Encryption Key Response
	p, err = GenEncryptionKeyResponse(key, er.PublicKey, er.VerifyToken)
	if err != nil {
		return fmt.Errorf("gen encryption key response fail: %v", err)
	}
	err = g.SendPacket(p)
	if err != nil {
		return err
	}

	// Encrypted connections
	g.Receiver = bufio.NewReader(cipher.StreamReader{ //Set receiver for AES
		S: decoStream,
		R: g.Conn,
	})
	g.Sender = cipher.StreamWriter{
		S: encoStream,
		W: g.Conn,
	}
	return nil
}
