package cast

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// launch starts the Default Media Receiver, if it isn't running.
func (c *Chromecast) launch(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	c.mu.Lock()
	running := c.transport != ""
	c.mu.Unlock()
	if !running {
		if _, err := c.request(ctx, receiverID, nsReceiver, map[string]any{"type": "LAUNCH", "appId": defaultReceiverApp}); err != nil {
			return err
		}
		// The launch reply is a RECEIVER_STATUS that already carries the
		// app; wait briefly if it didn't.
		for i := 0; i < 20; i++ {
			c.mu.Lock()
			running = c.transport != ""
			c.mu.Unlock()
			if running {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if !running {
			return fmt.Errorf("cast: %s did not start the media player", c.name)
		}
	}
	c.mu.Lock()
	transport := c.transport
	c.mu.Unlock()
	return c.send(transport, nsConnection, map[string]any{"type": "CONNECT"})
}

func (c *Chromecast) Load(ctx context.Context, m Media) error {
	if err := c.launch(ctx); err != nil {
		return err
	}
	media := map[string]any{
		"contentId":   m.URL,
		"contentType": m.ContentType,
		"streamType":  "BUFFERED",
		"metadata":    map[string]any{"metadataType": 0, "title": m.Title},
	}
	if m.Duration > 0 {
		media["duration"] = m.Duration
	}
	load := map[string]any{"type": "LOAD", "media": media, "autoplay": true, "currentTime": m.Start}
	if m.Subtitles != "" {
		media["tracks"] = []any{map[string]any{
			"trackId": 1, "type": "TEXT", "subtype": "SUBTITLES", "name": "Subtitles",
			"trackContentId": m.Subtitles, "trackContentType": "text/vtt", "language": m.SubtitleLang,
		}}
		media["textTrackStyle"] = map[string]any{"backgroundColor": "#00000000", "edgeType": "OUTLINE", "edgeColor": "#000000FF"}
		load["activeTrackIds"] = []int{1}
	}
	c.mu.Lock()
	transport, session := c.transport, c.session
	c.status.State = StateLoading
	c.mu.Unlock()
	load["sessionId"] = session
	_, err := c.request(ctx, transport, nsMedia, load)
	return err
}

func (c *Chromecast) mediaCommand(typ string, extra map[string]any) error {
	c.mu.Lock()
	transport, id := c.transport, c.mediaID
	c.mu.Unlock()
	if transport == "" || id == 0 {
		return errors.New("cast: nothing is playing")
	}
	p := map[string]any{"type": typ, "mediaSessionId": id}
	for k, v := range extra {
		p[k] = v
	}
	_, err := c.request(context.Background(), transport, nsMedia, p)
	return err
}

func (c *Chromecast) Play() error  { return c.mediaCommand("PLAY", nil) }
func (c *Chromecast) Pause() error { return c.mediaCommand("PAUSE", nil) }
func (c *Chromecast) Seek(s float64) error {
	return c.mediaCommand("SEEK", map[string]any{"currentTime": s})
}

// Stop ends playback and closes the media app, returning the TV to its
// home screen.
func (c *Chromecast) Stop() error {
	c.mu.Lock()
	session := c.session
	c.mu.Unlock()
	if session == "" {
		return nil
	}
	_, err := c.request(context.Background(), receiverID, nsReceiver, map[string]any{"type": "STOP", "sessionId": session})
	return err
}

func (c *Chromecast) SetVolume(level float64) error {
	if err := c.connect(context.Background()); err != nil {
		return err
	}
	_, err := c.request(context.Background(), receiverID, nsReceiver, map[string]any{
		"type": "SET_VOLUME", "volume": map[string]any{"level": max(0, min(1, level))},
	})
	return err
}

func (c *Chromecast) Status(ctx context.Context) (Status, error) {
	c.mu.Lock()
	transport, id := c.transport, c.mediaID
	c.mu.Unlock()
	if transport != "" {
		p := map[string]any{"type": "GET_STATUS"}
		if id != 0 {
			p["mediaSessionId"] = id
		}
		if _, err := c.request(ctx, transport, nsMedia, p); err != nil {
			return c.snapshot(), err
		}
	}
	return c.snapshot(), nil
}

func (c *Chromecast) Close() error {
	c.drop()
	return nil
}
