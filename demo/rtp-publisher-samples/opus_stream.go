package main

import (
	"fmt"
	"io"
	"livekit-samples/pkg/codecs"
	"livekit-samples/pkg/lksdksamplebuilder"
	"log"
	"net"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

func startOpus(room *lksdk.Room) {
	audioTrackSample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
		Channels:  2,
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(audioTrackSample.Codec(), audioTrackSample.IsBound())
	}

	if _, err := room.LocalParticipant.PublishTrack(audioTrackSample, &lksdk.TrackPublicationOptions{
		Name:   "opus_audio",
		Source: livekit.TrackSource_MICROPHONE,
	}); err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 21360})
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	opusSampleBuilder := lksdksamplebuilder.New(AUDIO_MAX_LATE, &codecs.OpusPacket{}, audioTrackSample.Codec().ClockRate)

	buf := make([]byte, 1500)
	rtpPacket := rtp.Packet{}
	for {
		n, _, err := listener.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}

		if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
			log.Fatal(err)
		}

		switch rtpPacket.PayloadType {
		case 97: // opus media
			opusSampleBuilder.Push(rtpPacket.Clone())
			for sample := opusSampleBuilder.Pop(); sample != nil; sample = opusSampleBuilder.Pop() {
				if err := audioTrackSample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
					log.Fatal(err)
				}
			}
		}
	}
}
