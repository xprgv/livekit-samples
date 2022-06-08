package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/webrtc/v3"
)

var (
	cfgPath = "./config.toml"
)

func main() {
	room, err := lksdk.ConnectToRoom(
		"ws://localhost:7880", lksdk.ConnectInfo{
			APIKey:              "APInAy27RUmYUnV",
			APISecret:           "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
			RoomName:            "stark-tower",
			ParticipantIdentity: "subscriber",
			ParticipantName:     "subscriber",
		})
	if err != nil {
		log.Fatal(err)
	}

	subscriberPeerConnection := room.LocalParticipant.GetSubscriberPeerConnection()
	subscriberPeerConnection.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		codec := tr.Codec()
		fmt.Println("New track:", codec.MimeType)

		switch codec.MimeType {
		case webrtc.MimeTypeH264:
			go func() {
				for {
					_, _, err := tr.ReadRTP()
					if err != nil {
						log.Fatal(err)
					}
					// fmt.Printf("%+v\n", packet)
				}
			}()

		case webrtc.MimeTypeVP8:
			go func() {
				for {
					_, _, err := tr.ReadRTP()
					if err != nil {
						log.Fatal(err)
					}
					// fmt.Printf("%+v\n", packet)
				}
			}()
		}

		// go func() {
		// 	for {
		// 		pkts, _, err := r.ReadRTCP()
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		for _, pkt := range pkts {
		// 			fmt.Println(pkt)
		// 		}
		// 	}
		// }()
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
