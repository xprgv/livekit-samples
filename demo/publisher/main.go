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
	"github.com/pion/webrtc/v3"
)

var (
	cfgPath = "./cmd/publisher/config.toml"
)

func main() {
	fmt.Println("connecting to room")
	room, err := lksdk.ConnectToRoom(
		"ws://localhost:7880", lksdk.ConnectInfo{
			APIKey:              "APInAy27RUmYUnV",
			APISecret:           "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
			RoomName:            "stark-tower",
			ParticipantIdentity: "publisher",
			ParticipantName:     "publisher",
		})
	if err != nil {
		log.Fatal(err)
	}
	track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "test_id")
	if err != nil {
		log.Fatal(err)
	}

	trackPublicationOptions := lksdk.TrackPublicationOptions{
		Name:        "my test h264 track",
		Source:      livekit.TrackSource_CAMERA,
		VideoWidth:  1920,
		VideoHeight: 1080,
	}

	trackPublication, err := room.LocalParticipant.PublishTrack(track, &trackPublicationOptions)
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

			// fmt.Println("pack")

			_, err = track.Write(buf[:n])
			if err != nil {
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
