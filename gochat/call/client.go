package call

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"gochat/auth"
	"gochat/chat"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	// === PortAudio ===
	if err := portaudio.Initialize(); err != nil {
		log.Fatal("PortAudio init error:", err)
	}
	defer portaudio.Terminate()

	// === WebRTC peer ===
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
			// For maximum reliability set up TURN server here – for now this would be overkill and not necessary
			// {
			//     URLs:       []string{"turn:your.turn.server:3478"},
			//     Username:   "user",
			//     Credential: "pass",
			// },
		},
	})
	if err != nil {
		log.Fatal(err)
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

	// === Microphone capture ===
	in := make([]int16, 960)
	micStream, err := portaudio.OpenDefaultStream(1, 0, 48000, len(in), &in)
	if err != nil {
		log.Fatal("Mic stream error:", err)
	}
	defer micStream.Close()
	micStream.Start()

	encoder, err := opus.NewEncoder(48000, 1, opus.AppVoIP)
	if err != nil {
		log.Fatal("Opus encoder error:", err)
	}

	go func() {
		for {
			if err := micStream.Read(); err != nil {
				log.Println("Mic read error:", err)
				continue
			}

			encoded := make([]byte, 4000)
			n, err := encoder.Encode(in, encoded)
			if err != nil {
				log.Println("Opus encode error:", err)
				continue
			}

			err = audioTrack.WriteSample(media.Sample{
				Data:     encoded[:n],
				Duration: 20 * time.Millisecond,
			})
			if err != nil {
				log.Println("WriteSample error:", err)
			}
		}
	}()

	// === Handle ICE ===
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
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
		_ = conn.WriteJSON(msg)
	})

	// === Handle incoming audio ===
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Println("Incoming track:", track.Kind(), track.Codec().MimeType)

		decoder, err := opus.NewDecoder(48000, 1)
		if err != nil {
			log.Fatal("Opus decoder error:", err)
		}

		out := make([]int16, 960)
		stream, err := portaudio.OpenDefaultStream(0, 1, 48000, len(out), &out)
		if err != nil {
			log.Fatal("Output stream error:", err)
		}
		stream.Start()

		go func() {
			defer stream.Close()
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
				if n > 0 {
					if err := stream.Write(); err != nil {
						log.Println("PortAudio write error:", err)
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
				return
			}

			switch msg.Type {
			case "offer":
				//log.Println("Received offer")
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
				err = conn.WriteJSON(answerMsg)
				if err != nil {
					log.Println("Error sending answer", err)
					continue
				}

			case "answer":
				//log.Println("Received answer")
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
				err = pc.SetRemoteDescription(answer)
				if err != nil {
					log.Println("Error setting remote description", err)
					continue
				}

			case "candidate":
				//log.Println("Received ICE candidate")
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
				err = pc.AddICECandidate(candidate)
				if err != nil {
					log.Println("Error adding ICE candidate:", err)
					continue
				}
			}
		}
	}()

	// === Start call ===
	log.Println("Press Enter to send offer...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = pc.SetLocalDescription(offer)
	if err != nil {
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
	err = conn.WriteJSON(msg)
	if err != nil {
		log.Fatal("Error sending offer:", err)
	}

	select {} // keep alive
}
