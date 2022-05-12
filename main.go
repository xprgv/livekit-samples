package main

import (
	"context"
	"fmt"
	"livekit-samples/internal/config"
	"log"
	"time"

	"github.com/livekit/protocol/auth"
	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
)

func getJoinToken(apiKey, apiSecret, room, identity string) (string, error) {
	at := auth.NewAccessToken(apiKey, apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	at.AddGrant(grant).
		SetIdentity(identity).
		SetValidFor(time.Hour)

	return at.ToJWT()
}

var (
	cfgPath = "./config.toml"
)

func main() {
	config, err := config.GetConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	roomServiceClient := lksdk.NewRoomServiceClient("http://localhost:7880", config.ApiKey, config.ApiSecret)

	// create a new room
	// room, err := roomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
	// 	Name: roomName,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("active recording", room.ActiveRecording)

	// room, err := roomServiceClient.UpdateRoomMetadata(context.Background(), &livekit.UpdateRoomMetadataRequest{
	// 	Room:     config.RoomName,
	// 	Metadata: "hello-world-meta",
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("metadata:", room.GetMetadata())

	// resp, err := roomServiceClient.RoomService.UpdateSubscriptions(
	// 	context.Background(),
	// 	&livekit.UpdateSubscriptionsRequest{},
	// )

	// list rooms
	roomsResponse, err := roomServiceClient.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	// localSampleTrack := lksdk.LocalSampleTrack{}
	// h264Track, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, func(s *lksdk.LocalSampleTrack) {})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// opusTrack, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, func(s *lksdk.LocalSampleTrack) {})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(roomsResponse.String())
	for _, room := range roomsResponse.Rooms {
		// fmt.Printf("%+v\n", room)
		fmt.Println(room.Name, room.NumParticipants, "codecs", room.EnabledCodecs)
		// codecs := room.GetEnabledCodecs()
		// for _, codec := range codecs {
		// 	fmt.Println(codec.Mime)
		// }
	}

	// // terminate a room and cause participants to leave
	// roomClient.DeleteRoom(context.Background(), &livekit.DeleteRoomRequest{
	// 	Room: roomId,
	// })

	// // list participants in a room
	// res, _ := roomClient.ListParticipants(context.Background(), &livekit.ListParticipantsRequest{
	// 	Room: roomName,
	// })

	// // disconnect a participant from room
	// roomClient.RemoveParticipant(context.Background(), &livekit.RoomParticipantIdentity{
	// 	Room:     roomName,
	// 	Identity: identity,
	// })

	// // mute/unmute participant's tracks
	// roomClient.MutePublishedTrack(context.Background(), &livekit.MuteRoomTrackRequest{
	// 	Room:     roomName,
	// 	Identity: identity,
	// 	TrackSid: "track_sid",
	// 	Muted:    true,
	// })
}
