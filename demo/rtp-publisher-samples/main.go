package main

import (
	"fmt"
	"io"

	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"livekit-samples/pkg/codecs"
	"livekit-samples/pkg/lksdksamplebuilder"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"

	// lksdksamplebuilder "github.com/livekit/server-sdk-go/pkg/samplebuilder"
	// pionsamplebuilder "github.com/pion/webrtc/v3/pkg/media/samplebuilder"
	jechsamplebuilder "github.com/jech/samplebuilder"
	// _ "net/http/pprof"
)

const (
	AUDIO_MAX_LATE = 32
	VIDEO_MAX_LATE = 256

	TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODc4MjIyNjMsImlzcyI6IkFQSW5BeTI3UlVtWVVuViIsImp0aSI6InRvbnlfc3RhcmsiLCJuYW1lIjoiVG9ueSBTdGFyayIsIm5iZiI6MTY1MTgyMjI2Mywic3ViIjoidG9ueV9zdGFyayIsInZpZGVvIjp7InJvb20iOiJzdGFyay10b3dlciIsInJvb21Kb2luIjp0cnVlfX0.XCuS0Rw73JI8vE6dBUD3WbYGFNz1zGzdUBaDmnuI9Aw"
)

func main() {
	// go http.ListenAndServe("127.0.0.1:8080", nil)

	// config := config.Config{
	// 	Host:      "ws://localhost:7880",
	// 	ApiKey:    "APInAy27RUmYUnV",
	// 	ApiSecret: "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
	// 	Identity:  "publisher-rtp-stream",
	// 	RoomName:  "stark-tower",
	// 	Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODc4MjIyNjMsImlzcyI6IkFQSW5BeTI3UlVtWVVuViIsImp0aSI6InRvbnlfc3RhcmsiLCJuYW1lIjoiVG9ueSBTdGFyayIsIm5iZiI6MTY1MTgyMjI2Mywic3ViIjoidG9ueV9zdGFyayIsInZpZGVvIjp7InJvb20iOiJzdGFyay10b3dlciIsInJvb21Kb2luIjp0cnVlfX0.XCuS0Rw73JI8vE6dBUD3WbYGFNz1zGzdUBaDmnuI9Aw",
	// }

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

	go startOpus(room)
	// go startH264(room)
	go startVP8(room)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()

}

func startOpus(room *lksdk.Room) {
	audioTrackSample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
		Channels:  2,
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(audioTrackSample.Codec(), audioTrackSample.IsBound())
	}

	if _, err := room.LocalParticipant.PublishTrack(audioTrackSample, &lksdk.TrackPublicationOptions{
		Name:   "opus_audio",
		Source: livekit.TrackSource_MICROPHONE,
	}); err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 21360})
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	opusSampleBuilder := lksdksamplebuilder.New(AUDIO_MAX_LATE, &codecs.OpusPacket{}, audioTrackSample.Codec().ClockRate)

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
		case 97: // opus media
			opusSampleBuilder.Push(rtpPacket.Clone())
			for sample := opusSampleBuilder.Pop(); sample != nil; sample = opusSampleBuilder.Pop() {
				if err := audioTrackSample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
					log.Fatal(err)
				}
			}
		}
	}
}

func startH264(room *lksdk.Room) {
	videoTrackH264Sample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
		Channels:  0,
		// SDPFmtpLine: "125, level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
		SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
		// RTCPFeedback: []webrtc.RTCPFeedback{
		// 	{Type: webrtc.TypeRTCPFBNACK},
		// 	{Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"},
		// },
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(videoTrackH264Sample.Codec())

		// if err := videoTrackH264Sample.StartWrite(NewH264SampleProvider(), func() {}); err != nil {
		// 	log.Fatal(err)
		// }
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

	// depacketizer := &codecs.H264Packet{}
	depacketizer := &codecs.H264Packet{}

	h264SampleBuffer := jechsamplebuilder.New(500, depacketizer, videoTrackH264Sample.Codec().ClockRate)

	// go func() {
	// 	timer := time.NewTicker(10 * time.Millisecond)

	// }()

	buf := make([]byte, 1500)

	for {
		rtpPacket := rtp.Packet{}

		// n, _, err := listener.ReadFromUDP(buf)
		n, _, err := listener.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
		}
		if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
			log.Fatal(err)
		}

		// fmt.Println(len(rtpPacket.Payload))

		// if len(rtpPacket.Payload) <= 2 {
		// 	continue
		// }

		switch rtpPacket.PayloadType {
		case 96: // h264 media
			h264SampleBuffer.Push(rtpPacket.Clone())

			for {
				sample, ts := h264SampleBuffer.PopWithTimestamp()
				if sample == nil {
					// fmt.Println("nil sample")
					break
				}
				sample.PacketTimestamp = ts

				if err := videoTrackH264Sample.WriteSample(*sample, nil); err != nil {
					log.Fatal(err)
				}
			}
		default:
			fmt.Println("Another packet")
		}
	}

	// select {}
}

func startVP8(room *lksdk.Room) {
	videoTrackVP8Sample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeVP8,
		ClockRate: 90000,
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(videoTrackVP8Sample.Codec())
	}

	if publication, err := room.LocalParticipant.PublishTrack(videoTrackVP8Sample, &lksdk.TrackPublicationOptions{
		Name:   "video_vp8_test",
		Source: livekit.TrackSource_CAMERA,
	}); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(publication.MimeType())
	}

	go func() {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5500})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		vp8SampleBuilder := lksdksamplebuilder.New(100, &codecs.VP8Packet{}, videoTrackVP8Sample.Codec().ClockRate)
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
			// fmt.Printf("%+v\n", rtpPacket)

			vp8SampleBuilder.Push(rtpPacket.Clone())

			for {
				sample := vp8SampleBuilder.Pop()
				if sample == nil {
					break
				}

				if err := videoTrackVP8Sample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
					log.Fatal(err)
				}
			}
		}
	}()
}
