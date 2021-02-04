package wc

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libs4go/errors"
)

// Session wallet connect session object
type Session struct {
	URL      *URL      `json:"url"`
	Peer     string    `json:"peer"`
	PeerInfo *PeerInfo `json:"peer-info"`
	key      []byte
}

// PeerInfo the peer information description
type PeerInfo struct {
	Description string   `json:"description"`
	URL         string   `json:"url,omitempty"`
	ICONs       []string `json:"icons,omitempty"`
	Name        string   `json:"name"`
}

// NewServerSession create server session with request connection url
func NewServerSession(url string) (*Session, error) {

	u, err := ParseURL(url)

	if err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(u.Key)

	if err != nil {
		return nil, errors.Wrap(err, "decode %s key error", url)
	}

	return &Session{
		URL: u,
		key: key,
	}, nil
}

// HandshakeSubscribe generate session request subscribe message
func (s *Session) HandshakeSubscribe() ([]byte, error) {
	return s.sub(s.URL.Topic)
}

// HandshakePublish generate handshake publish message
func (s *Session) HandshakePublish(data []byte) ([]byte, error) {
	return s.publish(s.URL.Topic, data)
}

// Publish .
func (s *Session) Publish(data []byte) ([]byte, error) {
	return s.publish(s.Peer, data)
}

// HandleSubscribe decrypt subscribe encryption data payload
func (s *Session) HandleSubscribe(data []byte) ([]byte, error) {
	var msg *socketMessage

	err := json.Unmarshal(data, &msg)

	if err != nil {
		return nil, errors.Wrap(err, "unmarshal socketMessage error %s", string(data))
	}

	var encryptData *encryptionPayload

	err = json.Unmarshal([]byte(msg.Payload), &encryptData)

	if err != nil {
		return nil, errors.Wrap(err, "unmarshal encryptionPayload error %s", msg.Payload)
	}

	buff, err := encryptData.decrypt(s.key)

	return buff, err
}

func (s *Session) sub(topic string) ([]byte, error) {
	msg := &socketMessage{
		Topic:   topic,
		Type:    "sub",
		Payload: "",
	}

	buff, err := json.Marshal(msg)

	if err != nil {
		return nil, errors.Wrap(err, "marshal socketMessage error")
	}

	return buff, nil
}

func (s *Session) publish(topic string, data []byte) ([]byte, error) {

	var msg *socketMessage = nil

	if len(data) != 0 {
		encryptData, err := encrypt(data, s.key)

		if err != nil {
			return nil, err
		}

		payload, err := json.Marshal(encryptData)

		if err != nil {
			return nil, errors.Wrap(err, "marshal encryptionPayload error")
		}

		msg = &socketMessage{
			Topic:   topic,
			Type:    "pub",
			Payload: string(payload),
		}

	} else {
		msg = &socketMessage{
			Topic:   topic,
			Type:    "pub",
			Payload: "",
		}
	}

	buff, err := json.Marshal(msg)

	if err != nil {
		return nil, errors.Wrap(err, "marshal socketMessage error")
	}

	return buff, nil
}
