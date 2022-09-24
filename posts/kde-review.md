---
title: KDE review
summary: Your next desktop environment (but not mine).
time: 1663978914
---

_Note: This article was originally written from [mathNEWS](https://mathnews.uwaterloo.ca) as part of my Operating System Review series, but I decided to publish it here instead because this is not an operating system, and mathNEWS may not be as conducive to longer articles as it used to be._

I write this article from a copy of Fedora 36 KDE installed on my main laptop. It's great, it's super intuitive, it's mostly seamless, I really like it.

I'm going back to GNOME.

Some background: I used to run Fedora 36 GNOME on this laptop until today. Recently, I saw a really good price on an SSD upgrade, and I decided to jump on it. Changing out the main drive in my laptop gave me an interesting option - rather than to do what everyone does and clone the old drive onto the new one, I could instead install a new Linux distro, see how I like it, and potentially switch to it full time (and manually copy all my data over from the other drive).

But when the time came for a distro-hop, I realized that I really liked the stability of Fedora. I did want to try out something new though, and a few friends of mine had said that KDE is actually really good these days, so I decided to try the Fedora 36 KDE spin. I loaded the ISO onto a USB, swapped out my SSD, and booted into the live environment.

First impressions? It feels exactly like Windows 10/11! There's a black taskbar at the bottom, with big icons for running apps, some apps are pinned, there's a notification shade on the side, there's a Fedora logo "Start button" that opens up a menu extremely similar looking to a Windows 11 menu... I think a Windows user would feel right at home here. In fact, if I installed Fedora 36 on my dad's office PC tomorrow and somehow made his bespoke accounting software work, I don't think he'd notice a difference.

I quickly installed Fedora onto the new SSD and rebooted into the new environment. Initial setup was not a problem at all - installing the NVIDIA drivers, enabling RPM Fusion, copying data over from my old drive, all worked flawlessly.

KDE's UI has a lot of elements that I wish GNOME did. A menu where all external USB devices are listed and you can individually mount/unmount each volume, a sound menu that gives you easy access to per-application volume control, a GUI software center that actually works, a really well-designed Launchpad-style application launcher bound to the exactly the keybind that I keep my 3rd party Launchpad-style application launcher bound to.... suffice to say, there are a whole bunch of amazing things here, and that's why I say in the title that KDE could very well be your next desktop environment.

So why am I going back?

I have this very... esoteric fear. Most people probably don't even think about it, but it informs every software decision I make. It is fear of becoming used to something that might go away. Some examples of this fear showing up in my daily life:

* My vimrc is four lines long, all features I make sure I can live without, because I don't want to use vim on a new/unknown computer and be stuck without my custom keybinds
* I tried as hard as possible to not rely on any Gmail-specific features for my email needs, and because of that I was able to switch email providers to Fastmail on a whim when it was more convenient for me
* I only use Linux distributions on my main computer that are known to be popular and stable; so I would use Fedora, Ubuntu, maybe Arch for its large community, but never Alpine or Nix or Devuan - I don't want to be stuck in a situation where my choice of distribution locks me out of software that I need to use

So I'm very, very careful not to rely on specific software. And yet, that's the exact reason why I'm sticking with GNOME now. Despite all my precautions, there are little bits and pieces of my workflow that are too integrated to GNOME that I'm now used to and KDE cannot provide. For example:

* I use a drop-down terminal. The way this works is that there is a hotkey you can press (in my case Ctrl+`), that would drop a terminal in front of you, overlaying it on top of all other windows. If you do some work in the terminal and then press the hotkey again, it goes away, and if you press it a third time, it comes back to front, retaining all your state. I am extremely used to using drop-down terminals now and can't live without them.<br><br>On GNOME I use Guake, on macOS I use iTerm2 (it has a dropdown mode), and on KDE I should be able to use Yakuake, the KDE dropdown terminal. However, Yakuake does not let you change font sizes for the terminal with a keyboard shortcut, which is something that I do multiple times every single day. So Yakuake doesn't work for me. Guake does work on KDE, but it's really slow to activate and opens on random parts of the screen rather than on the top of the screen as it should. So Guake is not usable for me either. I tried other alternatives too, like Xfce's terminal, but ran into the same issues as Guake.<br><br>
* I use trackpad gestures extremely heavily to change between workspaces and switch apps. GNOME has first-class support for these, and the gestures are customizable using a program called Touchegg. KDE only very recently got trackpad gestures, they are not customizable yet, and the defaults are very weird - they seem to treat a gesture as a hotkey, rather than a "gesture" indicating the movement that you want. So while GNOME has "three finger swipe up" to show all windows on current window and "three finger swipe down" to hide them again, whereas KDE would map "swipe up" and "swipe down" to completely different things, like showing your workspaces and showing your windows on the current workspace. This feels completely unintuitive to me, and as I said, there's no way to change it.

There are a few other small things, but these two were the main issues that kept me from switching to KDE full-time. Notice that you probably do not care about these things - I have seen very few people use trackpad gestures as much as I do, and the vast majority of people - even power users! - don't even know what a drop-down terminal is. I just have a very weird setup and that prevents me from being as productive as I normally am with KDE. But that doesn't mean KDE isn't good - KDE is very good, as I established above. But sometimes it's just an issue with culture fit, I suppose.

Overall, highly recommend trying it out, but it isn't for me.
