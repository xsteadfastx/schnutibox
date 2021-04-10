#!/bin/sh
set -ex

if [ -z "$RASPBIAN_IMAGE" ]; then
        echo "missing \$RASPBIAN_IMAGE environment variable"
        exit 1
fi
imagePath=$RASPBIAN_IMAGE

# resize partition
truncate -s +3G "$imagePath"
parted "$imagePath" resizepart 2 3G

# create loop devices
bootLoopDevice="$(kpartx -l "$imagePath" | sed -n 1p | awk '{print $1}')"
systemLoopDevice="$(kpartx -l "$imagePath" | sed -n 2p | awk '{print $1}')"

if [ -z "$bootLoopDevice" ] || [ -z "$systemLoopDevice" ]; then
        echo "could not extract loop devices"
        exit 1
fi

kpartx -avs "$imagePath"

# resize filesystem
resize2fs "/dev/mapper/${systemLoopDevice}"

# enable ssh
mkdir -p boot
mount -o loop "/dev/mapper/${bootLoopDevice}" boot
touch boot/ssh

# cleaning up
umount boot
kpartx -d "$imagePath"
rm -rf boot
