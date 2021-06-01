#!/bin/sh -eux

set -ex

RASPBIAN_IMAGE=2021-05-07-raspios-buster-armhf-lite.img

# prepare
# =======
mkdir /SWAP
mount -t 9p -o trans=virtio,version=9p2000.L host0 /SWAP

echo "http://dl-cdn.alpinelinux.org/alpine/v3.13/community/" >>/etc/apk/repositories
apk add aria2 coreutils e2fsprogs-extra parted util-linux multipath-tools

cd /SWAP

aria2c --seed-time=0 https://downloads.raspberrypi.org/raspios_lite_armhf/images/raspios_lite_armhf-2021-05-28/2021-05-07-raspios-buster-armhf-lite.zip.torrent
unzip 2021-05-07-raspios-buster-armhf-lite.zip

rm 2021-05-07-raspios-buster-armhf-lite.zip*

# building
# ========
imagePath=$RASPBIAN_IMAGE

# resize partition
truncate -s +5G "$imagePath"
parted "$imagePath" resizepart 2 5G

# create loop devices
kpartx -avs "$imagePath"

ls -la /dev/mapper

bootLoopDevice="$(find /dev/mapper -name 'loop*p1')"
systemLoopDevice="$(find /dev/mapper -name 'loop*p2')"

if [ -z "$bootLoopDevice" ] || [ -z "$systemLoopDevice" ]; then
        echo "could not extract loop devices"
        exit 1
fi

# resize filesystem
resize2fs "$systemLoopDevice"

# enable ssh
mkdir -p boot
mount "$bootLoopDevice" boot
touch boot/ssh

# cleaning up
umount boot
kpartx -d "$imagePath"
rm -rf boot
