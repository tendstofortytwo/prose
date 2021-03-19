---
title: "From the old blog: Getting My Second-Hand Laptop's Fingerprint Sensor Working"
summary: A guide to get VFS491 fingerprint scanner working in Windows 10 with Windows Hello.
time: 1552516920
---

***Note:** This article was posted on Notes, my old WordPress blog. Notes is gone now, but I figured I would preserve this guide in case it helps someone out in the future.*

Tested on my HP EliteBook 8470p. I’ve heard it being used on other 8x70 EliteBooks as well, and there may be other devices with which it works.

1. Install the HP driver. [sp81058](https://ftp.hp.com/pub/softpaq/sp81001-81500/sp81058.exe)
2. Register a fingerprint, then press Win+L to lock your computer and check if it unlocks consistently. If it does, you’re done. If not, continue.
3. Open Device Manager, and under Biometric Devices, right click the fingerprint scanner (003d). Uninstall the device and “delete the driver files”. It doesn’t actually forget anything so I don’t know why it says that.
4. Restart, then open Device Manager and repeat step 3. Now open the Action menu and scan for new hardware. It should come up as an “Unknown Device”.
5. Now download this CAB archive from Microsoft’s update servers and extract it: [VFS491](http://www.catalog.update.microsoft.com/Search.aspx?q=VFS491) While extracting, 7-Zip may complain about there being some data after the payload. I don’t know what that is, but I ignored it and everything worked.
6. Now right-click the Unknown Device in Device Manager, click “Update Driver”, then browse your computer for a driver, say that you want to pick a driver from a list, click the Have Disk button, and select the INF file that was in the CAB archive from the last step.
7. The driver will be installed but Windows Hello, at this point, will not work; check your Settings just to be sure. Uninstall the driver you just installed. Now open the Update Driver box again, browse your computer again, say that you want to pick from a list again, and this time, in the list, choose the Synaptics driver under Biometric Devices. This is the same driver you uninstalled in Step 3. Install it.
8. Now register a fingerprint again. Windows Hello works flawlessly; better than my phone, and that sensor is a few years newer.

I can’t wait to switch to Linux so I don’t have to deal with this stuff ever again. Sadly, the only Linux driver for my fingerprint sensor is available only for Gentoo and Arch, and I’ve had even less success getting that to work than this.
