package handler

import (
	"bufio"
	"crypto/cipher"
	"fmt"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

// HandleEncryptionRequest Encryption request
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
