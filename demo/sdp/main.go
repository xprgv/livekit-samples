package main

import (
	"fmt"
	"livekit-samples/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/webrtc/v3"
)

var (
	cfgPath = "./config.toml"
)

func main() {
	config := config.Config{
		Host:      "ws://localhost:7880",
		ApiKey:    "APInAy27RUmYUnV",
		ApiSecret: "90jQt67cwele8a6uIuIQLK0ZJ0cJKXnzz6iEI8h43dO",
		Identity:  "get-sdp",
		RoomName:  "stark-tower",
		Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODc4MjIyNjMsImlzcyI6IkFQSW5BeTI3UlVtWVVuViIsImp0aSI6InRvbnlfc3RhcmsiLCJuYW1lIjoiVG9ueSBTdGFyadsyIsIm5iZiI6MTY1MTgyMjI2Mywic3ViIjoidG9ueV9zdGFyayIsInZpZGVvIjp7InJvb20iOiJzdGFyay10b3dlciIsInJvb21Kb2luIjp0cnVlfX0.XCuS0Rw73JI8vE6dBUD3WbYGFNz1zGzdUBaDmnuI9Aw",
	}

	signalClient(config)
	// connectRoom(config)
}

func signalClient(config config.Config) {
	fmt.Println("Signal client")
	accessToken := auth.NewAccessToken(config.ApiKey, config.ApiSecret)

	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     config.RoomName,
	}
	accessToken.AddGrant(grant).SetIdentity(config.Identity).SetName("get sdp name")

	token, err := accessToken.ToJWT()
	if err != nil {
		log.Fatal(err)
	}

	signalClient := lksdk.NewSignalClient()

	joinResponse, err := signalClient.Join(config.Host, token, &lksdk.ConnectParams{AutoSubscribe: false})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("server version:", joinResponse.ServerVersion)

	signalClient.OnOffer = func(sd webrtc.SessionDescription) {
		fmt.Printf("SDP offer: %+v\n", sd)
	}

	signalClient.OnAnswer = func(sd webrtc.SessionDescription) {
		fmt.Printf("SDP answer: %+v\n", sd)
	}

	signalClient.OnParticipantUpdate = func(pi []*livekit.ParticipantInfo) {
		fmt.Println("participant update", pi)
	}

	select {}
}

func connectRoom(config config.Config) {
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

	room.Callback.OnParticipantConnected = func(rp *lksdk.RemoteParticipant) {
		fmt.Printf("participant connected: %+v\n", rp.Name())

		rp.Callback.OnTrackPublished = func(publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
			fmt.Println("publication:", publication.MimeType())
		}
	}

	room.Callback.OnParticipantDisconnected = func(rp *lksdk.RemoteParticipant) {
		fmt.Printf("participant disconnected: %+v\n", rp)
	}

	subscriberPeerConnection := room.LocalParticipant.GetSubscriberPeerConnection()

	subscriberPeerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i != nil {
			fmt.Printf("%+v\n", i)
		}
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan

	fmt.Println("disconnecting from room")
	room.Disconnect()
}
