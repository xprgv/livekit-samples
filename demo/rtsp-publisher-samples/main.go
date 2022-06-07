package main

import (
	"fmt"
	"io"
	"time"

	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/LdDl/vdk/av"
	"github.com/LdDl/vdk/codec/h264parser"
	"github.com/LdDl/vdk/format/rtsp"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	// lksdksamplebuilder "github.com/livekit/server-sdk-go/pkg/samplebuilder"
	// pionsamplebuilder "github.com/pion/webrtc/v3/pkg/media/samplebuilder"
	// _ "net/http/pprof"
)

const (
	RTSP_URL = "rtsp://encodertest:encodertest1.Q@192.168.6.12/rtsp_tunnel?p=0&h26x=4"

	AUDIO_MAX_LATE = 32
	VIDEO_MAX_LATE = 256

	TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODc4MjIyNjMsImlzcyI6IkFQSW5BeTI3UlVtWVVuViIsImp0aSI6InRvbnlfc3RhcmsiLCJuYW1lIjoiVG9ueSBTdGFyayIsIm5iZiI6MTY1MTgyMjI2Mywic3ViIjoidG9ueV9zdGFyayIsInZpZGVvIjp7InJvb20iOiJzdGFyay10b3dlciIsInJvb21Kb2luIjp0cnVlfX0.XCuS0Rw73JI8vE6dBUD3WbYGFNz1zGzdUBaDmnuI9Aw"
)

func main() {
	// go http.ListenAndServe("127.0.0.1:8080", nil)

	// config := config.Config{
	// 	Host:      "ws://localhost:7880",
	// 	ApiKey:    "APInAy27RUmYUnV",
	// 	ApiSecret: "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
	// 	Identity:  "publisher-rtp-stream",
	// 	RoomName:  "stark-tower",
	// 	Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODc4MjIyNjMsImlzcyI6IkFQSW5BeTI3UlVtWVVuViIsImp0aSI6InRvbnlfc3RhcmsiLCJuYW1lIjoiVG9ueSBTdGFyayIsIm5iZiI6MTY1MTgyMjI2Mywic3ViIjoidG9ueV9zdGFyayIsInZpZGVvIjp7InJvb20iOiJzdGFyay10b3dlciIsInJvb21Kb2luIjp0cnVlfX0.XCuS0Rw73JI8vE6dBUD3WbYGFNz1zGzdUBaDmnuI9Aw",
	// }

	fmt.Println("connecting to room")
	room, err := lksdk.ConnectToRoom(
		"ws://localhost:7880",
		lksdk.ConnectInfo{
			APIKey:              "APInAy27RUmYUnV",
			APISecret:           "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
			RoomName:            "stark-tower",
			ParticipantIdentity: "publisher-rtp-stream-identity",
			ParticipantName:     "rtp-stream-publisher",
		})
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(room.Metadata())

	// go startOpus(room)
	go startH264(room)
	// go startVP8(room)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()

}

func startH264(room *lksdk.Room) {
	videoTrackH264Sample, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
		Channels:  0,
		// SDPFmtpLine: "125, level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
		SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
		// RTCPFeedback: []webrtc.RTCPFeedback{
		// 	{Type: webrtc.TypeRTCPFBNACK},
		// 	{Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"},
		// },
	})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(videoTrackH264Sample.Codec())

		// if err := videoTrackH264Sample.StartWrite(NewH264SampleProvider(), func() {}); err != nil {
		// 	log.Fatal(err)
		// }
	}

	if _, err := room.LocalParticipant.PublishTrack(videoTrackH264Sample, &lksdk.TrackPublicationOptions{
		Name:   "h264_video",
		Source: livekit.TrackSource_CAMERA,
	}); err != nil {
		log.Fatal(err)
	}

	annexbNALUStartCode := func() []byte { return []byte{0x00, 0x00, 0x00, 0x01} }

	client, err := rtsp.Dial(RTSP_URL)
	if err != nil {
		log.Fatal(err)
	}
	codecs, err := client.Streams()
	if err != nil {
		log.Fatal("failed to get streams:", err)
	}
	for i, t := range codecs {
		log.Println("Stream", i, "is of type", t.Type().String())
	}
	if codecs[0].Type() != av.H264 {
		log.Fatal("RTSP feed must begin with a H264 codec")
	}
	if len(codecs) != 1 {
		log.Println("Ignoring all but the first stream.")
	}

	var previousTime time.Duration
	for {

		pkt, err := client.ReadPacket()
		if err != nil {
			break
		}

		if pkt.Idx != 0 {
			//audio or other stream, skip it
			continue
		}

		pkt.Data = pkt.Data[4:]

		// For every key-frame pre-pend the SPS and PPS
		if pkt.IsKeyFrame {
			pkt.Data = append(annexbNALUStartCode(), pkt.Data...)
			pkt.Data = append(codecs[0].(h264parser.CodecData).PPS(), pkt.Data...)
			pkt.Data = append(annexbNALUStartCode(), pkt.Data...)
			pkt.Data = append(codecs[0].(h264parser.CodecData).SPS(), pkt.Data...)
			pkt.Data = append(annexbNALUStartCode(), pkt.Data...)
		}

		bufferDuration := pkt.Time - previousTime
		previousTime = pkt.Time
		if err = videoTrackH264Sample.WriteSample(media.Sample{Data: pkt.Data, Duration: bufferDuration}, nil); err != nil && err != io.ErrClosedPipe {
			log.Fatal(err)
		}

	}

	// select {}
}
