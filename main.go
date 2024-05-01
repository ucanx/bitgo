package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	ProtocolVersion = 70015 // Adjust the protocol version as needed
)

func main() {
	// Connect to the Bitcoin node
	// fixme: // Replace with the actual IP address of the Bitcoin node
	conn, err := net.Dial("tcp", "bitcoin.node.ip:8333")
	if err != nil {
		fmt.Println("Error connecting to Bitcoin node:", err)
		return
	}
	defer conn.Close()

	// Perform handshake
	err = performHandshake(conn)
	if err != nil {
		fmt.Println("Handshake failed:", err)
		return
	}

	fmt.Println("Handshake successful!")
}

func performHandshake(conn net.Conn) error {
	// Construct version message
	version := make([]byte, 85) // 85 bytes for version message
	binary.LittleEndian.PutUint32(version[:4], ProtocolVersion)

	// fixme: Add other fields of the version message as per Bitcoin protocol specification
	// ...

	// Send version message
	_, err := conn.Write(version)
	if err != nil {
		return err
	}

	// Receive verack message
	verack := make([]byte, 24) // 24 bytes for verack message
	_, err = conn.Read(verack)
	if err != nil {
		return err
	}

	// Validate verack message
	if !bytes.Equal(verack[:4], []byte{0xf9, 0xbe, 0xb4, 0xd9}) || !bytes.Equal(verack[4:16], []byte{0x76, 0x65, 0x72, 0x61, 0x63, 0x6b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) || !bytes.Equal(verack[16:24], []byte{0x0, 0x0, 0x0, 0x0}) {
		return fmt.Errorf("invalid verack message received")
	}

	return nil
}
