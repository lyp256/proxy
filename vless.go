package proxy

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	VLESSVersion = 0

	AddrTypeIPv4   = 1
	AddrTypeDomain = 2
	AddrTypeIPv6   = 3

	VLESSCommandTCP = 0x01
	VLESSCommandUDP = 0x02
	VLESSCommandMux = 0x03
)

type VLESSRequest struct {
	effective bool
	// fixed
	fixedData [22]byte
	// Variable
	addons   []byte
	destAddr string
}

func (r *VLESSRequest) Version() byte {
	r.checkEffective()
	return r.fixedData[0]
}
func (r *VLESSRequest) UUID() (uuid [16]byte) {
	r.checkEffective()
	copy(uuid[:], r.fixedData[1:18])
	return
}
func (r *VLESSRequest) Addons() []byte {
	r.checkEffective()
	return r.addons
}
func (r *VLESSRequest) Command() byte {
	r.checkEffective()
	return r.fixedData[18]
}

func (r *VLESSRequest) DestAddr() string {
	r.checkEffective()
	port := binary.BigEndian.Uint16(r.fixedData[19:21])
	return fmt.Sprintf("%s:%d", r.destAddr, port)
}

func (r *VLESSRequest) FromReader(reader io.Reader) error {
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

func (r *VLESSRequest) checkEffective() {
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

type VLESSResponse struct {
	Version byte
	Addons  []byte
}

func (r *VLESSResponse) ToWriter(w io.Writer) error {
	_, err := w.Write([]byte{r.Version, byte(len(r.Addons))})
	if err != nil {
		return err
	}
	_, err = w.Write(r.Addons)
	return err
}

func NewVLESSHandler() http.Handler {
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             false,
		TimestampFormat:           "15:04:05",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	})
	return &vlessHandler{
		log: logrus.StandardLogger(),
		dialer: &net.Dialer{
			Timeout:       time.Second * 10,
			Deadline:      time.Time{},
			LocalAddr:     nil,
			FallbackDelay: 0,
			KeepAlive:     0,
			Resolver:      nil,
			Control:       nil,
		},
	}
}

type vlessHandler struct {
	log    logrus.Ext1FieldLogger
	dialer *net.Dialer
}

func (h *vlessHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.log.Debugf("%s %s %s", request.RemoteAddr, request.Method, request.URL.String())
	writer.WriteHeader(http.StatusOK)
	if f, ok := writer.(http.Flusher); ok {
		f.Flush()
	}
	downstream, err := hijack(writer, request)
	if err != nil {
		h.log.Error(err)
		return
	}
	defer downstream.Close()

	vReq := vlessRequestPool.Get().(*VLESSRequest)
	defer vlessRequestPool.Put(vReq)
	err = vReq.FromReader(downstream)
	if err != nil {
		logrus.Error(err)
		return
	}

	var network string
	switch vReq.Command() {
	case VLESSCommandTCP:
		network = "tcp"
	case VLESSCommandUDP:
		network = "udp"
	default:
		h.log.Errorf("not support Command %d", vReq.Command())
		return
	}
	h.log.Debugf("dial:%s %s", network, vReq.DestAddr())
	upStream, err := h.dialer.DialContext(context.Background(), network, vReq.DestAddr())
	if err != nil {
		h.log.Errorf("dial %s %s fail:%s", network, vReq.DestAddr(), err)
		return
	}
	defer upStream.Close()
	resp := VLESSResponse{
		Version: vReq.Version(),
		Addons:  nil,
	}
	err = resp.ToWriter(downstream)
	if err != nil {
		h.log.Error(err)
		return
	}
	up, down, err := transport(upStream, downstream)
	if err != nil && !errors.Is(err, io.EOF) {
		h.log.Error(err)
	}
	h.log.Debugf("up:%d down:%d", up, down)
}

var vlessRequestPool = sync.Pool{
	New: func() interface{} {
		return new(VLESSRequest)
	},
}
