---
title: Fixing Bluetooth audio issues with Jabra Elite Active 75t and Ubuntu 21.04
summary: Bluetooth audio on Linux is a nightmare.
time: 1631387096
---

I bought the Jabra Elite Active 75t because it had support for Bluetooth multipoint -- basically, the ability to simultaneously connect your earphones to two separate devices and be able to listen to audio on either of them (one at a time). However, I haven't been able to actually use this functionality for the vast majority of time I have owned these earphones, simply because pulseuadio doesn't let me.

When you connect your earphones via Bluetooth, it (usually) connects in one of two common profiles -- either it is A2DP (stereo audio, no mic support, multipoint works), or HSP/HFP (mono audio, mic support, multipoint doesn't work). Ideally, I want my earphones to stay in A2DP, since I want high quality audio and multipoint, and don't really care about using their internal mic. However, whenever a VoIP program is running on my computer and I connect my earphones, PulseAudio notices this, thinks that I need a mic, and connects the earphones in HSP mode automatically.

This is not ideal, since Discord is considered to be a "VoIP program" and I nearly always have that running. I've been trying to fix this issue for ages at this point, but haven't been able to. Finally, however, the following combination of things worked:

1. Install Pulseaudio Volume Control (`pavucontrol` in repositories). Use this tool to set your earphones to `High Fidelity Playback (A2DP Sink)` under the Configureation tab.

2. In `/etc/pulse/default.pa`, find this block of text:

        .ifexists module-bluetooth-policy.so
        load-module module-bluetooth-policy
        .endif

    and edit the second line to have a space, followed by `auto_switch=false`, at the end. I found this step on the [ArchWiki](https://wiki.archlinux.org/title/Bluetooth_headset#Disable_auto_switching_headset_to_HSP/HFP).

3. In `/etc/bluetooth/audio.conf`, add the following block of text:

        [General]
        Enable=Source,Sink,Media,Socket

    I found this step on [this Medium post](https://medium.com/@overcode/fixing-bluetooth-in-ubuntu-pop-os-18-04-d4b8dbf7ddd6).

4. Restart the services responsible:

        sudo systemctl restart bluetooth
        pulseaudio -k

And now, even with Discord running, earphones should automatically connect in A2DP.

You might, however, experience a faint hissing sound in your left ear. This supposedly a [known problem](https://www.reddit.com/r/headphones/comments/9t6s5t/jabra_elite_active_65t_left_earbud_hissing_past_a/) in the Jabra 65t line, and it affects my 75t as well. The solution is to disconnect system volume and earphones volume, turn the earphones volume down to zero, and the system volume up to whatever amount you want. I guess this means I can't use my earphones to control volume anymore. While this is annoying, I think I can live without that functionality.
