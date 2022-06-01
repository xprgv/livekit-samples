package main

import (
	"fmt"
	"io"
	"livekit-samples/internal/config"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
)

const (
	H264_FRAME_DURATION = time.Millisecond * 33

	FFMPEG_CMD = "ffmpeg -rtbufsize 100M -i ../../media/files/big_buck_bunny.mp4 -pix_fmt yuv420p -c:v libx264 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -f h264 -"
)

func main() {
	config := config.Config{
		Host:      "ws://localhost:7880",
		ApiKey:    "APInAy27RUmYUnV",
		ApiSecret: "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
		Identity:  "publisher-from-file",
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

	track, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264})
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

	subscriberPeerConnection := room.LocalParticipant.GetSubscriberPeerConnection()
	subscriberPeerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		fmt.Println("ICE candidate subscriber:", i.String())
	})

	peerConnectionPublisher := room.LocalParticipant.GetPublisherPeerConnection()
	peerConnectionPublisher.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		fmt.Println("ICE candidate publisher ", i.String())
	})

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
		cmdStr := strings.Split(FFMPEG_CMD, " ")

		cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
		dataPipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		h264Reader, err := h264reader.NewReader(dataPipe)
		if err != nil {
			log.Fatal(err)
		}

		spsAndPpsCache := []byte{}
		ticker := time.NewTicker(H264_FRAME_DURATION)

		for {
			<-ticker.C

			nal, err := h264Reader.NextNAL()
			if err != nil {
				if err == io.EOF {
					fmt.Println("All video frames parsed and sent")
					os.Exit(0)
				}
				log.Fatal(err)
			}

			fmt.Println(nal.UnitType.String())

			nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)

			if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
				spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
				continue
			} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
				nal.Data = append(spsAndPpsCache, nal.Data...)
				spsAndPpsCache = []byte{}
			}

			if err := track.WriteSample(media.Sample{
				Data:     nal.Data,
				Duration: time.Second,
			}, nil); err != nil {
				log.Fatal(err)
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
