package vless

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/uuid"
)

var requestPool = sync.Pool{
	New: func() interface{} {
		return new(requestInfo)
	},
}

type requestInfo struct {
	effective bool
	// fixed
	fixedData [22]byte
	// Variable
	addons   []byte
	destAddr string
}

func (r *requestInfo) Version() byte {
	r.checkEffective()
	return r.fixedData[0]
}

func (r *requestInfo) UUID() uuid.UUID {
	r.checkEffective()
	uid, _ := uuid.ParseBytes(r.fixedData[1:17])
	return uid
}

func (r *requestInfo) Addons() []byte {
	r.checkEffective()
	return r.addons
}

func (r *requestInfo) Command() protocol.RequestCommand {
	r.checkEffective()
	return protocol.RequestCommand(r.fixedData[18])
}

func (r *requestInfo) DestAddr() string {
	r.checkEffective()
	port := binary.BigEndian.Uint16(r.fixedData[19:21])
	return fmt.Sprintf("%s:%d", r.destAddr, port)
}

func (r *requestInfo) FromReader(reader io.Reader) error {
	r.effective = false
	// read version uuid addonsLength
	_, err := reader.Read(r.fixedData[:18])
	if err != nil {
		return err
	}
	// read addons
	r.addons = make([]byte, r.fixedData[17])
	_, err = reader.Read(r.addons)
	if err != nil {
		return err
	}
	// read command port addrType
	_, err = reader.Read(r.fixedData[18:22])
	if err != nil {
		return err
	}
	// read address
	addrType := r.fixedData[21]
	addrLength, err := getAddrLength(addrType, reader)
	if err != nil {
		return err
	}
	buf := make([]byte, addrLength)
	_, err = reader.Read(buf)
	if err != nil {
		return err
	}
	switch addrType {
	case AddrTypeIPv4:
		r.destAddr = net.IP(buf).To4().String()
	case AddrTypeIPv6:
		r.destAddr = fmt.Sprintf("[%s]", net.IP(buf).To16().String())
	case AddrTypeDomain:
		r.destAddr = string(buf)
	default:
		return fmt.Errorf("not implement addr type %d", addrType)
	}
	r.effective = true
	return nil
}

func (r *requestInfo) checkEffective() {
	if !r.effective {
		panic("request invalidate")
	}
}

func getAddrLength(addrType byte, reader io.Reader) (byte, error) {
	switch addrType {
	case AddrTypeIPv4:
		return 4, nil
	case AddrTypeIPv6:
		return 16, nil
	case AddrTypeDomain:
		domainLength := make([]byte, 1)
		_, err := reader.Read(domainLength)
		if err != nil {
			return 0, err
		}
		return domainLength[0], nil
	default:
		return 0, fmt.Errorf("not implement addr type %d", addrType)
	}
}

// 协议详情见 https://github.com/v2ray/v2ray-core/issues/2636

const (
	Version = 0

	AddrTypeIPv4   = 1
	AddrTypeDomain = 2
	AddrTypeIPv6   = 3

	TCP = protocol.RequestCommandTCP // 0x01
	UDP = protocol.RequestCommandUDP // 0x02
	MUX = protocol.RequestCommandMux // 0x03
)

type Requester interface {
	Version() byte
	UUID() uuid.UUID
	Addons() []byte
	Command() protocol.RequestCommand
	DestAddr() string
}

type Response struct {
	Version byte
	Addons  []byte
}

func (r *Response) ToWriter(w io.Writer) error {
	_, err := w.Write([]byte{r.Version, byte(len(r.Addons))})
	if err != nil {
		return err
	}
	_, err = w.Write(r.Addons)
	return err
}
