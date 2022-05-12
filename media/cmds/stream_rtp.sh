#!/bin/bash


ffmpeg -re -i output.h264 -f lavfi -i testsrc=size=1920x1080:rate=60 -pix_fmt yuv420p -c:v libx264 -g 10 -preset ultrafast -tune zerolatency -f rtp 'rtp://127.0.0.1:5500?pkt_size=1200'

