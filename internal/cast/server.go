package cast

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Server serves local files to cast devices over HTTP, on the address
// the device can reach. Each file gets an unguessable token, so nothing
// else on the network can browse the computer.
type Server struct {
	srv      *http.Server
	listener net.Listener
	host     string

	mu     sync.Mutex
	shares map[string]*share
}

type share struct {
	path      string
	convert   bool
	videoCopy bool
	audioCopy bool
	subtitles string // subtitle file, or "embedded"
	subIndex  int    // subtitle stream for "embedded"
}

// Source describes a local file to share.
type Source struct {
	Path string
	// Convert remuxes (and where needed transcodes) the file to
	// fragmented MP4 with FFmpeg as it streams.
	Convert bool
	// VideoCodec and AudioCodec are the file's codecs (FFmpeg names),
	// deciding what Convert can copy.
	VideoCodec, AudioCodec string
	// Subtitles is an external subtitle file, or "embedded" to extract
	// subtitle stream SubtitleIndex; "" for none.
	Subtitles     string
	SubtitleIndex int
}

// NewServer starts a server on the local address used to reach
// deviceAddr ("host:port" or a host).
func NewServer(deviceAddr string) (*Server, error) {
	ip, err := LocalIPFor(deviceAddr)
	if err != nil {
		return nil, err
	}
	l, err := net.Listen("tcp", net.JoinHostPort(ip, "0"))
	if err != nil {
		return nil, err
	}
	s := &Server{listener: l, host: l.Addr().String(), shares: map[string]*share{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/media/", s.serveMedia)
	mux.HandleFunc("/subs/", s.serveSubs)
	s.srv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go s.srv.Serve(l)
	return s, nil
}

// Close stops the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// Share makes src reachable and returns its URL, MIME type and the
// WebVTT subtitle URL (or "").
func (s *Server) Share(src Source) (media, contentType, subs string) {
	tok := token()
	sh := &share{path: src.Path, convert: src.Convert, subtitles: src.Subtitles, subIndex: src.SubtitleIndex}
	sh.videoCopy = castableVideo(src.VideoCodec)
	sh.audioCopy = castableAudio(src.AudioCodec)
	s.mu.Lock()
	s.shares[tok] = sh
	s.mu.Unlock()

	base := "http://" + s.host
	ext := strings.ToLower(filepath.Ext(src.Path))
	if src.Convert {
		media = base + "/media/" + tok + "/video.mp4"
		contentType = "video/mp4"
	} else {
		media = base + "/media/" + tok + "/media" + ext
		contentType = ContentType(src.Path)
	}
	if src.Subtitles != "" {
		subs = base + "/subs/" + tok + "/subtitles.vtt"
	}
	return media, contentType, subs
}

func (s *Server) lookup(r *http.Request, prefix string) (*share, bool) {
	rest := strings.TrimPrefix(r.URL.Path, prefix)
	tok, _, _ := strings.Cut(rest, "/")
	s.mu.Lock()
	defer s.mu.Unlock()
	sh, ok := s.shares[tok]
	return sh, ok
}

// LocalIPFor returns this computer's address on the network that reaches
// addr. No packets are sent: dialing UDP only picks the route.
func LocalIPFor(addr string) (string, error) {
	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	c, err := net.Dial("udp", net.JoinHostPort(host, "9"))
	if err != nil {
		return "", fmt.Errorf("cast: no route to %s: %w", host, err)
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).IP.String(), nil
}

func token() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
