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
	"github.com/livekit/server-sdk-go/pkg/samplebuilder"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
	// "github.com/pion/webrtc/v3/pkg/media/samplebuilder"
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

	go func() {
		time.Sleep(5 * time.Second)
		for _, participant := range room.GetParticipants() {
			for _, track := range participant.Tracks() {
				fmt.Printf("%+v\n", track)
			}
		}
	}()

	room.Callback.OnTrackSubscribed = onTrackSubscribed

	// videoTrackSample, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
	// 	MimeType:  webrtc.MimeTypeH264,
	// 	ClockRate: 90000,
	// }, "video", "test_video_id")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	videoTrackSample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
		// Channels:  1,
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(videoTrackSample.Codec())

		videoTrackSample.OnBind(func() {
			fmt.Println("Video track on bind")
		})
	}

	// audioTrackSample, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
	// 	MimeType:    webrtc.MimeTypeOpus,
	// 	ClockRate:   48000,
	// 	Channels:    2,
	// 	SDPFmtpLine: "flt",
	// }, "audio", "test_audio_id")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	audioTrackSample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
		// Channels:  2,
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(audioTrackSample.Codec())

		audioTrackSample.OnBind(func() {
			fmt.Println("Audio track on bind")
		})
	}

	if _, err := room.LocalParticipant.PublishTrack(videoTrackSample, &lksdk.TrackPublicationOptions{
		Name:        "video_test",
		Source:      livekit.TrackSource_CAMERA,
		VideoWidth:  480,
		VideoHeight: 360,
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
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20364})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		targetAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5500}
		outputSocket, err := net.ListenUDP("udp", nil)
		if err != nil {
			log.Fatal(err)
		}

		h264SampleBuilder := samplebuilder.New(200, &codecs.H264Packet{}, videoTrackSample.Codec().ClockRate)
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

			// fmt.Printf("%+v\n", rtpPacket)

			if _, err := outputSocket.WriteToUDP(buf[:n], targetAddr); err != nil {
				log.Fatal(err)
			}

			switch rtpPacket.PayloadType {
			case 96: // h264 media

				h264SampleBuilder.Push(&rtpPacket)

				for sample := h264SampleBuilder.Pop(); sample != nil; sample = h264SampleBuilder.Pop() {

					if err := videoTrackSample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
						log.Fatal(err)
					} else {
						// fmt.Println("Write video sample")
					}
				}
			}
		}
	}()

	go func() {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 21360})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		opusSampleBuilder := samplebuilder.New(0, &codecs.OpusPacket{}, audioTrackSample.Codec().ClockRate)
		// opusSampleBuilder := galenesamplebuilder.New(200, &codecs.OpusPacket{}, audioTrackSample.Codec().ClockRate)
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
			// fmt.Printf("%+v\n", rtpPacket)

			switch rtpPacket.PayloadType {
			case 97: // opus media
				opusSampleBuilder.Push(&rtpPacket)
				for sample := opusSampleBuilder.Pop(); sample != nil; sample = opusSampleBuilder.Pop() {
					if err := audioTrackSample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
						log.Fatal(err)
					} else {
						// fmt.Println("Write audio sample")
					}
				}
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}

func onTrackSubscribed(track *webrtc.TrackRemote, publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	fmt.Println("on track subscribed")
}
