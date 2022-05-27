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

	"github.com/jech/samplebuilder"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
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

	// videoTrackH264Sample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
	// 	MimeType: webrtc.MimeTypeH264,
	// 	// ClockRate: 90000,
	// 	// Channels:  1,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	// myProvider, err := sampler.NewRtpH264StreamSampleProvider()
	// 	// if err != nil {
	// 	// 	log.Fatal(err)
	// 	// }
	// 	// if err := videoTrackH264Sample.StartWrite(myProvider, func() {}); err != nil {
	// 	// 	log.Fatal(err)
	// 	// }
	// 	// fmt.Println(videoTrackH264Sample.Codec())
	// }

	// videoTrackVP8Sample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8})
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	fmt.Println(videoTrackVP8Sample.Codec())
	// }

	audioTrackSample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
		Channels:  2,
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(audioTrackSample.Codec())

		// sock, err := net.ListenUDP("udp", nil)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// opusSampleProvider := sampleprovider.NewRtpSampleProvider(
		// 	audioTrackSample.Codec().MimeType,
		// 	sock,
		// 	samplebuilder.New(0, &codecs.OpusPacket{}, audioTrackSample.Codec().ClockRate),
		// )
		// opusSampleProvider, err := sampler.NewRtpOpusSampleProvider()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// if err := audioTrackSample.StartWrite(opusSampleProvider, func() {}); err != nil {
		// 	log.Fatal(err)
		// }
	}

	// if _, err := room.LocalParticipant.PublishTrack(videoTrackH264Sample, &lksdk.TrackPublicationOptions{
	// 	Name:   "video_test",
	// 	Source: livekit.TrackSource_CAMERA,
	// }); err != nil {
	// 	log.Fatal(err)
	// }

	// if publication, err := room.LocalParticipant.PublishTrack(videoTrackVP8Sample, &lksdk.TrackPublicationOptions{
	// 	Name:   "video_vp8_test",
	// 	Source: livekit.TrackSource_CAMERA,
	// }); err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	fmt.Println(publication.MimeType())
	// }

	if _, err := room.LocalParticipant.PublishTrack(audioTrackSample, &lksdk.TrackPublicationOptions{
		Name:   "audio_test",
		Source: livekit.TrackSource_MICROPHONE,
	}); err != nil {
		log.Fatal(err)
	}

	// go func() {
	// 	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5500})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer listener.Close()

	// 	// mysampleProvider := lksdk.NewNullSampleProvider(90000)

	// 	h264SampleBuilder := samplebuilder.New(1000, &codecs.H264Packet{}, videoTrackH264Sample.Codec().ClockRate)
	// 	buf := make([]byte, 1500)
	// 	for {
	// 		n, _, err := listener.ReadFromUDP(buf)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		rtpPacket := rtp.Packet{}
	// 		if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		switch rtpPacket.PayloadType {
	// 		case 96: // h264 media
	// 			h264SampleBuilder.Push(&rtpPacket)
	// 			for sample := h264SampleBuilder.Pop(); sample != nil; sample = h264SampleBuilder.Pop() {
	// 				// videoTrackH264Sample.WriteSample()
	// 				if err := videoTrackH264Sample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
	// 					log.Fatal(err)
	// 				} else {
	// 					fmt.Println("Write h264 video sample")
	// 				}
	// 			}
	// 		}
	// 	}
	// }()

	// go func() {
	// 	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5500})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer listener.Close()

	// 	vp8SampleBuilder := samplebuilder.New(200, &codecs.VP8Packet{}, videoTrackVP8Sample.Codec().ClockRate)
	// 	buf := make([]byte, 1500)
	// 	for {
	// 		n, _, err := listener.ReadFromUDP(buf)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		rtpPacket := rtp.Packet{}
	// 		if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		// fmt.Printf("%+v\n", rtpPacket)

	// 		vp8SampleBuilder.Push(&rtpPacket)
	// 		for sample := vp8SampleBuilder.Pop(); sample != nil; sample = vp8SampleBuilder.Pop() {
	// 			if err := videoTrackVP8Sample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
	// 				log.Fatal(err)
	// 			} else {
	// 				// log.Println("Write video sample")
	// 			}
	// 		}

	// 		// switch rtpPacket.PayloadType {
	// 		// case 96: // h264 media

	// 		// }
	// 	}
	// }()

	go func() {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 21360})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		opusSampleBuilder := samplebuilder.New(0, &codecs.OpusPacket{}, audioTrackSample.Codec().ClockRate)
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
