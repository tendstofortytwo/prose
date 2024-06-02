---
title: Fixing sleep on Linux on a Dell Latitude 7290/7390/7490
summary: I cannot believe Linux sleep does not Just Work™ in Year Of Our Linux Desktop 2024
time: 1717293444
---

I got `nickel`, a Dell Latitude 7290, recently to replace `manganese`, my old Lenovo ThinkPad X250, and so far the greatest laptop of all time. The Dell is working reasonably well, though it needs a few hardware upgrades once I can afford them. In the meantime, I set up my usual installation of Arch Linux with a side of The Entire KDE Suite (Including The Kitchen Sink).

Everything works expectedly great, except, annoyingly, sleep. When I try to put the laptop to sleep, the screen turns black, the laptop does not shut off, and the caps lock light starts blinking (which I learned means "kernel panic"). A bunch of solutions I tried, on the Arch wiki and on random forums, did not work, and updating the BIOS did not do anything either (though I got some critial security updates so that's nice I guess). What did work was this answer, tested by *lamargo* and *davze* on the Arch Linux forums: https://bbs.archlinux.org/viewtopic.php?pid=1902330#p1902330.

In a nutshell, you need to add the following parameters to your kernel's cmdline:

```
acpi_enforce_resources=lax i915.enable_dc=0
```

In my case, this involved adding these flags to `GRUB_LINUX_CMDLINE_DEFAULT` in `/etc/default/grub`, then regenerating `grub.cfg` with `grub-mkconfig -o /boot/grub/grub.cfg`. These instructions might vary according to your distro.

Posting this pretty much so that I know what to do once I inevitably reinstall the OS on this machine. :D
