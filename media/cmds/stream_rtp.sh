#!/bin/bash



# ffmpeg -re -i ../files/output.h264 -f lavfi -x264-params keyint=120 -pix_fmt yuv420p -c:v libx264 -g 10 -preset ultrafast -tune zerolatency -f rtp 'rtp://127.0.0.1:5500?pkt_size=1200'

# ffmpeg -re -i ../files/output.h264 -f lavfi -i testsrc=size=1920x1080:rate=60 -x264-params keyint=120 -pix_fmt yuv420p -c:v libx264 -g 10 -preset ultrafast -tune zerolatency -f rtp 'rtp://127.0.0.1:5500?pkt_size=1200'

# ffmpeg -re -i ../files/big_buck_bunny.mp4 -an -c:v libx264 -preset ultrafast -b:v 2M -pix_fmt yuv420p -x264-params keyint=120 -tune zerolatency -f rtp "rtp://127.0.0.1:5500"

# working!
ffmpeg -re -i ../files/output.h264 -c:v libx264 -bsf:v h264_mp4toannexb -preset ultrafast -b:v 2M -profile baseline -x264-params keyint=120 -tune zerolatency -f rtp 'rtp://127.0.0.1:5500?pkt_size=1200'
