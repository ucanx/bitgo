package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	TestNetMagic    = 0x0709110B // magic value for testnet
	MainNetMagic    = 0xD9B4BEF9 // magic value for mainnet
	CommandLength   = 12
	ProtocolVersion = 70015
)

type NetAddress struct {
	Services uint64
	IP       [16]byte
	Port     uint16
}

type VersionMessage struct {
	Version     int32
	Services    uint64
	Timestamp   int64
	AddrRecv    NetAddress
	AddrFrom    NetAddress
	Nonce       uint64
	UserAgent   string
	StartHeight int32
	Relay       bool
}

// init initializes the logger
func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	log.Println("Starting the Bitcoin node handshake client...")
	conn, err := net.Dial("tcp", "testnet-seed.bitcoin.jonasschnelli.ch:18333")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	log.Println("Connected to the node.")

	versionMsg := createVersionMessage()
	msg := serializeVersionMsg(versionMsg)

	_, err = conn.Write(msg)
	if err != nil {
		log.Fatalf("Failed to send version message: %v", err)
	}
	log.Println("Version message sent.")

	handleIncomingMessages(conn)
}

// createVersionMessage creates a new VersionMessage object
func createVersionMessage() VersionMessage {
	return VersionMessage{
		Version:     ProtocolVersion,
		Services:    1,
		Timestamp:   time.Now().Unix(),
		AddrRecv:    NewNetAddress(net.ParseIP("127.0.0.1"), 18333),
		AddrFrom:    NewNetAddress(net.ParseIP("127.0.0.1"), 8333),
		Nonce:       0,
		UserAgent:   "/MyNode:0.1/",
		StartHeight: 0,
		Relay:       false,
	}
}

// NewNetAddress creates a new NetAddress object
func NewNetAddress(ip net.IP, port uint16) NetAddress {
	var addr NetAddress
	addr.Services = 1 // NODE_NETWORK
	copy(addr.IP[:], ip.To16())
	addr.Port = port
	return addr
}

// serializeVersionMsg serializes a VersionMessage into a byte slice
func serializeVersionMsg(msg VersionMessage) []byte {
	var result bytes.Buffer
	binary.Write(&result, binary.LittleEndian, msg.Version)
	binary.Write(&result, binary.LittleEndian, msg.Services)
	binary.Write(&result, binary.LittleEndian, msg.Timestamp)
	binary.Write(&result, binary.LittleEndian, msg.AddrRecv)
	binary.Write(&result, binary.LittleEndian, msg.AddrFrom)
	binary.Write(&result, binary.LittleEndian, msg.Nonce)

	result.WriteByte(byte(len(msg.UserAgent)))
	result.WriteString(msg.UserAgent)

	binary.Write(&result, binary.LittleEndian, msg.StartHeight)
	binary.Write(&result, binary.LittleEndian, msg.Relay)

	payload := result.Bytes()
	var buffer bytes.Buffer
	buffer.Write([]byte{0x0B, 0x11, 0x09, 0x07}) // Test network magic
	// buffer.Write([]byte{0xF9, 0xBE, 0xB4, 0xD9}) // Main network magic
	buffer.Write(padCommandName("version"))
	binary.Write(&buffer, binary.LittleEndian, uint32(len(payload)))
	buffer.Write(payload)

	log.Printf("Sending serialized version message: %x", buffer.Bytes())

	return buffer.Bytes()
}

// padCommandName pads the command name with null bytes to make it 12 bytes long
func padCommandName(cmd string) []byte {
	var padded [CommandLength]byte
	copy(padded[:], cmd)
	return padded[:]
}

// handleIncomingMessages reads and processes incoming messages from the node
func handleIncomingMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		header := make([]byte, 24)
		_, err := io.ReadFull(reader, header)
		if err != nil {
			log.Printf("Error reading header: %v", err)
			break
		}
		command := strings.TrimSpace(string(header[4:16]))
		payloadLength := binary.LittleEndian.Uint32(header[16:20])
		log.Printf("Received command: %s, Payload length: %d", command, payloadLength)

		payload := make([]byte, payloadLength)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			log.Printf("Error reading payload: %v", err)
			break
		}
		log.Printf("Received payload for command %s", command)

		if command == "version" {
			log.Println("Processing version message")
			sendVerack(conn)
		} else if command == "verack" {
			log.Println("Received verack, handshake complete.")
			break // Handshake completed successfully
		}
	}
}

// sendVerack sends a verack message to the node
func sendVerack(conn net.Conn) {
	var buffer bytes.Buffer
	buffer.Write([]byte{0x0B, 0x11, 0x09, 0x07}) // Test network magic
	// buffer.Write([]byte{0xF9, 0xBE, 0xB4, 0xD9}) // Main network magic
	buffer.Write(padCommandName("verack"))
	binary.Write(&buffer, binary.LittleEndian, uint32(0)) // No payload
	_, err := conn.Write(buffer.Bytes())
	if err != nil {
		log.Printf("Failed to send verack: %v", err)
	} else {
		log.Println("Sent verack")
	}
}
