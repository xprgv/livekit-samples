package main

import (
	"fmt"
	"livekit-samples/internal/config"
	"livekit-samples/internal/room"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pion/webrtc/v3"
)

func main() {
	config, err := config.GetConfig("./config.toml")
	if err != nil {
		log.Fatal(err)
	}

	room, _ := room.New(config)

	if err := room.Connect(); err != nil {
		log.Fatal(err)
	}

	h264track, ok := room.Tracks[webrtc.MimeTypeH264]
	if ok {
		// h264track.WriteRTP()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	room.Disconnect()

	fmt.Println("disconnecting from room")
}
