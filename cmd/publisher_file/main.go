package main

import (
	"fmt"
	"livekit-samples/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	lksdk "github.com/livekit/server-sdk-go"
)

var (
	mediaFile = "./media/files/output.h264"

	cfgPath = "./cmd/publisher_file/config.toml"
)

func main() {
	config, err := config.GetConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("connecting to the room")
	room, err := lksdk.ConnectToRoom(config.Host, lksdk.ConnectInfo{
		APIKey:              config.ApiKey,
		APISecret:           config.ApiSecret,
		RoomName:            config.RoomName,
		ParticipantIdentity: config.Identity,
	})
	if err != nil {
		log.Fatal(err)
	}

	track, err := lksdk.NewLocalFileTrack(
		mediaFile, lksdk.FileTrackWithFrameDuration(33*time.Millisecond),
		lksdk.FileTrackWithOnWriteComplete(func() { fmt.Println("track finished") }),
	)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{}); err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
