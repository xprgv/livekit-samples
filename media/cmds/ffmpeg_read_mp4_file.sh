#!/bin/bash

ffmpeg -rtbufsize 100M -i ./media/files/big_buck_bunny.mp4 -pix_fmt yuv420p -c:v libx264 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -f h264 -
