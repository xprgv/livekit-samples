package room

import (
	"errors"
	"fmt"
	"livekit-samples/internal/config"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/webrtc/v3"
)

type Room struct {
	room   *lksdk.Room
	config config.Config

	Tracks map[string]*webrtc.TrackLocalStaticRTP
}

func New(config config.Config) (*Room, error) {
	r := &Room{
		config: config,
		Tracks: make(map[string]*webrtc.TrackLocalStaticRTP),
	}

	return r, nil
}

func (r *Room) Connect() error {
	room, err := lksdk.ConnectToRoom(
		r.config.Livekit.Host,
		lksdk.ConnectInfo{
			RoomName:            r.config.Livekit.RoomName,
			APIKey:              r.config.Livekit.ApiKey,
			APISecret:           r.config.Livekit.ApiSecret,
			ParticipantName:     r.config.Livekit.ParticipantName,
			ParticipantIdentity: r.config.Livekit.ParticipantIdentity,
		},
	)
	if err != nil {
		return err
	}
	r.room = room
	// r.room.Callback.OnDisconnected = func() {
	// 	fmt.Println("disconnected")
	// }

	for _, output := range r.config.Outputs {
		switch output.Codec {
		case webrtc.MimeTypeH264:
			codecParams := webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypeH264,
				ClockRate: 90000,
			}
			track, err := webrtc.NewTrackLocalStaticRTP(codecParams, "video", "h264_video")
			if err != nil {
				return err
			}
			r.Tracks[webrtc.MimeTypeH264] = track

			if _, err := r.room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{
				Name:   "video",
				Source: livekit.TrackSource_CAMERA,
			}); err != nil {
				return err
			}

		case webrtc.MimeTypeOpus:
			codecParams := webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypeOpus,
				ClockRate: 48000,
			}
			track, err := webrtc.NewTrackLocalStaticRTP(codecParams, "audio", "opus_audio")
			if err != nil {
				return err
			}
			r.Tracks[webrtc.MimeTypeOpus] = track

			if _, err := r.room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{
				Name:   "audio",
				Source: livekit.TrackSource_MICROPHONE,
			}); err != nil {
				return err
			} else {

			}

		default:
			return errors.New(fmt.Sprintf("unknown codec: %s", output.Codec))
		}
	}
	return nil
}

func (r *Room) Disconnect() {
	r.room.Disconnect()
}
