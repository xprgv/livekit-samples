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
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
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

	// h264Track1080p, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "h264_high_res_video", func(tlsr *webrtc.TrackLocalStaticRTP) {})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	videoTrackSample, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "test_video_id")
	if err != nil {
		log.Fatal(err)
	}

	audioTrackSample, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "test_audio_id")
	if err != nil {
		log.Fatal(err)
	}

	// if publication, err := room.LocalParticipant.PublishTrack(h264Track1080p, &lksdk.TrackPublicationOptions{
	// 	Name:   "video",
	// 	Source: livekit.TrackSource_CAMERA,
	// }); err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	fmt.Println(publication.Name(), publication.IsSubscribed())
	// }

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

		h264SampleBuilder := samplebuilder.New(0, &codecs.H264Packet{}, 0)
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

			// fmt.Println(rtpPacket.Header)

			switch rtpPacket.PayloadType {
			case 96: // h264 packet
				h264SampleBuilder.Push(&rtpPacket)

				for sample := h264SampleBuilder.Pop(); sample != nil; sample = h264SampleBuilder.Pop() {
					// fmt.Println(sample.Data)
					// sample = &media.Sample{Data: []byte{0x5, 0xff, 0xff, 0xff, 0xff}, Duration: 30 * time.Millisecond}
					if err := videoTrackSample.WriteSample(*sample); err != nil && err != io.ErrClosedPipe {
						log.Fatal(err)
					} else {
						// fmt.Println("Write video sample")
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

		opusSampleBuilder := samplebuilder.New(960, &codecs.OpusPacket{}, 0)
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

			// fmt.Printf("%+v\n", rtpPacket.Header)

			switch rtpPacket.PayloadType {
			case 97: // opus packet
				opusSampleBuilder.Push(&rtpPacket)

				for sample := opusSampleBuilder.Pop(); sample != nil; sample = opusSampleBuilder.Pop() {
					// fmt.Println(sample.Timestamp)
					// fmt.Println(sample.Data)
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
