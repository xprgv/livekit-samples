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
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"

	jechsamplebuilder "github.com/jech/samplebuilder"
)

const (
	AUDIO_MAX_LATE = 32
	VIDEO_MAX_LATE = 256

	TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODc4MjIyNjMsImlzcyI6IkFQSW5BeTI3UlVtWVVuViIsImp0aSI6InRvbnlfc3RhcmsiLCJuYW1lIjoiVG9ueSBTdGFyayIsIm5iZiI6MTY1MTgyMjI2Mywic3ViIjoidG9ueV9zdGFyayIsInZpZGVvIjp7InJvb20iOiJzdGFyay10b3dlciIsInJvb21Kb2luIjp0cnVlfX0.XCuS0Rw73JI8vE6dBUD3WbYGFNz1zGzdUBaDmnuI9Aw"
)

func main() {
	fmt.Println("connecting to room")
	room, err := lksdk.ConnectToRoom(
		"ws://localhost:7880",
		lksdk.ConnectInfo{
			APIKey:              "APInAy27RUmYUnV",
			APISecret:           "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
			RoomName:            "stark-tower",
			ParticipantIdentity: "publisher-rtp-stream-identity",
			ParticipantName:     "rtp-stream-publisher",
		})
	if err != nil {
		log.Fatal(err)
	}

	go startH264(room)
	// go startVP8(room)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()

}

func startH264(room *lksdk.Room) {
	videoTrackH264Sample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:    webrtc.MimeTypeH264,
		ClockRate:   90000,
		SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := room.LocalParticipant.PublishTrack(videoTrackH264Sample, &lksdk.TrackPublicationOptions{
		Name:   "h264_video",
		Source: livekit.TrackSource_CAMERA,
	}); err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5500})
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	depacketizer := &codecs.H264Packet{}
	h264SampleBuffer := jechsamplebuilder.New(500, depacketizer, videoTrackH264Sample.Codec().ClockRate)

	buf := make([]byte, 1500)
	for {
		rtpPacket := rtp.Packet{}

		n, _, err := listener.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
		}
		if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
			log.Fatal(err)
		}

		switch rtpPacket.PayloadType {
		case 96: // h264 media
			h264SampleBuffer.Push(rtpPacket.Clone())

			for {
				sample := h264SampleBuffer.Pop()
				if sample == nil {
					break
				}

				if err := videoTrackH264Sample.WriteSample(*sample, nil); err != nil {
					log.Fatal(err)
				}
			}
		default:
			fmt.Println("Another packet")
		}
	}
}
