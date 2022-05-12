#!/bin/bash

# ffmpeg -re -i output.h264 -f rtp 'rtp://127.0.0.1:5500'

ffmpeg -re -i output.h264 -c:v libx264 -bsf:v h264_mp4toannexb -preset ultrafast -b:v 2M -profile baseline -x264-params keyint=120 -tune zerolatency -f rtp 'rtp://127.0.0.1:5500?pkt_size=1200'

