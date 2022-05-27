package main

import (
	"fmt"
	"livekit-samples/internal/config"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

func main() {
	config := config.Config{
		Host:      "ws://localhost:7880",
		ApiKey:    "APInAy27RUmYUnV",
		ApiSecret: "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
		Identity:  "publisher-rtp-stream",
		RoomName:  "stark-tower",
		Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODc4MjIyNjMsImlzcyI6IkFQSW5BeTI3UlVtWVVuViIsImp0aSI6InRvbnlfc3RhcmsiLCJuYW1lIjoiVG9ueSBTdGFyayIsIm5iZiI6MTY1MTgyMjI2Mywic3ViIjoidG9ueV9zdGFyayIsInZpZGVvIjp7InJvb20iOiJzdGFyay10b3dlciIsInJvb21Kb2luIjp0cnVlfX0.XCuS0Rw73JI8vE6dBUD3WbYGFNz1zGzdUBaDmnuI9Aw",
	}

	fmt.Println("connecting to room")
	room, err := lksdk.ConnectToRoom(config.Host, lksdk.ConnectInfo{
		APIKey:              config.ApiKey,
		APISecret:           config.ApiSecret,
		RoomName:            config.RoomName,
		ParticipantIdentity: config.Identity,
		ParticipantName:     "rtp-stream-publisher",
	})
	if err != nil {
		log.Fatal(err)
	}

	h264Track1080p, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
	}, "video", "h264_high_res_video")
	if err != nil {
		log.Fatal(err)
	}

	opusTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
		Channels:  2,
	}, "audio", "audio")
	if err != nil {
		log.Fatal(err)
	}

	if publication, err := room.LocalParticipant.PublishTrack(h264Track1080p, &lksdk.TrackPublicationOptions{
		Name:   "video",
		Source: livekit.TrackSource_CAMERA,
	}); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(publication.Name(), publication.IsSubscribed())
	}

	if _, err := room.LocalParticipant.PublishTrack(opusTrack, &lksdk.TrackPublicationOptions{
		Name:   "audio",
		Source: livekit.TrackSource_MICROPHONE,
	}); err != nil {
		log.Fatal(err)
	}

	go func() {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20360})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		buf := make([]byte, 1500)
		for {
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}

			packet := rtp.Packet{}
			if err := packet.Unmarshal(buf[:n]); err != nil {
				log.Fatal(err)
			} else {
				// fmt.Println(
				// 	"payload_type:", packet.PayloadType,
				// 	"sequence_number:", packet.SequenceNumber,
				// 	"ssrc:", packet.SSRC,
				// 	"csrc:", packet.CSRC,
				// 	"timestamp:", packet.Timestamp,
				// )
				if err := h264Track1080p.WriteRTP(&packet); err != nil {
					log.Fatal(err)
				}
			}

			// if _, err := h264Track1080p.Write(buf[:n]); err != nil {
			// 	log.Fatal(err)
			// }
		}
	}()

	go func() {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 21360})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		buf := make([]byte, 1500)
		for {
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}
			packet := rtp.Packet{}
			if err := packet.Unmarshal(buf[:n]); err != nil {
				log.Fatal(err)
			} else {
				// fmt.Println(
				// 	"payload_type:", packet.PayloadType,
				// 	"sequence_number:", packet.SequenceNumber,
				// 	"ssrc:", packet.SSRC,
				// 	"csrc:", packet.CSRC,
				// 	"timestamp:", packet.Timestamp,
				// )
				if err := opusTrack.WriteRTP(&packet); err != nil {
					log.Fatal(err)
				}
			}
			// if _, err := opusTrack.Write(buf[:n]); err != nil {
			// 	log.Fatal(err)
			// }
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
