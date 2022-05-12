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
	cfgPath = "./config.toml"
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

	subscriberPeerConnection := room.LocalParticipant.GetSubscriberPeerConnection()

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
		fmt.Println("New track")
		codec := tr.Codec()

		switch codec.MimeType {
		case webrtc.MimeTypeH264:
			fmt.Println("have h264 track")

			addr := net.UDPAddr{
				IP:   net.ParseIP("238.0.0.1"),
				Port: 8000,
			}
			conn, err := net.ListenUDP("udp", nil)
			if err != nil {
				log.Fatal(err)
			}

			go func() {
				for {
					// buf := make([]byte, 1500)
					pack, _, err := tr.ReadRTP()
					// n, _, err := tr.Read(buf)
					if err != nil {
						if err == io.EOF {
							fmt.Println(err)
							return
						}
						fmt.Println(err)
						continue
					}

					bin, err := pack.Marshal()
					if err != nil {
						fmt.Println("failed to marshal packet", err)
					} else {
						if _, err := conn.WriteToUDP(bin, &addr); err != nil {
							fmt.Println(err)
						}
					}
					// fmt.Println("have rtp packet", pack.CSRC)

					// if _, err := conn.WriteToUDP(buf[:n], &addr); err != nil {
					// 	fmt.Println(err)
					// }
				}
			}()
		}
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
