package call

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gochat/connections"
	"log"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
	"github.com/hraban/opus"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type CallSession struct {
	User      string
	Peer      *webrtc.PeerConnection
	writeJSON func(v interface{}) error

	// candidate buffering
	remoteCands   []webrtc.ICECandidateInit
	remoteCandsMu sync.Mutex
}

// --- Constructor ---
func NewCallSession(user string, conn *websocket.Conn, recipient string) (*CallSession, error) {
	var writeMutex sync.Mutex
	writeJSON := func(v interface{}) error {
		writeMutex.Lock()
		defer writeMutex.Unlock()
		return conn.WriteJSON(v)
	}

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	})
	if err != nil {
		return nil, err
	}

	// Debug logging for connection state
	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
	    log.Println("PeerConnection state:", state)
	})

	sess := &CallSession{
		User:      user,
		Peer:      pc,
		writeJSON: writeJSON,
	}

	// === Setup Peer ===
	sess.setupICE(recipient)
	sess.setupAudio()

	return sess, nil
}

func Call(username string, recipient string) {
	conn, err := connections.GetConnection(username)
	if err != nil {
		log.Fatalf("Error establishing connection: %v", err)
		return
	}
	defer conn.Close()

	sess, err := NewCallSession(username, conn, recipient) // use real username from token
	if err != nil {
		log.Fatalf("Error creating call session: %v", err)
	}

	if err := sess.StartCall(recipient); err != nil {
		log.Fatalf("Error starting call: %v", err)
	}

	select {}
}

// --- Caller flow: create & send offer ---
func (s *CallSession) StartCall(recipient string) error {
	offer, err := s.Peer.CreateOffer(nil)
	if err != nil {
		return err
	}
	if err := s.Peer.SetLocalDescription(offer); err != nil {
		return err
	}

	offerJSON, _ := json.Marshal(offer)
	msg := Message{
		Username:  s.User,
		Type:      "offer",
		Recipient: recipient,
		Payload:   base64.StdEncoding.EncodeToString(offerJSON),
		Sent:      time.Now(),
	}
	fmt.Println("offer sent!")
	return s.writeJSON(msg)
}

// --- Callee flow: nothing to send until offer arrives ---
func (s *CallSession) AcceptCall() {
	// just wait for incoming "offer" handled by HandleSignalMessage
}

// --- Handle signaling ---
func (s *CallSession) HandleSignalMessage(msg Message) {
	flushRemoteCands := func() {
		s.remoteCandsMu.Lock()
		cands := s.remoteCands
		s.remoteCands = nil
		s.remoteCandsMu.Unlock()
		for _, c := range cands {
			if err := s.Peer.AddICECandidate(c); err != nil {
				log.Println("AddICECandidate (flushed) error:", err)
			}
		}
	}

	switch msg.Type {
	case "offer":
		fmt.Println("Offer received...")
		var offer webrtc.SessionDescription
		decoded, _ := base64.StdEncoding.DecodeString(msg.Payload)
		json.Unmarshal(decoded, &offer)

		if err := s.Peer.SetRemoteDescription(offer); err != nil {
			log.Println("SetRemoteDescription offer error:", err)
			return
		}
		flushRemoteCands()

		answer, err := s.Peer.CreateAnswer(nil)
		if err != nil {
			log.Println("CreateAnswer error:", err)
			return
		}
		if err := s.Peer.SetLocalDescription(answer); err != nil {
			log.Println("SetLocalDescription answer error:", err)
			return
		}

		answerJSON, err := json.Marshal(answer)
		if err != nil {
			log.Println("Error marshalling answer: %v", err)
		}
		answerMsg := Message{
			Username:  s.User,
			Type:      "answer",
			Recipient: msg.Username,
			Payload:   base64.StdEncoding.EncodeToString(answerJSON),
			Sent:      time.Now(),
		}
		s.writeJSON(answerMsg)
		fmt.Println("Answer sent!")

	case "answer":
		fmt.Println("Answer received...")
		var answer webrtc.SessionDescription
		decoded, _ := base64.StdEncoding.DecodeString(msg.Payload)
		json.Unmarshal(decoded, &answer)

		if err := s.Peer.SetRemoteDescription(answer); err != nil {
			log.Println("SetRemoteDescription answer error:", err)
			return
		}
		flushRemoteCands()

	case "candidate":
		fmt.Println("Candidate received...")
		var cand webrtc.ICECandidateInit
		decoded, _ := base64.StdEncoding.DecodeString(msg.Payload)
		json.Unmarshal(decoded, &cand)

		if s.Peer.RemoteDescription() == nil {
			s.remoteCandsMu.Lock()
			s.remoteCands = append(s.remoteCands, cand)
			s.remoteCandsMu.Unlock()
		} else {
			if err := s.Peer.AddICECandidate(cand); err != nil {
				log.Println("AddICECandidate error:", err)
			}
		}
	}
}

// --- Setup ICE ---
func (s *CallSession) setupICE(recipient string) {
	s.Peer.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		candidateJSON, _ := json.Marshal(c.ToJSON())
		msg := Message{
			Username:  s.User,
			Type:      "candidate",
			Recipient: recipient,
			Payload:   base64.StdEncoding.EncodeToString(candidateJSON),
			Sent:      time.Now(),
		}
		fmt.Println("Candidate sent!")
		if err := s.writeJSON(msg); err != nil {
			log.Println("Error sending candidate:", err)
		}
	})
}

// --- Setup audio input/output ---
func (s *CallSession) setupAudio() {
	if err := portaudio.Initialize(); err != nil {
		log.Fatal("PortAudio init error:", err)
	}

	// Track for microphone
	audioTrack, _ := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion",
	)
	s.Peer.AddTrack(audioTrack)

	// Microphone capture
	in := make([]int16, 960)
	micStream, _ := portaudio.OpenDefaultStream(1, 0, 48000, len(in), &in)
	encoder, _ := opus.NewEncoder(48000, 1, opus.AppVoIP)

	micStream.Start()
	go func() {
		defer micStream.Close()
		for {
			if err := micStream.Read(); err != nil {
				continue
			}
			encoded := make([]byte, 4000)
			n, err := encoder.Encode(in, encoded)
			if err == nil && n > 0 {
				audioTrack.WriteSample(media.Sample{Data: encoded[:n], Duration: 20 * time.Millisecond})
			}
		}
	}()

	// Playback (remote audio)
	s.Peer.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		decoder, err := opus.NewDecoder(48000, 1)
		if err != nil {
			log.Println("Error decoding: %v", err)
		}
		out := make([]int16, 960)
		stream, err := portaudio.OpenDefaultStream(0, 1, 48000, len(out), &out)
		if err != nil {
			log.Println("Error opening default stream: %v", err)
		}
		stream.Start()

		go func() {
			defer stream.Close()
			for {
				pkt, _, err := track.ReadRTP()
				if err != nil {
					return
				}
				n, err := decoder.Decode(pkt.Payload, out)
				if err == nil && n > 0 {
					stream.Write()
				}
			}
		}()
	})
}
