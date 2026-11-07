package cast

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// connect opens the TLS connection and the receiver channel, once.
func (c *Chromecast) connect(ctx context.Context) error {
	c.mu.Lock()
	if c.conn != nil {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()
	d := tls.Dialer{Config: &tls.Config{InsecureSkipVerify: true}} // devices use self-signed certificates
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("cast: connect to %s: %w", c.name, err)
	}
	c.mu.Lock()
	c.conn = conn
	c.w = bufio.NewWriter(conn)
	c.waiters = map[int]chan map[string]any{}
	c.closed = make(chan struct{})
	c.status.Volume = -1
	c.mu.Unlock()
	go c.readLoop(conn)
	go c.heartbeat()
	return c.send(receiverID, nsConnection, map[string]any{"type": "CONNECT"})
}

func (c *Chromecast) heartbeat() {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-c.closed:
			return
		case <-t.C:
			if c.send(receiverID, nsHeartbeat, map[string]any{"type": "PING"}) != nil {
				return
			}
		}
	}
}

func (c *Chromecast) readLoop(conn net.Conn) {
	r := bufio.NewReader(conn)
	for {
		var size uint32
		if err := binary.Read(r, binary.BigEndian, &size); err != nil {
			c.drop()
			return
		}
		buf := make([]byte, size)
		if _, err := io.ReadFull(r, buf); err != nil {
			c.drop()
			return
		}
		msg, err := decodeCastMessage(buf)
		if err != nil {
			continue
		}
		var payload map[string]any
		if json.Unmarshal([]byte(msg.payload), &payload) != nil {
			continue
		}
		c.handle(msg, payload)
	}
}

func (c *Chromecast) drop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return
	}
	c.conn.Close()
	c.conn = nil
	close(c.closed)
	for id, ch := range c.waiters {
		close(ch)
		delete(c.waiters, id)
	}
	c.transport, c.session, c.mediaID = "", "", 0
	c.status.State = StateIdle
}

func (c *Chromecast) send(dest, ns string, payload map[string]any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	frame := encodeCastMessage(castMessage{source: senderID, dest: dest, namespace: ns, payload: string(data)})
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return errors.New("cast: not connected")
	}
	c.conn.SetWriteDeadline(time.Now().Add(requestTimeout))
	binary.Write(c.w, binary.BigEndian, uint32(len(frame)))
	c.w.Write(frame)
	return c.w.Flush()
}

// request sends a payload with a fresh requestId and waits for the reply.
func (c *Chromecast) request(ctx context.Context, dest, ns string, payload map[string]any) (map[string]any, error) {
	c.mu.Lock()
	c.requestID++
	id := c.requestID
	ch := make(chan map[string]any, 1)
	if c.waiters == nil {
		c.mu.Unlock()
		return nil, errors.New("cast: not connected")
	}
	c.waiters[id] = ch
	c.mu.Unlock()
	payload["requestId"] = id
	if err := c.send(dest, ns, payload); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	select {
	case reply, ok := <-ch:
		if !ok {
			return nil, errors.New("cast: connection lost")
		}
		if t, _ := reply["type"].(string); t == "LOAD_FAILED" || t == "LAUNCH_ERROR" || t == "INVALID_REQUEST" {
			reason, _ := reply["reason"].(string)
			if reason == "" {
				reason = t
			}
			return reply, fmt.Errorf("cast: %s refused: %s", c.name, reason)
		}
		return reply, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.waiters, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("cast: %s did not answer", c.name)
	}
}
