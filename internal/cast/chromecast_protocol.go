package cast

func (c *Chromecast) handle(msg castMessage, payload map[string]any) {
	typ, _ := payload["type"].(string)
	switch {
	case msg.namespace == nsHeartbeat && typ == "PING":
		c.send(msg.source, nsHeartbeat, map[string]any{"type": "PONG"})
		return
	case msg.namespace == nsConnection && typ == "CLOSE":
		c.mu.Lock()
		if msg.source == c.transport {
			c.transport, c.session, c.mediaID = "", "", 0
			c.status.State = StateIdle
		}
		c.mu.Unlock()
	case typ == "RECEIVER_STATUS":
		c.updateReceiver(payload)
	case typ == "MEDIA_STATUS":
		c.updateMedia(payload)
	}
	if id, ok := payload["requestId"].(float64); ok && id > 0 {
		c.mu.Lock()
		ch := c.waiters[int(id)]
		delete(c.waiters, int(id))
		c.mu.Unlock()
		if ch != nil {
			ch <- payload
		}
	}
}

func (c *Chromecast) updateReceiver(p map[string]any) {
	st, _ := p["status"].(map[string]any)
	c.mu.Lock()
	defer c.mu.Unlock()
	if vol, ok := st["volume"].(map[string]any); ok {
		if l, ok := vol["level"].(float64); ok {
			c.status.Volume = l
		}
		c.status.Muted, _ = vol["muted"].(bool)
	}
	apps, _ := st["applications"].([]any)
	for _, a := range apps {
		app, _ := a.(map[string]any)
		if app["appId"] == defaultReceiverApp {
			c.transport, _ = app["transportId"].(string)
			c.session, _ = app["sessionId"].(string)
		}
	}
}

func (c *Chromecast) updateMedia(p map[string]any) {
	list, _ := p["status"].([]any)
	if len(list) == 0 {
		return
	}
	st, _ := list[0].(map[string]any)
	c.mu.Lock()
	defer c.mu.Unlock()
	if id, ok := st["mediaSessionId"].(float64); ok {
		c.mediaID = int(id)
	}
	if t, ok := st["currentTime"].(float64); ok {
		c.status.Position = t
	}
	if m, ok := st["media"].(map[string]any); ok {
		if d, ok := m["duration"].(float64); ok {
			c.status.Duration = d
		}
	}
	switch st["playerState"] {
	case "PLAYING":
		c.status.State = StatePlaying
	case "PAUSED":
		c.status.State = StatePaused
	case "BUFFERING":
		c.status.State = StateBuffering
	case "IDLE":
		c.status.State = StateIdle
		if st["idleReason"] == "FINISHED" || st["idleReason"] == "CANCELLED" {
			c.status.State = StateStopped
		}
	}
}

func (c *Chromecast) snapshot() Status {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}
