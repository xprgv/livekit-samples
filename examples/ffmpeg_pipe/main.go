package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/pion/webrtc/v3/pkg/media/h264reader"
)

const (
	H264_FRAME_DURATION = time.Millisecond * 33

	FFMPEG_CMD = "ffmpeg -rtbufsize 100M -i ./media/files/input.mp4 -pix_fmt yuv420p -c:v libx264 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -f h264 -"
)

func main() {
	cmdStr := strings.Split(FFMPEG_CMD, " ")
	cmd := exec.Command(cmdStr[0], cmdStr[1:]...)

	dataPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	reader, err := h264reader.NewReader(dataPipe)
	if err != nil {
		log.Fatal(err)
	}

	spsAndPpsCache := []byte{}
	ticker := time.NewTicker(H264_FRAME_DURATION)

	for ; true; <-ticker.C {
		nal, err := reader.NextNAL()
		if err != nil {
			// if err == io.EOF {
			// 	fmt.Println("All video frames parsed and sent")
			// 	os.Exit(0)
			// }
			// log.Fatal(err)
			fmt.Println(err)
		}
		if nal != nil {
			nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)

			if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
				spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
				continue
			} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
				nal.Data = append(spsAndPpsCache, nal.Data...)
				spsAndPpsCache = []byte{}
			}
		}

	}
}
