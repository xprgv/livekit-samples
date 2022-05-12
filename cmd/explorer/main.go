package main

import (
	"fmt"
	"io"
	"livekit-samples/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"

	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/webrtc/v3"
)

var (
	cfgPath = "./cmd/explorer/config.toml"
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

	pc := room.LocalParticipant.GetSubscriberPeerConnection()
	pc.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		codec := tr.Codec()
		log.Println("track codec:", codec.MimeType)
		switch codec.MimeType {
		case webrtc.MimeTypeH264:
			go func() {
				for {
					rtpH264Packet, _, err := tr.ReadRTP()
					if err != nil {
						if err == io.EOF {
							log.Println(err)
							break
						}
						log.Println(err)
						continue
					}
					log.Println("rtp h264 packet: ", rtpH264Packet.Timestamp)
				}
			}()
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
