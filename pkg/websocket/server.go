package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"strings"
)

var (
	keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
	// ErrBadRequestMethod bad request method
	ErrBadRequestMethod = errors.New("bad method")
	// ErrNotWebSocket not websocket protocal
	ErrNotWebSocket = errors.New("not websocket protocol")
	// ErrBadWebSocketVersion bad websocket version
	ErrBadWebSocketVersion = errors.New("missing or bad WebSocket Version")
	// ErrChallengeResponse mismatch challenge response
	ErrChallengeResponse = errors.New("mismatch challenge/response")
)

// Upgrade Switching Protocols
func Upgrade(conn *Conn, req *Request) (err error) {
	if req.Method != "GET" {
		return ErrBadRequestMethod
	}
	if req.Header.Get("Sec-Websocket-Version") != "13" {
		return ErrBadWebSocketVersion
	}
	if strings.ToLower(req.Header.Get("Upgrade")) != "websocket" {
		return ErrNotWebSocket
	}
	if !strings.Contains(strings.ToLower(req.Header.Get("Connection")), "upgrade") {
		return ErrNotWebSocket
	}
	challengeKey := req.Header.Get("Sec-Websocket-Key")
	if challengeKey == "" {
		return ErrChallengeResponse
	}
	_, _ = conn.w.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n")
	_, _ = conn.w.WriteString("Sec-WebSocket-Accept: " + computeAcceptKey(challengeKey) + "\r\n\r\n")
	if err = conn.w.Flush(); err != nil {
		return
	}
	return nil
}

func computeAcceptKey(challengeKey string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(challengeKey))
	_, _ = h.Write(keyGUID)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
