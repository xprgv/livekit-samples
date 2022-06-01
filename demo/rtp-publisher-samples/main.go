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

	// "github.com/jech/samplebuilder"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/livekit/server-sdk-go/pkg/samplebuilder"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
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

	go startOpus(room)
	go startH264(room)
	// go startVP8(room)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()

}

func startOpus(room *lksdk.Room) {
	audioTrackSample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:    webrtc.MimeTypeOpus,
		ClockRate:   48000,
		Channels:    2,
		SDPFmtpLine: "flt",
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(audioTrackSample.Codec())
	}

	if _, err := room.LocalParticipant.PublishTrack(audioTrackSample, &lksdk.TrackPublicationOptions{
		Name:   "audio_test",
		Source: livekit.TrackSource_MICROPHONE,
	}); err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 21360})
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	opusSampleBuilder := samplebuilder.New(0, &codecs.OpusPacket{}, audioTrackSample.Codec().ClockRate, samplebuilder.WithPacketDroppedHandler(func() {
		fmt.Println("opus packet dropped")
	}))

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
		switch rtpPacket.PayloadType {
		case 97: // opus media
			opusSampleBuilder.Push(rtpPacket.Clone())
			for sample := opusSampleBuilder.Pop(); sample != nil; sample = opusSampleBuilder.Pop() {
				if err := audioTrackSample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
					log.Fatal(err)
				} else {
					// fmt.Println("Write audio sample")
				}
			}
		}
	}

}

func startH264(room *lksdk.Room) {
	videoTrackH264Sample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
		// SDPFmtpLine: "packetization-mode=0;annexb=yes;flt",
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(videoTrackH264Sample.Codec())
	}

	if _, err := room.LocalParticipant.PublishTrack(videoTrackH264Sample, &lksdk.TrackPublicationOptions{
		Name:   "video_test",
		Source: livekit.TrackSource_CAMERA,
	}); err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20360})
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	h264SampleBuilder := samplebuilder.New(100, &codecs.H264Packet{}, videoTrackH264Sample.Codec().ClockRate, samplebuilder.WithPacketDroppedHandler(func() {
		fmt.Println("h264 packet droped")
	}))

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
		// fmt.Printf("%+v\n", rtpPacket.SequenceNumber)

		switch rtpPacket.PayloadType {
		case 96: // h264 media
			h264SampleBuilder.Push(rtpPacket.Clone())
			for sample := h264SampleBuilder.Pop(); sample != nil; sample = h264SampleBuilder.Pop() {
				// fmt.Println(sample.Duration)
				if err := videoTrackH264Sample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
					log.Fatal(err)
				} else {
					// fmt.Println("Write h264 video sample")
				}
			}
		}
	}
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

		vp8SampleBuilder := samplebuilder.New(0, &codecs.VP8Packet{}, videoTrackVP8Sample.Codec().ClockRate, samplebuilder.WithPacketDroppedHandler(func() {
			fmt.Println("vp8 packet dropped")
		}), samplebuilder.WithPacketReleaseHandler(func(p *rtp.Packet) {
			// fmt.Println("vp8 packet released")
		}))
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
			for sample := vp8SampleBuilder.Pop(); sample != nil; sample = vp8SampleBuilder.Pop() {
				if err := videoTrackVP8Sample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
					log.Fatal(err)
				} else {
					// log.Println("Write video sample")
				}
			}

			// switch rtpPacket.PayloadType {
			// case 96: // h264 media

			// }
		}
	}()
}
