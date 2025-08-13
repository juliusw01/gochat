package call

import (
	"encoding/base64"
	"encoding/json"
	"gochat/auth"
	"gochat/chat"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
	"github.com/hraban/opus"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

func Call(user string, recipient string) {
	// === AUTH ===
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	tokenPath := filepath.Join(homeDir, ".gochat", user, "authToken.txt")
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatalf("Read token error: %v", err)
	}
	token := string(tokenBytes)
	if _, err := auth.ExtractUserFromToken(token); err != nil {
		log.Fatalf("Token invalid: %v", err)
	}

	// === WEBSOCKET ===
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	conn, _, err := websocket.DefaultDialer.Dial("ws://raspberrypi.fritz.box:8080/ws", header)
	if err != nil {
		log.Fatal("WebSocket dial error:", err)
	}
	defer conn.Close()

	// protect concurrent writes on websocket
	var writeMutex sync.Mutex
	writeJSON := func(v interface{}) error {
		writeMutex.Lock()
		defer writeMutex.Unlock()
		return conn.WriteJSON(v)
	}

	// === PortAudio ===
	if err := portaudio.Initialize(); err != nil {
		log.Fatal("PortAudio init error:", err)
	}
	defer portaudio.Terminate()

	// === WebRTC peer ===
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// cleanup helper
	cleanupOnce := sync.Once{}
	cleanup := func() {
		cleanupOnce.Do(func() {
			log.Println("Cleaning up call resources")
			// Close PeerConnection (this will also close tracks)
			if pc != nil {
				_ = pc.Close()
			}
			// Close websocket (deferred in caller too)
			// conn.Close()
			// PortAudio terminated by defer
		})
	}

	// buffer incoming remote ICE candidates until remote description set
	var remoteCandsMu sync.Mutex
	var remoteCands []webrtc.ICECandidateInit
	flushRemoteCands := func() {
		remoteCandsMu.Lock()
		cands := remoteCands
		remoteCands = nil
		remoteCandsMu.Unlock()
		for _, c := range cands {
			if err := pc.AddICECandidate(c); err != nil {
				log.Println("AddICECandidate (flushed) error:", err)
			}
		}
	}

	// === Track (Opus audio) ===
	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion",
	)
	if err != nil {
		log.Fatal(err)
	}
	_, err = pc.AddTrack(audioTrack)
	if err != nil {
		log.Fatal(err)
	}

	// === Microphone capture & Opus encode ===
	in := make([]int16, 960)
	micStream, err := portaudio.OpenDefaultStream(1, 0, 48000, len(in), &in)
	if err != nil {
		log.Fatal("Mic stream error:", err)
	}
	// ensure mic stream closed on cleanup
	defer func() {
		_ = micStream.Close()
	}()
	if err := micStream.Start(); err != nil {
		log.Fatal("Mic stream start error:", err)
	}

	encoder, err := opus.NewEncoder(48000, 1, opus.AppVoIP)
	if err != nil {
		log.Fatal("Opus encoder error:", err)
	}

	// produce encoded Opus frames and write to the local track
	go func() {
		defer func() {
			_ = micStream.Stop()
		}()

		for {
			if err := micStream.Read(); err != nil {
				log.Println("Mic read error:", err)
				// small sleep to avoid hot loop on persistent error
				time.Sleep(10 * time.Millisecond)
				continue
			}

			encoded := make([]byte, 4000)
			n, err := encoder.Encode(in, encoded)
			if err != nil {
				log.Println("Opus encode error:", err)
				continue
			}

			if n == 0 {
				continue
			}

			if err := audioTrack.WriteSample(media.Sample{
				Data:     encoded[:n],
				Duration: 20 * time.Millisecond,
			}); err != nil {
				// WriteSample can return ErrNotConnected if PeerConnection not connected yet — log and continue
				log.Println("WriteSample error:", err)
			}
		}
	}()

	// === Handle ICE ===
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		// marshal candidate object
		candidateJSON, err := json.Marshal(c.ToJSON())
		if err != nil {
			log.Println("ICE candidate marshal error:", err)
			return
		}
		msg := chat.Message{
			Username:  user,
			Type:      "candidate",
			Recipient: recipient,
			Payload:   base64.StdEncoding.EncodeToString(candidateJSON),
			Sent:      time.Now(),
		}
		if err := writeJSON(msg); err != nil {
			log.Println("Error sending candidate:", err)
		}
	})

	// flush buffered remote candidates when remote description arrives
	pc.OnSignalingStateChange(func(s webrtc.SignalingState) {
		// when we have remote description set, flush
		if pc.RemoteDescription() != nil {
			flushRemoteCands()
		}
	})

	// cleanup on connection state changes
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Println("PeerConnection state:", s.String())
		if s == webrtc.PeerConnectionStateFailed || s == webrtc.PeerConnectionStateDisconnected || s == webrtc.PeerConnectionStateClosed {
			// cleanup resources
			cleanup()
		}
	})

	// === Handle incoming audio ===
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Println("Incoming track:", track.Kind(), track.Codec().MimeType)

		decoder, err := opus.NewDecoder(48000, 1)
		if err != nil {
			log.Println("Opus decoder error:", err)
			return
		}

		out := make([]int16, 960) // enough for 20ms @ 48kHz mono
		stream, err := portaudio.OpenDefaultStream(0, 1, 48000, len(out), &out)
		if err != nil {
			log.Println("Output stream error:", err)
			return
		}
		if err := stream.Start(); err != nil {
			_ = stream.Close()
			log.Println("Failed to start output stream:", err)
			return
		}

		// read loop for track -> decode -> play
		go func() {
			defer func() {
				_ = stream.Stop()
				_ = stream.Close()
				log.Println("Output stream closed for track")
			}()

			for {
				pkt, _, err := track.ReadRTP()
				if err != nil {
					log.Println("RTP read error:", err)
					return
				}

				n, err := decoder.Decode(pkt.Payload, out)
				if err != nil {
					log.Println("Opus decode error:", err)
					continue
				}

				// decoder.Decode writes PCM samples into 'out' and returns samples-per-channel count
				// PortAudio writes the buffer pointed to by 'out' (as opened). We expect 'out' to contain valid PCM.
				if n > 0 {
					// PortAudio stream.Write uses the buffer we supplied on Open; since out is reused, this is safe
					if err := stream.Write(); err != nil {
						log.Println("PortAudio write error:", err)
						return
					}
				}
			}
		}()
	})

	// === Handle signaling ===
	go func() {
		for {
			var msg chat.Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("WebSocket read error:", err)
				cleanup()
				return
			}

			switch msg.Type {
			case "offer":
				log.Println("Received offer from", msg.Username)
				decoded, err := base64.StdEncoding.DecodeString(msg.Payload)
				if err != nil {
					log.Println("Offer decode error:", err)
					continue
				}
				var offer webrtc.SessionDescription
				if err := json.Unmarshal(decoded, &offer); err != nil {
					log.Println("Offer unmarshal error:", err)
					continue
				}
				if err := pc.SetRemoteDescription(offer); err != nil {
					log.Println("Error setting remote description:", err)
					continue
				}
				// flush candidates (if any) now that remote description is set
				flushRemoteCands()
				answer, err := pc.CreateAnswer(nil)
				if err != nil {
					log.Println("Error creating answer:", err)
					continue
				}
				if err := pc.SetLocalDescription(answer); err != nil {
					log.Println("Error setting local description:", err)
					continue
				}
				answerJSON, err := json.Marshal(answer)
				if err != nil {
					log.Println("Error creating answer JSON", err)
					continue
				}
				answerMsg := chat.Message{
					Username:  user,
					Type:      "answer",
					Recipient: msg.Username,
					Payload:   base64.StdEncoding.EncodeToString(answerJSON),
					Sent:      time.Now(),
				}
				if err := writeJSON(answerMsg); err != nil {
					log.Println("Error sending answer", err)
					continue
				}

			case "answer":
				log.Println("Received answer from", msg.Username)
				decoded, err := base64.StdEncoding.DecodeString(msg.Payload)
				if err != nil {
					log.Println("Error decoding answer:", err)
					continue
				}
				var answer webrtc.SessionDescription
				if err := json.Unmarshal(decoded, &answer); err != nil {
					log.Println("Error unmarshalling answer JSON:", err)
					continue
				}
				if err := pc.SetRemoteDescription(answer); err != nil {
					log.Println("Error setting remote description:", err)
					continue
				}
				// flush any buffered remote candidates after remote description set
				flushRemoteCands()

			case "candidate":
				decoded, err := base64.StdEncoding.DecodeString(msg.Payload)
				if err != nil {
					log.Println("Error decoding candidate:", err)
					continue
				}
				var candidate webrtc.ICECandidateInit
				if err := json.Unmarshal(decoded, &candidate); err != nil {
					log.Println("Error unmarshalling candidate:", err)
					continue
				}
				// If remote description is not set yet, buffer the candidate
				if pc.RemoteDescription() == nil {
					remoteCandsMu.Lock()
					remoteCands = append(remoteCands, candidate)
					remoteCandsMu.Unlock()
				} else {
					if err := pc.AddICECandidate(candidate); err != nil {
						log.Println("Error adding ICE candidate:", err)
					}
				}
			}
		}
	}()

	// === Start call ===
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		log.Fatal("Error setting local description:", err)
	}
	offerJSON, err := json.Marshal(offer)
	if err != nil {
		log.Fatal("Error marshalling offer:", err)
	}
	msg := chat.Message{
		Username:  user,
		Type:      "offer",
		Recipient: recipient,
		Payload:   base64.StdEncoding.EncodeToString(offerJSON),
		Sent:      time.Now(),
	}
	if err := writeJSON(msg); err != nil {
		log.Fatal("Error sending offer:", err)
	}

	// block until cleanup triggers (connection state change or errors)
	select {}
}
