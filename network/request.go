package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
)

// NetReqType defines a request's type.
type NetReqType int16

// NetReqType types.
const (
	Message NetReqType = iota + 1
	MessageReceived
	BroadcastMessage
	BroadcastMessageReceived
	Ping
	Pong
	ReqNodeTable
	NodeTable
)

// Request is the data structure exchanged by nodes.
type Request struct {
	Version   [4]byte
	Length    int32 // length of request
	Timestamp int64
	Type      NetReqType
	FromLen   int16
	From      []byte
	Body      []byte
}

// NetVersion is the network's version.
var NetVersion = [4]byte{'i', 'o', 's', 't'}

func isNetVersionMatch(buf []byte) bool {
	if len(buf) >= len(NetVersion) {
		return buf[0] == NetVersion[0] &&
			buf[1] == NetVersion[1] &&
			buf[2] == NetVersion[2] &&
			buf[3] == NetVersion[3]
	}
	return false
}

func newRequest(typ NetReqType, from string, data []byte) *Request {
	r := &Request{
		Version:   NetVersion,
		Timestamp: time.Now().UnixNano(),
		Type:      typ,
		FromLen:   int16(len(from)),
		From:      []byte(from),
		Body:      data,
	}
	//len(timestamp) + len(type) + len(fromLen) + len(from) + len(body)
	r.Length = int32(8 + 2 + 2 + len(r.From) + len(data))

	return r
}

// Pack serializes a request to bytes.
func (r *Request) Pack() ([]byte, error) {
	var err error
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, &r.Version)
	err = binary.Write(buf, binary.BigEndian, &r.Length)
	err = binary.Write(buf, binary.BigEndian, &r.Timestamp)
	err = binary.Write(buf, binary.BigEndian, &r.Type)
	err = binary.Write(buf, binary.BigEndian, &r.FromLen)
	err = binary.Write(buf, binary.BigEndian, &r.From)
	err = binary.Write(buf, binary.BigEndian, &r.Body)
	return buf.Bytes(), err
}

// Unpack unserializes bytes.
func (r *Request) Unpack(reader io.Reader) error {
	var err error
	err = binary.Read(reader, binary.BigEndian, &r.Version)
	err = binary.Read(reader, binary.BigEndian, &r.Length)
	err = binary.Read(reader, binary.BigEndian, &r.Timestamp)
	err = binary.Read(reader, binary.BigEndian, &r.Type)
	err = binary.Read(reader, binary.BigEndian, &r.FromLen)
	r.From = make([]byte, r.FromLen)
	err = binary.Read(reader, binary.BigEndian, &r.From)
	r.Body = make([]byte, r.Length-8-2-2-int32(r.FromLen))
	err = binary.Read(reader, binary.BigEndian, &r.Body)
	return err
}

// String implements fmt.Stringer.
func (r *Request) String() string {
	return fmt.Sprintf("version:%s length:%d type:%d timestamp:%s from:%s Body:%v",
		r.Version,
		r.Length,
		r.Type,
		time.Unix(r.Timestamp/1e9, r.Timestamp%1e9).Format("2006-01-02 15:04:05"),
		r.From,
		r.Body,
	)
}

func prometheusReceivedBlockTx(req *message.Message) {
	if req.ReqType == int32(ReqPublishTx) {
		receivedBroadTransactionCount.Inc()
	}
	if req.ReqType == int32(ReqNewBlock) {
	}
}

func (r *Request) handle(base *BaseNetwork, conn net.Conn) {
	switch r.Type {
	case Message:
		appReq := &message.Message{}
		if _, err := appReq.Unmarshal(r.Body); err == nil {
			base.log.D("[net] msg from =%v, to = %v, typ = %v,  ttl = %v", appReq.From, appReq.To, appReq.ReqType, appReq.TTL)
			base.RecvCh <- *appReq
			prometheusReceivedBlockTx(appReq)

		} else {
			base.log.E("[net] failed to unmarshal recv msg:%v, err:%v", r, err)
		}
		r.msgHandle(base)
	case MessageReceived:
		base.log.D("[net] MessageReceived: %v", string(r.From), common.BytesToInt64(r.Body))
	case BroadcastMessage:
		appReq := &message.Message{}
		if _, err := appReq.Unmarshal(r.Body); err == nil {
			if appReq.ReqType == int32(ReqBlockHeight) {
				appReq.From = string(r.From)
			}
			base.RecvCh <- *appReq

			prometheusReceivedBlockTx(appReq)
			if appReq.ReqType != int32(ReqDownloadBlock) {
				base.Broadcast(*appReq)
			}
		}
		r.msgHandle(base)
	case BroadcastMessageReceived:
	//request for nodeTable
	case ReqNodeTable:

		if isValidNode(r, base) {
			base.putNode(string(r.From))
			base.peers.SetAddr(string(r.From), newPeer(conn, nil, base.localNode.Addr(), conn.RemoteAddr().String()))
			base.sendNodeTable(r.From, conn)
		} else {
			conn.Close()
		}
	//got nodeTable and save
	case NodeTable:
		base.log.D("[net] response node table: %v", string(r.Body))
		base.putNode(string(r.Body))
	default:
		base.log.E("[net] wrong request :", r)
	}
}

func isValidNode(r *Request, base *BaseNetwork) bool {
	strs := strings.Split(string(r.From), ":")
	if NetMode == PublicMode && !common.IsPublicIP(net.ParseIP(strs[0])) {
		base.log.D("[net] the node's ip is not public ip: %v", strs[0])
		return false
	}
	return true
}

func (r *Request) msgHandle(net *BaseNetwork) {
	msg := &message.Message{}
	if _, err := msg.Unmarshal(r.Body); err == nil {
		switch msg.ReqType {
		case int32(RecvBlockHeight):
			var rh message.ResponseHeight
			rh.Decode(msg.Body)
			net.SetNodeHeightMap(string(r.From), rh.BlockHeight)
		default:
		}
	}
}
