---
title: Getting your SD card to detect on a Nintendo 2DS/3DS 
summary: A fix for "Could not detect an SD Card. The software on the SD Card could not be displayed."
time: 1690692039
---

So I recently got a Nintendo 2DS for homebrew reasons, and ran into this incredibly annoying error while trying to follow the guide laid out by the lovely people at https://3ds.hacks.guide:

<figure>
<img src="/img/2ds-sd-card/2ds.jpg" class="vertical">
<figcaption>The home screen on my Nintendo 2DS, with some Miiverse thing on the top and the following error message on the bottom: "Could not detect an SD Card. The software on the SD Card could not be displayed."</figcaption>
</figure>

Googling around did not show anything fruitful, but I found a hint on this [Nintendo support page](https://en-americas-support.nintendo.com/app/answers/detail/a_id/220/~/how-to-format-an-sd-card-or-microsd-card), which claims that an SD card must be formatted with the official SD Association formatting utility. If you're on Windows or macOS, you should use [this tool](https://www.sdcard.org/downloads/formatter/) and that will probably format your card correctly. If you're on Linux or can't use this tool for some reason, read on.

It turns out that:

1. Your SD card must be formatted with an MBR (Master Boot Record)
2. The partition type of your (presumably one and only) partition on the SD card must be FAT32

Note that this doesn't _just_ mean your partition should be formatted as FAT32. It should be formatted as FAT32, _and also_ there's a "partition type" byte inside your MBR that should be changed to FAT32. On a Linux system with `fdisk`, this is how you'd do that:

* Unmount but don't safely eject your SD card
* Run `sudo fdisk /dev/sdX`, where `sdX` is your SD card
* Type `t` to change your partition type, `1` to select your first partition, and `b` to make the partition "W95 FAT32" type
* Type `w` to save and quit

And now your SD card should detect fine and you should be able to homebrew it or play Nintendo-licensed games on it or whatever.