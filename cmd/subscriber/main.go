package main

import (
	"fmt"
	"livekit-samples/internal/config"
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
	// config, err := config.GetConfig(cfgPath)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	config := config.Config{
		Host:      "ws://localhost:7880",
		ApiKey:    "APInAy27RUmYUnV",
		ApiSecret: "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
		Identity:  "subscriber",
		RoomName:  "stark-tower",
		Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODc4MjIyNjMsImlzcyI6IkFQSW5BeTI3UlVtWVVuViIsImp0aSI6InRvbnlfc3RhcmsiLCJuYW1lIjoiVG9ueSBTdGFyayIsIm5iZiI6MTY1MTgyMjI2Mywic3ViIjoidG9ueV9zdGFyayIsInZpZGVvIjp7InJvb20iOiJzdGFyay10b3dlciIsInJvb21Kb2luIjp0cnVlfX0.XCuS0Rw73JI8vE6dBUD3WbYGFNz1zGzdUBaDmnuI9Aw",
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

	publisherPeerConnection := room.LocalParticipant.GetPublisherPeerConnection()
	publisherPeerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		fmt.Println("ICE candidate from publisher pc:", i.String())
	})

	subscriberPeerConnection := room.LocalParticipant.GetSubscriberPeerConnection()

	subscriberPeerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		fmt.Println("ICE candidate from subscriber pc:", i.String())
	})

	subscriberPeerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		fmt.Println("have datachannel", dc.Label())

		dc.OnClose(func() {
			fmt.Println("datachannel closed")
		})

		dc.OnError(func(err error) {
			fmt.Println("Error in datachannel", err)
		})

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Println("datachannel message:", string(msg.Data))
		})
	})

	subscriberPeerConnection.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		codec := tr.Codec()
		fmt.Println("New track:", codec.MimeType)

		// go func() {
		// 	for {
		// 		pack, _, err := r.ReadRTCP()
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		fmt.Println(pack)
		// 	}
		// }()

		// go func() {
		// 	for {
		// 		pack, _, err := tr.ReadRTP()
		// 		if err != nil {
		// 			if err == io.EOF {
		// 				return
		// 			} else {
		// 				log.Fatal(err)
		// 			}

		// 		}
		// 		fmt.Println(pack.PayloadType)
		// 	}
		// }()

		// switch codec.MimeType {
		// case webrtc.MimeTypeH264:
		// 	addr := net.UDPAddr{
		// 		IP:   net.ParseIP("238.0.0.1"),
		// 		Port: 8000,
		// 	}
		// 	conn, err := net.ListenUDP("udp", nil)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}

		// 	go func() {
		// 		for {
		// 			// buf := make([]byte, 1500)
		// 			pack, _, err := tr.ReadRTP()
		// 			// n, _, err := tr.Read(buf)
		// 			if err != nil {
		// 				if err == io.EOF {
		// 					fmt.Println(err)
		// 					return
		// 				}
		// 				fmt.Println(err)
		// 				continue
		// 			}
		// 			// fmt.Println("get packet")

		// 			bin, err := pack.Marshal()
		// 			if err != nil {
		// 				fmt.Println("failed to marshal packet", err)
		// 			} else {
		// 				if _, err := conn.WriteToUDP(bin, &addr); err != nil {
		// 					fmt.Println(err)
		// 				}
		// 			}
		// 			// fmt.Println("have rtp packet", pack.CSRC)

		// 			// if _, err := conn.WriteToUDP(buf[:n], &addr); err != nil {
		// 			// 	fmt.Println(err)
		// 			// }
		// 		}
		// 	}()
		// }
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
