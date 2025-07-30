package call

import (
	"bufio"
	"encoding/json"
	"gochat/auth"
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

func Call(user string) {
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
	conn, _, err := websocket.DefaultDialer.Dial("ws://raspberrypi.fritz.box:8080/signal/call", header)
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
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
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
		msg, err := json.Marshal(map[string]interface{}{
			"type":      "candidate",
			"candidate": c.ToJSON(),
		})
		if err != nil {
			log.Fatal(err)
		}
		conn.WriteMessage(websocket.TextMessage, msg)
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
		//defer stream.Close()
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
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket read error:", err)
				return
			}

			var data map[string]interface{}
			if err := json.Unmarshal(msg, &data); err != nil {
				continue
			}

			switch data["type"] {
			case "offer":
				log.Println("Received offer")
				b, err := json.Marshal(data)
				if err != nil {
					log.Fatal(err)
				}
				var offer webrtc.SessionDescription
				err = json.Unmarshal(b, &offer)
				if err != nil {
					log.Fatal(err)
				}
				err = pc.SetRemoteDescription(offer)
				if err != nil {
					log.Fatal(err)
				}
				answer, err := pc.CreateAnswer(nil)
				if err != nil {
					log.Fatal(err)
				}
				err = pc.SetLocalDescription(answer)
				if err != nil {
					log.Fatal(err)
				}
				res, err := json.Marshal(answer)
				if err != nil {
					log.Fatal(err)
				}
				conn.WriteMessage(websocket.TextMessage, res)

			case "answer":
				log.Println("Received answer")
				b, _ := json.Marshal(data)
				var answer webrtc.SessionDescription
				err = json.Unmarshal(b, &answer)
				if err != nil {
					log.Fatal(err)
				}
				err = pc.SetRemoteDescription(answer)
				if err != nil {
					log.Fatal(err)
				}

			case "candidate":
				log.Println("Received ICE candidate")
				candRaw := data["candidate"]
				candJSON, err := json.Marshal(candRaw)
				if err != nil {
					log.Fatal(err)
				}
				var candidate webrtc.ICECandidateInit
				err = json.Unmarshal(candJSON, &candidate)
				if err != nil {
					log.Fatal(err)
				}
				err = pc.AddICECandidate(candidate)
				if err != nil {
					log.Fatal(err)
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
	_ = pc.SetLocalDescription(offer)
	offerJSON, err := json.Marshal(offer)
	if err != nil {
		log.Fatal(err)
	}
	conn.WriteMessage(websocket.TextMessage, offerJSON)

	select {} // keep alive
}
