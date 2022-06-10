package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

var (
	h264RtpChan = make(chan rtp.Packet, 100)
	opusRtpChan = make(chan rtp.Packet, 100)
)

func main() {
	subscriberRoom, err := lksdk.ConnectToRoom(
		"ws://localhost:7880", lksdk.ConnectInfo{
			APIKey:              "APInAy27RUmYUnV",
			APISecret:           "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
			RoomName:            "stark-tower",
			ParticipantIdentity: "publisher-rtp-stream",
			ParticipantName:     "rtp-stream-publisher",
		})
	if err != nil {
		log.Fatal(err)
	}

	publisherRoom, err := lksdk.ConnectToRoom(
		"ws://localhost:7880", lksdk.ConnectInfo{
			APIKey:              "APInAy27RUmYUnV",
			APISecret:           "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
			RoomName:            "stark-tower",
			ParticipantIdentity: "publisher-rtp-stream",
			ParticipantName:     "rtp-stream-publisher",
		})
	if err != nil {
		log.Fatal(err)
	}

	h264Track1080p, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
	}, "video", "h264_video")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := publisherRoom.LocalParticipant.PublishTrack(h264Track1080p, &lksdk.TrackPublicationOptions{
		Name:   "video",
		Source: livekit.TrackSource_CAMERA,
	}); err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			rtpPacket := <-h264RtpChan
			if err := h264Track1080p.WriteRTP(&rtpPacket); err != nil {
				log.Fatal(err)
			}
		}
	}()

	opusTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
		Channels:  2,
	}, "audio", "audio")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := publisherRoom.LocalParticipant.PublishTrack(opusTrack, &lksdk.TrackPublicationOptions{
		Name:   "audio",
		Source: livekit.TrackSource_MICROPHONE,
	}); err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			rtpPacket := <-opusRtpChan
			if err := opusTrack.WriteRTP(&rtpPacket); err != nil {
				log.Fatal(err)
			}
		}
	}()

	subscriberPeerConnection := subscriberRoom.LocalParticipant.GetSubscriberPeerConnection()
	subscriberPeerConnection.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		codec := tr.Codec()
		fmt.Println("new track", codec.MimeType)
		switch codec.MimeType {
		case webrtc.MimeTypeH264:
			rtpPacket, _, err := tr.ReadRTP()
			if err != nil {
				log.Fatal(err)
			}
			h264RtpChan <- *rtpPacket
		case webrtc.MimeTypeOpus:
			rtpPacket, _, err := tr.ReadRTP()
			if err != nil {
				log.Fatal(err)
			}
			opusRtpChan <- *rtpPacket
		default:
			log.Println("Unhandled codec:", codec.MimeType)
		}
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	subscriberRoom.Disconnect()
	publisherRoom.Disconnect()

	// fmt.Println("disconnecting from room")
}
