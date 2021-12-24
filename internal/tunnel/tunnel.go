package tunnel

import (
	"code.cloudfoundry.org/lager"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/timotto/ardumower-relay/internal/model"
	"io"
	"net"
	"sync"
	"time"
)

type (
	tunnel struct {
		logger lager.Logger
		param  Parameters
		conn   *websocket.Conn
		user   model.User

		listen model.TunnelListener

		stats model.TunnelStats

		lock  *sync.Mutex
		rdCh  chan string
		rdErr chan error
		wrCh  chan string
		wrErr chan error
		close chan bool

		closed bool
		clLock *sync.Mutex
	}
)

func NewTunnel(logger lager.Logger, param Parameters, conn *websocket.Conn, user model.User) *tunnel {
	a := &tunnel{
		logger: logger.Session("tunnel"),
		param:  param,
		conn:   conn,
		user:   user,

		stats: model.TunnelStats{Created: time.Now()},

		lock:  &sync.Mutex{},
		rdCh:  make(chan string),
		rdErr: make(chan error, 1),
		wrCh:  make(chan string),
		wrErr: make(chan error, 1),
		close: make(chan bool),

		clLock: &sync.Mutex{},
	}

	go a.reader()
	a.writer()

	return a
}

func (t *tunnel) Transfer(ctx context.Context, send string) (string, error) {
	l := t.logger.Session("transfer")

	t.lock.Lock()
	defer t.lock.Unlock()

	if t.isClosed() {
		return "", io.EOF
	}

	go func() {
		t.wrCh <- send
	}()

	select {
	case line, open := <-t.rdCh:
		if !open {
			l.Debug("read-channel-closed")
			return "", io.EOF
		}
		t.stats.TransferCount++

		return line, nil

	case err, open := <-t.rdErr:
		if !open {
			l.Debug("read-error-channel-closed")
			return "", io.EOF
		}
		return "", fmt.Errorf("read failed: %w", err)

	case err, open := <-t.wrErr:
		if !open {
			l.Debug("write-error-channel-closed")
			return "", io.EOF
		}
		return "", fmt.Errorf("write failed: %w", err)

	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (t *tunnel) Close() error {
	_ = t.conn.Close()

	return nil
}

func (t *tunnel) Stats() model.TunnelStats {
	return t.stats
}

func (t *tunnel) SetListener(l model.TunnelListener) {
	t.listen = l
}

func (t *tunnel) reader() {
	l := t.logger.Session("reader")
	defer t.onClose()

	l.Debug("starting")

	for {
		typ, data, err := t.conn.ReadMessage()
		if err != nil {
			t.rdErr <- err
			return
		}

		if typ != websocket.TextMessage {
			l.Error("message-type", fmt.Errorf("unexpected message type: %v", typ))
			continue
		}

		line := string(data)

		select {
		case t.rdCh <- line:
			t.stats.ReadCount++
			continue

		default:
			t.stats.DropCount++
			l.Info("message-dropped")
		}
	}
}

func (t *tunnel) onClose() {
	t.setClosed()

	close(t.close)
	_ = t.conn.Close()
	close(t.rdCh)
	close(t.rdErr)

	if t.listen != nil {
		t.listen.RemoveTunnel(t.user, t)
	}

	t.logger.Debug("closed")
}

func (t *tunnel) writer() {
	l := t.logger.Session("writer")

	t.conn.SetPongHandler(func(_ string) error {
		t.stats.PongRxCount++
		l.Debug("pong-received")

		return nil
	})

	t.conn.SetPingHandler(func(message string) error {
		t.stats.PingRxCount++

		l.Debug("ping-received")
		err := t.conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(t.param.PingTimeout))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		if err == nil {
			t.stats.PongTxCount++
		}

		return err
	})

	tick := time.NewTicker(t.param.PingInterval)

	go func() {
		defer tick.Stop()
		defer func() {
			_ = t.conn.Close()
			close(t.wrErr)
			l.Debug("stopped")
		}()

		for {
			select {
			case now := <-tick.C:
				if err := t.conn.WriteControl(websocket.PingMessage, nil, now.Add(t.param.PingTimeout)); err != nil {
					l.Error("write-control", err)
					t.wrErr <- err
					return
				}
				t.stats.PingTxCount++

			case line := <-t.wrCh:
				if err := t.conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
					l.Error("write-message", err)
					t.wrErr <- err
					return
				}

			case <-t.close:
				l.Debug("close-received")
				return
			}
		}
	}()
}
