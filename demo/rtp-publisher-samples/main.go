package main

import (
	"fmt"
	"io"
	"livekit-samples/internal/config"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"

	"github.com/jech/samplebuilder"
	galenesamplebuilder "github.com/jech/samplebuilder"
)

const (
	H264_FRAME_DURATION = time.Millisecond * 33
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

	videoTrackSample, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
		MimeType: webrtc.MimeTypeH264,
		// ClockRate: 48000,
	}, "video", "test_video_id")
	if err != nil {
		log.Fatal(err)
	}

	audioTrackSample, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
		MimeType:    webrtc.MimeTypeOpus,
		ClockRate:   48000,
		Channels:    2,
		SDPFmtpLine: "flt",
	}, "audio", "test_audio_id")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := room.LocalParticipant.PublishTrack(videoTrackSample, &lksdk.TrackPublicationOptions{
		Name:   "video_test",
		Source: livekit.TrackSource_CAMERA,
	}); err != nil {
		log.Fatal(err)
	}

	if _, err := room.LocalParticipant.PublishTrack(audioTrackSample, &lksdk.TrackPublicationOptions{
		Name:   "audio_test",
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

		h264SampleBuilder := samplebuilder.New(0, &codecs.H264Packet{}, 48000)
		buf := make([]byte, 1500)

		for {
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}
			rtpPacket := rtp.Packet{}
			if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
				log.Fatal(err)
			}

			switch rtpPacket.PayloadType {
			case 96: // h264 media
				h264SampleBuilder.Push(&rtpPacket)
				for sample := h264SampleBuilder.Pop(); sample != nil; sample = h264SampleBuilder.Pop() {
					if err := videoTrackSample.WriteSample(*sample); err != nil && err != io.ErrClosedPipe {
						log.Fatal(err)
					} else {
						fmt.Println("Write video sample")
					}
				}
			default:
			}
		}
	}()

	go func() {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 21360})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		// opusSampleBuilder := samplebuilder.New(100, &codecs.OpusPacket{}, 48000)
		opusSampleBuilder := galenesamplebuilder.New(5, &codecs.OpusPacket{}, 48000)
		buf := make([]byte, 1500)

		for {
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}
			rtpPacket := rtp.Packet{}
			if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
				log.Fatal(err)
			}

			switch rtpPacket.PayloadType {
			case 97: // opus media
				opusSampleBuilder.Push(&rtpPacket)
				for sample := opusSampleBuilder.Pop(); sample != nil; sample = opusSampleBuilder.Pop() {
					if err := audioTrackSample.WriteSample(*sample); err != nil && err != io.ErrClosedPipe {
						log.Fatal(err)
					} else {
						// fmt.Println("Write audio sample")
					}
				}
			default:
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
