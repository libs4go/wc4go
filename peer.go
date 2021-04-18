package wc

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/libs4go/errors"
	"github.com/libs4go/slf4go"
)

// Peer wallet connect peer
type Peer struct {
	slf4go.Logger
	conn        *websocket.Conn
	session     *Session
	isServer    bool
	peerInfo    *PeerInfo
	peerID      string
	chainID     int64
	accounts    []string
	handshakeID int64
}

// NewServer create server peer
func NewServer(url string, chainID int64, accounts []string, peerInfo *PeerInfo) (*Peer, error) {

	session, err := NewServerSession(url)

	if err != nil {
		return nil, err
	}

	var bridge = session.URL.Bridge

	if strings.HasPrefix(bridge, "http://") {
		bridge = "ws" + strings.TrimPrefix(bridge, "http")
	} else if strings.HasPrefix(bridge, "https://") {
		bridge = "wss" + strings.TrimPrefix(bridge, "https")
	}

	peer := &Peer{
		Logger:   slf4go.Get("wc-peer-server"),
		session:  session,
		isServer: true,
		peerInfo: peerInfo,
		peerID:   uuid.New().String(),
		chainID:  chainID,
		accounts: accounts,
	}

	peer.D("dial to {@url}", bridge)

	conn, _, err := websocket.DefaultDialer.Dial(bridge, nil)

	if err != nil {
		return nil, errors.Wrap(err, "dial to websocket server %s error", session.URL.Bridge)
	}

	peer.D("dial to {@url} -- success", bridge)

	peer.conn = conn

	return peer, nil
}

// Handshake complete handshake
func (peer *Peer) Handshake() error {
	if peer.isServer {
		return peer.serverHandshake()
	}

	return nil
}

// HandshakeApprove .
func (peer *Peer) HandshakeApprove(approved bool) error {

	rsp := &sessionResponse{
		PeerID:   peer.peerID,
		PeerMeta: peer.peerInfo,
		ChainID:  peer.chainID,
		Approved: approved,
		Accounts: peer.accounts,
	}

	rpc := &jsonRPCResponse{
		ID:      peer.handshakeID,
		JSONRPC: "2.0",
		Result:  rsp,
	}

	buff, err := json.Marshal(rpc)

	if err != nil {
		return errors.Wrap(err, "marshal sessionResponse error")
	}

	return peer.WriteMessage(buff, false)
}

func (peer *Peer) serverHandshake() error {
	msg, err := peer.session.HandshakeSubscribe()

	if err != nil {
		return err
	}

	println("send handshake")

	peer.D("send handshake {@buff}", string(msg))

	err = peer.conn.WriteMessage(websocket.TextMessage, msg)

	if err != nil {
		return errors.Wrap(err, "websocket send message error")
	}

	peer.D("send handshake -- success")

	buff, err := peer.ReadMessage()

	if err != nil {
		return err
	}

	peer.D("handshake recv {@buff}", string(buff))

	var request *jsonRPCRequest

	err = json.Unmarshal(buff, &request)

	if err != nil {
		return errors.Wrap(err, "unmarshal handshake request error")
	}

	if request.Method != "wc_sessionRequest" {
		return errors.Wrap(ErrMessage, "expect wc_sessionRequest but got %s", request.Method)
	}

	if len(request.Params) != 1 {
		return errors.Wrap(ErrFormat, "wc_sessionRequest params number must be 1")
	}

	buff, err = json.Marshal(request.Params[0])

	if err != nil {
		return errors.Wrap(err, "marshal sessionRequest request error")
	}

	var sr *sessionRequest

	err = json.Unmarshal(buff, &sr)

	if err != nil {
		return errors.Wrap(err, "unmarshal sessionRequest request error")
	}

	peer.session.Peer = sr.PeerID
	peer.session.PeerInfo = &sr.PeerMeta
	peer.handshakeID = request.ID

	return nil
}

// ReadMessage read message from websocket conn
func (peer *Peer) ReadMessage() ([]byte, error) {

	for {
		t, message, err := peer.conn.ReadMessage()

		if err != nil {
			return nil, errors.Wrap(err, "read from websocket error")
		}

		if t != websocket.TextMessage {
			peer.D("recv message {@type} -- skip", t)
			continue
		}

		peer.D("recv message {@msg}", string(message))

		return peer.session.HandleSubscribe(message)

	}

}

// WriteMessage .
func (peer *Peer) WriteMessage(data []byte, handshake bool) error {
	var buff []byte

	var err error

	peer.D("publish message {@buff}", string(data))

	if handshake {
		buff, err = peer.session.HandshakePublish(data)
	} else {
		buff, err = peer.session.Publish(data)
	}

	if err != nil {
		return err
	}

	peer.D("publish encrypt message {@buff}", string(buff))

	err = peer.conn.WriteMessage(websocket.TextMessage, buff)

	if err != nil {
		return errors.Wrap(err, "websocket send message error")
	}

	return nil
}
