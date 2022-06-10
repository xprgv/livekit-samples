package main

import (
	"fmt"
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

const (
	TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODc4MjIyNjMsImlzcyI6IkFQSW5BeTI3UlVtWVVuViIsImp0aSI6InRvbnlfc3RhcmsiLCJuYW1lIjoiVG9ueSBTdGFyayIsIm5iZiI6MTY1MTgyMjI2Mywic3ViIjoidG9ueV9zdGFyayIsInZpZGVvIjp7InJvb20iOiJzdGFyay10b3dlciIsInJvb21Kb2luIjp0cnVlfX0.XCuS0Rw73JI8vE6dBUD3WbYGFNz1zGzdUBaDmnuI9Aw"
)

func main() {
	room, err := lksdk.ConnectToRoom(
		"ws://localhost:7880", lksdk.ConnectInfo{
			APIKey:              "APInAy27RUmYUnV",
			APISecret:           "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
			RoomName:            "stark-tower",
			ParticipantIdentity: "publisher-rtp-stream",
			ParticipantName:     "rtp-stream-publisher",
		})
	if err != nil {
		log.Fatal(err)
	}

	go startOpus(room)
	go startH264(room)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}

func startH264(room *lksdk.Room) {
	h264Track1080p, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
	}, "video", "h264_high_res_video")
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

	go func() {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20360})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		buf := make([]byte, 1500)
		rtpPacket := rtp.Packet{}
		for {
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}

			if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
				log.Fatal(err)
			}

			switch rtpPacket.PayloadType {
			case 96:
				// fmt.Println(packet.SequenceNumber)
				if err := h264Track1080p.WriteRTP(rtpPacket.Clone()); err != nil {
					log.Fatal(err)
				}
			default:
				// log in debug mode packet with another type
				log.Println("Another payload type in h264 rtp stream")
				// log.Printf("%+v\n", rtpPacket)
			}
		}
	}()
}

func startOpus(room *lksdk.Room) {
	opusTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
		Channels:  2,
	}, "audio", "audio")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := room.LocalParticipant.PublishTrack(opusTrack, &lksdk.TrackPublicationOptions{
		Name:   "audio",
		Source: livekit.TrackSource_MICROPHONE,
	}); err != nil {
		log.Fatal(err)
	}

	go func() {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 21360})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		buf := make([]byte, 1500)
		rtpPacket := rtp.Packet{}
		for {
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}

			if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
				log.Fatal(err)
			}

			switch rtpPacket.PayloadType {
			case 97:
				if err := opusTrack.WriteRTP(rtpPacket.Clone()); err != nil {
					log.Fatal(err)
				}
			default:
				log.Println("another payload type in opus rtp stream")
			}
		}
	}()
}
