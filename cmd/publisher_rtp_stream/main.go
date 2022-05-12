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

	track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "test_id")
	if err != nil {
		log.Fatal(err)
	}

	if err := room.LocalParticipant.PublishData([]byte("test data"), livekit.DataPacket_RELIABLE, []string{}); err != nil {
		log.Fatal("Failed to publish data")
	}

	trackPublication, err := room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{
		Name:        "my test h264 track",
		Source:      livekit.TrackSource_CAMERA,
		VideoWidth:  1920,
		VideoHeight: 1080,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(trackPublication.Name())

	peerConnectionPublisher := room.LocalParticipant.GetPublisherPeerConnection()

	rtpSender, err := peerConnectionPublisher.AddTrack(track)
	go func() {
		buf := make([]byte, 1500)
		for {
			if _, _, err := rtpSender.Read(buf); err != nil {
				return
			}
		}
	}()

	go func() {
		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:5500")
		if err != nil {
			log.Fatal(err)
		}
		listener, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		fmt.Println("listen")

		for {
			buf := make([]byte, 1600)
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}

			if _, err := track.Write(buf[:n]); err != nil {
				log.Println(err)
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
