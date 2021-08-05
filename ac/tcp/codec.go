package tcp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/funny/link"
	"github.com/funny/slab"
)

type protocol struct {
	pool          slab.Pool
	maxPacketSize int
}

func (p *protocol) alloc(size int) []byte {
	return p.pool.Alloc(size)
}

func (p *protocol) free(msg []byte) {
	p.pool.Free(msg)
}

func (p *protocol) sendv(session *link.Session, buffers [][]byte) error {
	err := session.Send(buffers)
	if err != nil {
		session.Close()
	}
	return err
}

func (p *protocol) send(session *link.Session, msg []byte) error {
	err := session.Send(msg)
	if err != nil {
		session.Close()
	}
	return err
}

// =======================================================================

var _ = (link.Codec)((*codec)(nil))

var headFlag []byte                               // 头部标示
var headLen int                                   // 头部长度
var packLenFieldSize = 4                          // 报文长度大小
var packLenFieldOffset = 0                        // 报文偏移
var packLenAddto = 0                              // 报文长度追加，一般是追加校验字段
var packLenIsBodyLen = false                      // 报文长度字段是指主体报文长度
var endian binary.ByteOrder = binary.LittleEndian // 字节序大小端(默认小端)
var errorPackIsKick = true                        // 错误包是否踢出

// SetHeadFlag 设置头部标示
func SetHeadFlag(hf []byte) {
	headFlag = hf
}

// SetHeadLen 设置头部长度
func SetHeadLen(l int) {
	headLen = l
}

// SetpackLenAddto 长度追加
func SetpackLenAddto(l int) {
	packLenAddto = l
}

// SetErrorPackIsKick 设置错误报文是否kick
func SetErrorPackIsKick(isKick bool) {
	errorPackIsKick = isKick
}

// SetEndian 设置大小端
func SetEndian(e binary.ByteOrder) {
	endian = e
}

// SetLenFieldIndex 设置包长度字段所在位置
func SetLenFieldIndex(offset, size int) {
	packLenFieldOffset = offset
	packLenFieldSize = size
}

// SetPackLenIsBodyLen 设置报文长度字段是指主体报文长度
func SetPackLenIsBodyLen() {
	packLenIsBodyLen = true
}

// ErrTooLargePacket 超长数据包
var ErrTooLargePacket = errors.New("长度错误")

// ErrHeadPacket 包头错误
var ErrHeadPacket = errors.New("包头错误")

type codec struct {
	*protocol
	conn     net.Conn
	reader   *bufio.Reader
	headBuf  []byte
	evseid   string // 后台唯一id
	clientid string // pn-sn

	recvSum int32 // 接收数据总和
	sendSum int32 // 发送数据总和
}

func (p *protocol) newCodec(conn net.Conn, bufferSize int) *codec {
	c := &codec{
		protocol: p,
		conn:     conn,
		reader:   bufio.NewReaderSize(conn, bufferSize),
	}
	if headLen != 0 {
		c.headBuf = make([]byte, headLen)
	} else {
		headLen = packLenFieldSize + packLenFieldOffset
		c.headBuf = make([]byte, headLen)
	}
	return c
}

// Receive implements link/Codec.Receive() method.
func (c *codec) Receive() (interface{}, error) {
	if _, err := io.ReadFull(c.reader, c.headBuf); err != nil {
		return nil, err
	}
	hflen := len(headFlag)
	if (hflen == 1 && c.headBuf[0] != headFlag[0]) ||
		(hflen == 2 && endian.Uint16(c.headBuf[:hflen]) != endian.Uint16(headFlag[:hflen])) {
		logrus.Infof("包头错误:[%x]", c.headBuf)
		if errorPackIsKick {
			return nil, ErrHeadPacket
		}
		return nil, nil
	}

	var length int
	if packLenFieldSize == 1 {
		length = int(c.headBuf[packLenFieldOffset])
	} else if packLenFieldSize == 2 {
		length = int(endian.Uint16(c.headBuf[packLenFieldOffset : packLenFieldOffset+packLenFieldSize]))
	} else {
		length = int(endian.Uint32(c.headBuf[packLenFieldOffset : packLenFieldOffset+packLenFieldSize]))
	}

	if length > c.maxPacketSize {
		if errorPackIsKick {
			return nil, ErrTooLargePacket
		}
		return nil, nil
	}
	length += packLenAddto
	if packLenIsBodyLen {
		length += headLen
	}

	atomic.AddInt32(&c.recvSum, int32(length))
	buffer := c.alloc(length)
	copy(buffer, c.headBuf)
	if _, err := io.ReadFull(c.reader, buffer[headLen:]); err != nil {
		c.free(buffer)
		return nil, err
	}
	return &buffer, nil
}

// Send implements link/Codec.Send() method.
func (c *codec) Send(msg interface{}) error {
	if buffers, ok := (msg.([][]byte)); ok {
		netBuf := net.Buffers(buffers)
		l, err := netBuf.WriteTo(c.conn)
		if err == nil {
			atomic.AddInt32(&c.sendSum, int32(l))
		}
		return err
	}
	l, err := c.conn.Write(msg.([]byte))
	if err == nil {
		atomic.AddInt32(&c.sendSum, int32(l))
	}
	return err
}

// Close implements link/Codec.Close() method.
func (c *codec) Close() error {
	return c.conn.Close()
}
