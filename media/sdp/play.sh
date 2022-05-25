#!/bin/bash

# ffplay -protocol_whitelist file,rtp,udp -i video_low.sdp &
ffplay -protocol_whitelist file,rtp,udp -i video_mid.sdp &
# ffplay -protocol_whitelist file,rtp,udp -i video_high.sdp &
# ffplay -protocol_whitelist file,rtp,udp -i audio.sdp
