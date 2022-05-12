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

var (
	cfgPath = "./config.toml"
)

const (
	H264_FRAME_DURATION = time.Millisecond * 33

	FFMPEG_CMD = "ffmpeg -rtbufsize 100M -i ./media/files/big_buck_bunny.mp4 -pix_fmt yuv420p -c:v libx264 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -f h264 -"
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

	track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "test_id")
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
			}); err != nil {
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
