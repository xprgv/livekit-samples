package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/livekit/server-sdk-go/pkg/samplebuilder"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
)

func startVP8(room *lksdk.Room) {
	videoTrackVP8Sample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeVP8,
		ClockRate: 90000,
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(videoTrackVP8Sample.Codec())
	}

	if publication, err := room.LocalParticipant.PublishTrack(videoTrackVP8Sample, &lksdk.TrackPublicationOptions{
		Name:   "video_vp8_test",
		Source: livekit.TrackSource_CAMERA,
	}); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(publication.MimeType())
	}

	go func() {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5500})
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		vp8SampleBuilder := samplebuilder.New(0, &codecs.VP8Packet{}, videoTrackVP8Sample.Codec().ClockRate, samplebuilder.WithPacketDroppedHandler(func() {
			fmt.Println("vp8 packet dropped")
		}), samplebuilder.WithPacketReleaseHandler(func(p *rtp.Packet) {
			// fmt.Println("vp8 packet released")
		}))
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
			// fmt.Printf("%+v\n", rtpPacket)

			vp8SampleBuilder.Push(rtpPacket.Clone())
			for sample := vp8SampleBuilder.Pop(); sample != nil; sample = vp8SampleBuilder.Pop() {
				if err := videoTrackVP8Sample.WriteSample(*sample, nil); err != nil && err != io.ErrClosedPipe {
					log.Fatal(err)
				} else {
					// log.Println("Write video sample")
				}
			}

			// switch rtpPacket.PayloadType {
			// case 96: // h264 media

			// }
		}
	}()
}
