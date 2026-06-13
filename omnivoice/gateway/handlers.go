package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// handleInboundCall handles incoming Twilio webhook for new calls.
// Returns TwiML to connect the call to Media Streams.
func (g *Gateway) handleInboundCall(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10) // 64KB
	if err := r.ParseForm(); err != nil {
		g.logger.Error("failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Get caller info from Twilio webhook
	callSID := sanitize(r.Form.Get("CallSid"))
	from := sanitize(r.Form.Get("From"))
	to := sanitize(r.Form.Get("To"))
	direction := sanitize(r.Form.Get("Direction"))

	g.logger.Info("incoming call",
		"call_sid", callSID,
		"from", from,
		"to", to,
		"direction", direction)

	// Create session
	session := g.createSession(callSID, from, to, direction)

	// Check if call should be accepted
	g.mu.RLock()
	handler := g.callHandler
	g.mu.RUnlock()

	if handler != nil {
		callInfo := &CallInfo{
			CallID:    callSID,
			From:      from,
			To:        to,
			Direction: direction,
			StartTime: session.startTime,
		}
		if err := handler(callInfo); err != nil {
			g.logger.Info("call rejected by handler", "call_sid", callSID, "error", err)
			g.removeSession(callSID)
			// Return TwiML to reject the call
			twiml := `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Say>Sorry, this call cannot be accepted at this time.</Say>
    <Hangup/>
</Response>`
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(twiml))
			return
		}
	}

	// Build WebSocket URL for Media Streams
	wsURL := g.config.PublicURL + "/ws/media-stream"
	// Convert https:// to wss://
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)

	// Return TwiML for Media Streams (bidirectional for full-duplex audio)
	twiml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Connect>
        <Stream url="%s" mode="bidirectional">
            <Parameter name="callSid" value="%s"/>
            <Parameter name="from" value="%s"/>
            <Parameter name="to" value="%s"/>
        </Stream>
    </Connect>
</Response>`, wsURL, callSID, from, to)

	w.Header().Set("Content-Type", "application/xml")
	//nolint:gosec // G705: TwiML response generated from code, not user input
	if _, err := w.Write([]byte(twiml)); err != nil {
		g.logger.Error("failed to write TwiML response", "error", err, "call_sid", callSID)
	}
}

// handleCallStatus handles Twilio status callbacks.
func (g *Gateway) handleCallStatus(w http.ResponseWriter, r *http.Request) {
	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	callSID := sanitize(r.Form.Get("CallSid"))
	status := sanitize(r.Form.Get("CallStatus"))
	duration := sanitize(r.Form.Get("CallDuration"))

	g.logger.Info("call status update",
		"call_sid", callSID,
		"status", status,
		"duration", duration)

	// Update call system
	g.callSystem.HandleStatusCallback(callSID, status)

	// If call ended, clean up session
	if status == "completed" || status == "failed" || status == "busy" || status == "no-answer" || status == "canceled" {
		if session, ok := g.getSessionInternal(callSID); ok {
			session.emitEvent(EventSessionEnded, map[string]string{
				"status":   status,
				"duration": duration,
			})
			_ = session.Close()
		}
	}

	w.WriteHeader(http.StatusOK)
}

// handleMediaStream handles the WebSocket connection for Media Streams.
func (g *Gateway) handleMediaStream(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	wsConn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		g.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	g.logger.Info("media stream websocket connected",
		"remote_addr", r.RemoteAddr)

	// Create connection wrapper
	// Audio buffers sized for ~10 seconds of 8kHz μ-law audio (500 chunks * 20ms)
	conn := &mediaStreamConn{
		wsConn:        wsConn,
		gateway:       g,
		events:        make(chan Event, 100),
		audioIn:       make(chan []byte, 250), // ~5 sec input buffer
		audioOut:      make(chan []byte, 500), // ~10 sec output buffer
		done:          make(chan struct{}),
		sessionReady:  make(chan struct{}),
		pipelineReady: make(chan struct{}),
	}

	// Start read/write loops
	go conn.readLoop()
	go conn.writeLoop()

	// Wait for session to be associated (from "start" message)
	select {
	case <-conn.done:
		return
	case <-time.After(30 * time.Second):
		g.logger.Warn("no start message received, closing connection")
		_ = wsConn.Close()
		return
	case <-conn.sessionReady:
		// Session is ready, start the pipeline
	}

	// Get the session
	if conn.session == nil {
		g.logger.Error("session not set after start message")
		_ = wsConn.Close()
		return
	}

	// Start the voice pipeline
	conn.session.startPipeline(conn)

	// Wait for connection to close
	<-conn.done
}

// sanitize removes newlines and carriage returns to prevent log injection.
func sanitize(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}
