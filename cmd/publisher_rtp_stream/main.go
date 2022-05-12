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

	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/webrtc/v3"
)

var (
	cfgPath = "./cmd/publisher_rtp_stream/config.toml"
)

func main() {
	config, err := config.GetConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("connecting to room")
	room, err := lksdk.ConnectToRoom(config.Host, lksdk.ConnectInfo{
		APIKey:              config.ApiKey,
		APISecret:           config.ApiSecret,
		RoomName:            config.RoomName,
		ParticipantIdentity: config.Identity,
	})
	if err != nil {
		log.Fatal(err)
	}

	// track, err := lksdk.NewLocalSampleTrack(
	// 	webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
	// 	func(s *lksdk.LocalSampleTrack) {
	// 		fmt.Println("local sample track")
	// 	},
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// track.SetTransceiver(&webrtc.RTPTransceiver{})
	// track.SetTransceiver()

	track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "test_id")
	if err != nil {
		log.Fatal(err)
	}

	pcPublisher := room.LocalParticipant.GetPublisherPeerConnection()

	pcPublisher.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
		fmt.Println("Connection state has changed:", is.String())
	})

	rtpSender, err := pcPublisher.AddTrack(track)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		buf := make([]byte, 1500)
		for {
			if _, _, err := rtpSender.Read(buf); err != nil {
				return
			}
		}
	}()

	go func() {
		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:5500")
		if err != nil {
			log.Fatal(err)
		}
		listener, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		for {
			buf := make([]byte, 1600)
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}

			_, err = track.Write(buf[:n])
			if err != nil {
				log.Println(err)
			}
		}
	}()

	pcSubscriber := room.LocalParticipant.GetSubscriberPeerConnection()

	pcSubscriber.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		codec := tr.Codec()
		log.Println("track codec:", codec.MimeType)
		switch codec.MimeType {
		case webrtc.MimeTypeVP8:
			go func() {
				for {
					rtpOpusPacket, _, err := tr.ReadRTP()
					if err != nil {
						if err == io.EOF {
							log.Println(err)
							break
						}
						log.Println(err)
						continue
					}
					log.Println("rtp packet", rtpOpusPacket.Timestamp)
				}
			}()
		}

	})

	// pcPublisher.AddTransceiverFromTrack(&webrtc.TrackLocalStaticRTP{})
	// pcPublisher.AddTrack

	// videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// go func() {
	// 	for {
	// 		if err := videoTrack.WriteRTP(&rtp.Packet{}); err != nil {
	// 			log.Println(err)
	// 		}
	// 		time.Sleep(100 * time.Millisecond)
	// 	}
	// }()

	// rtpVideoSender, err := pcPublisher.AddTrack(videoTrack)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// go func() {
	// 	rtcpBuf := make([]byte, 1500)
	// 	for {
	// 		if _, _, err := rtpVideoSender.Read(rtcpBuf); err != nil {
	// 			return
	// 		}
	// 	}
	// }()

	// go func() {
	// 	file, err := os.Open("./output.mp4")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer file.Close()

	// 	reader, err := h264reader.NewReader(file)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	nal, err := reader.NextNAL()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	ivf, header, err := ivfreader.NewWith(file)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	ticker := time.NewTicker(time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000))
	// 	for {
	// 		<-ticker.C
	// 		frame, _, err := ivf.ParseNextFrame()
	// 		if errors.Is(err, io.EOF) {
	// 			fmt.Printf("All video frames parsed and sent")
	// 			os.Exit(0)
	// 		}

	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		if err = videoTrack.WriteSample(media.Sample{Data: frame, Duration: time.Second}); err != nil {
	// 			log.Fatal(err)
	// 		}
	// 	}
	// }()

	// transceiver, err := pcPublisher.AddTransceiverFromTrack(videoTrack)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// transceiver.Mid()
	// pcPublisher.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{})

	// audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "stream_id_test")
	// videoRtpSender, err := pcPublisher.AddTrack(videoTrack)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// go func() {
	// 	file, err := os.Open("./video")
	// }()

	// track, err := lksdk.NewLocalFileTrack(
	// 	filePath,
	// 	lksdk.FileTrackWithFrameDuration(33*time.Millisecond),
	// 	lksdk.FileTrackWithOnWriteComplete(func() {
	// 		fmt.Println("track finished")
	// 	}),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// webrtc.NewTrackLo

	// if _, err := room.LocalParticipant.PublishTrack(track); err != nil {
	// 	log.Fatal(err)
	// }

	// participants := room.GetParticipants()
	// for _, participant := range participants {
	// 	log.Println(
	// 		participant.Name(),
	// 		participant.SID(),
	// 		participant.Identity(),
	// 		"screen:", participant.IsScreenShareEnabled(),
	// 		"mic:", participant.IsMicrophoneEnabled(),
	// 		"camera:", participant.IsCameraEnabled(),
	// 	)
	// }

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
