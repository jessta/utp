package utp

import (
	//"fmt"
	"log"
	"net"
	"time"
	"bytes"
	"math/rand"
)

// 	http://bittorrent.org/beps/bep_0029.html

type Header struct {
	versionType uint8 //4 bits version/4 bits type
	extension   uint8
	connId      uint16
	time        uint32
	timeDiff    uint32
	windowSize  uint32
	seqNum      uint16
	ackNum      uint16
}

const (
	Data     = 0
	Finalise = 1
	State    = 2
	Reset    = 3
	Syn      = 4
)

//Extension Types
const (
	SelectiveAcks = 1
	ExtensionBits = 2
)

//Conn states
const (
	SynSent   = 1
	Connected = 2
	Closing   = 3
	Closed    = 4
)
const (
	Version uint8 = 1 << 4
)

type packet struct {
	Header
	data []byte
}

func (p *packet) Decode(buf []byte) (*packet,error) {
	
	return p,nil
}

type UTPConn struct {
	localAddr         net.Addr
	remoteAddr        net.Addr
	state             int
	seqNum            uint16
	ackNum		uint16
	connIdSend        uint16
	connIdRecv        uint16
	rrt               time.Duration
	rrt_var           time.Duration
	recieved          []packet
	sent              []packet
	dataBuffer        bytes.Buffer
	maxWindow         int64
	curWindow         int64
	replyMicroseconds time.Duration
	windowSize        int64
	recievechan           <-chan packet
	sendchan              chan<- packet
	eof               *packet
}

type UTPAddr struct {
	net.UDPAddr
}

func NewUTPConn(localAddr net.Addr, remoteAddr net.Addr, recieve <-chan packet) {
	//make([]byte, 1024*1000)
}
func DialUTP(net string, laddr, raddr *UTPAddr) (*UTPConn, error) {
	connId := uint16(rand.Int())
	return &UTPConn{state: SynSent, seqNum: 1, connIdSend: connId, connIdRecv: connId + 1},nil
}

func (c *UTPConn) recv() {
	for {
		p := <-c.recievechan
		kind := (p.versionType << 4) >> 4
		c.HandlePacket(&p)
		switch kind {
		case Data:
			log.Println("recieved data packet")
			c.HandleData(&p)
		case Finalise:
			log.Println("recieved finalise packet")
			c.HandleFinalise(&p)
		case State:
			log.Println("recieved state packet")
			c.HandleState(&p)
		case Reset:
			log.Println("recieved reset packet")
			c.HandleReset(&p)
		case Syn:
			log.Println("recieved syn packet")
			c.HandleSyn(&p)
		}
	}
}
func (c *UTPConn) HandlePacket(p *packet) {

}

func (c *UTPConn) send(p packet) {
	//reply with state packet
	c.sendchan <- p
}

func (c *UTPConn) SynPacket() {
	p := Header{
		versionType:                     Version + Syn,
		extension:                       0,
		connId:                          c.connIdRecv,
		time:           uint32(time.Now().UnixNano()/int64(time.Millisecond)),
		timeDiff: 0,
		windowSize:                      0,
		seqNum:                          c.seqNum,
		ackNum:                          0}

	c.seqNum++
	_ = p
}

func (c *UTPConn) HandleSyn(p *packet) {
	if c.state == Connected {
		//ignore syn packets on already open connections.
		return
	}
	// set up connected state
	c.connIdRecv = p.connId + 1
	c.connIdSend = p.connId
	c.seqNum = uint16(rand.Int())
	c.ackNum = p.seqNum
	c.state = Connected

}

func (c *UTPConn) HandleState(p *packet) {
	c.state = Connected
	c.ackNum = p.seqNum
}

func (c *UTPConn) HandleReset(p *packet) {
	c.state = Closed
}

func (c *UTPConn) HandleFinalise(p *packet) {
	c.eof = p
	c.state = Closing
}

func (c *UTPConn) HandleData(p *packet) {
	//buffer packet for Reads
}

func (c *UTPConn) DataPacket() {
/*	Header{versionType: Version + Data,
		extension:                       0,
		connId:                          c.connIdSend,
		time:           uint32(time.Now().UnixNano()/int64(time.Millisecond)),
		timeDiff: 0,
		windowSize:                      0,
		seqNum:                          c.seqNum,
		ackNum:                          c.ackNum}

	c.seqNum++*/
}

func (c *UTPConn) Read(b []byte) (n int, err error) {
	//Wait for data packets
	return 0,nil
}

func (c *UTPConn) Write(b []byte) (n int, err error) {
	//Send data packets
return 0,nil
}

func (c *UTPConn) Close() error {
	return nil
}

func (c *UTPConn)LocalAddr()net.Addr {
	return c.localAddr
}
func (c *UTPConn)RemoteAddr()net.Addr {
	return c.remoteAddr
}
func (c *UTPConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *UTPConn)   SetReadDeadline(t time.Time) error {
	return nil
}

func (c *UTPConn)  SetWriteDeadline(t time.Time) error {
	return nil
}




type UTPListener struct {
	conn    net.PacketConn
	recieve chan  packet
	send chan packet
}

func (l *UTPListener) run() error{
	var buf [1024 * 64]byte
	for {
		//wait for Syn packet
		n, addr, err := l.conn.ReadFrom(buf[:])
		if n < 1 {
			if err != nil {
				return err
			}
		}
		recv := buf[0:n]
		var p packet
		_, err = p.Decode(recv)
		if err != nil {
			continue
		}
		l.recieve <- p
		_, err = l.conn.WriteTo(recv, addr)
	}
	return nil
}
func (l *UTPListener) SetDeadline(t time.Time) error { return nil }
func (l *UTPListener) Close() error                  { return nil }
func (l *UTPListener) Addr() net.Addr                { return nil }
func (l *UTPListener) Accept() (c net.Conn, err error) {
	return l.AcceptUTP()
}


func (l *UTPListener) AcceptUTP() (c *UTPConn, err error) {
	p := <-l.recieve
	c = new(UTPConn)
	c.connIdRecv = p.connId + 1
	c.connIdSend = p.connId
	c.seqNum = uint16(rand.Int())
	c.ackNum = p.seqNum
	c.state = Connected

	send := packet{Header:Header{versionType: Version + State,
		extension:                       0,
		connId:                          c.connIdSend,
		time:           uint32(time.Now().UnixNano()/int64(time.Millisecond)),
		timeDiff: 0,
		windowSize:                      0,
		seqNum:                          c.seqNum,
		ackNum:                          c.ackNum}}
	l.send <- send
	return c, nil
}

func ListenUTP(net string, laddr *UTPAddr) (*UTPListener, error) {
	return nil, nil
}

type Mux struct {
	net.UDPConn
}

func (*Mux) Recv() {
	// recv UTP packet
	// lookup connection

	// forward packet to connection's buffer
	// foward packet to accept buffer
	// 
}

func (*Mux) Send() {
	// send UTP 
}

const cControlTarget = 100
/*
var ourDelay = p.timestampDifferenceMicroseconds - c.baseDelay
var offTarget = cControlTarget - ourDelay
var delayFactor = offTarget / cControlTarget
var windowFactor = outstandingPacket / maxWindow
var scaledGain = MAX_CWND_INCREASE_PACKETS_PER_RTT * delayFactor * windowFactor
var maxWindow = 0
*/
//var maxWindow += scaledGain
