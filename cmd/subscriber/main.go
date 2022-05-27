package main

import (
	"fmt"
	"livekit-samples/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	go func() {
		time.Sleep(5 * time.Second)
		for _, participant := range room.GetParticipants() {
			for _, track := range participant.Tracks() {
				fmt.Printf("%+v\n", track)
			}
		}
	}()

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

		switch codec.MimeType {
		case webrtc.MimeTypeH264:
			go func() {
				for {
					packet, _, err := tr.ReadRTP()
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf("%+v\n", packet)

				}
			}()
		case webrtc.MimeTypeOpus:
			// go func() {
			// 	for {
			// 		packet, _, err := tr.ReadRTP()
			// 		if err != nil {
			// 			log.Fatal(err)
			// 		}
			// 		fmt.Printf("%+v\n", packet)

			// 	}
			// }()
		case webrtc.MimeTypeVP8:
			go func() {
				for {
					packet, _, err := tr.ReadRTP()
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf("%+v\n", packet)
				}
			}()
			// outputAddr := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5510}
			// sock, err := net.ListenUDP("udp", nil)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// go func() {
			// 	buf := make([]byte, 1500)
			// 	for {
			// 		n, _, err := tr.Read(buf)
			// 		if err != nil {
			// 			log.Fatal(err)
			// 		}

			// 		if _, err := sock.WriteToUDP(buf[:n], &outputAddr); err != nil {
			// 			log.Fatal(err)
			// 		}
			// 	}
			// }()
			// targetAddr := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5500}
			// socket, err := net.ListenUDP("udp", nil)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// go func() {
			// 	buf := make([]byte, )
			// 	for {
			// 		// packet, _, err := tr.ReadRTP()
			// 		// if err != nil {
			// 		// 	log.Fatal(err)
			// 		// }
			// 		// fmt.Printf("%+v\n", packet)

			// 		tr.Read
			// 	}
			// }()
		}

		go func() {
			for {
				if _, _, err := r.ReadRTCP(); err != nil {
					log.Fatal(err)
				}
			}
		}()
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
