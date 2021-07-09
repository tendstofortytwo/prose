---
title: Running Windows 11 without Secure Boot or TPM
summary: I'm not saying it's a good idea, I'm just saying it can be done.
time: 1625805418
---

New Windows just dropped! Windows 11 looks pretty exciting to me -- the new interface is really pretty (fight me), the snapping windows should be an excellent productivity boost, and running Android apps on Windows sounds like a really interesting idea (especially since Chromebooks can do that as well). These things are probably not enough to sway me over from running Linux full-time, but I still want to try them out.

Sadly, trying them out is a bit of a challenge. I don't want to install Windows on my main laptop, mostly because of storage concerns. But none of my other machines are recent enough to support Windows 11. They don't have CPUs on the supported CPU list (which, as of writing, is 8th generation Intel or newer, and 2nd generation AMD Ryzen or newer), they don't support Secure Boot, and they don't have support for a Trusted Platform Module (TPM). All of these are requirements for Windows 11. The CPU can be skirted around by using a Windows Insider dev build, which I'm open to, but the other two requirements are still show-stoppers for me.

The machine of choice was a Lenovo ThinkPad x201e, with an AMD E-350 dual core processor, 4GB of RAM, and a 320GB spinning hard drive. It might surprise you to know that the hard drive is not the slowest component here -- the machine is CPU-bottlenecked regularly, even on tasks that would normally be I/O-bound on a normal Windows computer, like installing updates. Given all this, I was not super optimistic about my first plan of action -- running Windows in a virtual machine.

You see, [Linus Tech Tips](https://www.youtube.com/watch?v=odZSCdNTFPw) ran into compatibility issues as well, when trying out a leaked build of Windows 11. The solution was they found was to run it in a virtual machine. It seemed that if Windows setup detected you to be running in one, it would simply skip the required compatibility checks. Fine by me!

So I needed a virtual machine running. That means running a host operating system. I looked around for something as light as possible, to maximize the amount of resources available to the VM. I tried to use Alpine Linux first, but ran into weird hitching/stuttering issues that I couldn't quickly diagnose. After mucking about with the old `fglrx` proprietary driver for the ATi graphics for a while without much success, I gave up and switched to another distro.

Debian Xfce installed and worked without any fuss, and from there I was able to install `virt-manager` and `libvirt` to setup a virtual machine. It didn't recognize Windows 11 as an operating system (understandable) but claimed that Windows 10 was an end-of-life OS (???), which led to some confusion because it hid EOL OSs by default. After finding that toggle, the virtual machine was fairly easy to setup.

Now to install Windows 11. My first instinct was to install Windows 10 and try to upgrade from there. I popped a Windows 10 ISO into the virtual machine, and started installing... and holy hell was the installer slow. I was able to finish the first phase of the setup, the one which used the ISO, but even after hours of waiting, the second stage of the setup -- the part where you boot from the hard drive, it installs hardware, and starts and out-of-box setup -- simply would not finish. sigh.

At this point, I remembered that I had had better experience with using VMWare for virtual machines on Linux in the past, especially in cases where GPU passthrough wasn't a requirement. So I installed the free version of VMWare Player, and tried again. It was still slow, but much more tolerable. And I was finally able to get past setup onto the Windows 10 desktop. I installed VMWare Tools, signed into my Microsoft account (ugh) to enable Windows Insider builds, and tried to setup the Dev channel, which it wouldn't let me. I also tried to run Microsoft's official PC Health Check app (which has since been taken down), which showed me the following screen:

<figure>
<img src="/img/windows11-unsupported/win10-said-no.jpg" class="vertical">
<figcaption>PC Health Check tells me I cannot run Windows 11 because I do not have Secure Boot. Overlaid text on the image says, "damn."</figcaption>
</figure>

*(This image was taken from an Instagram story I posted, hence the text on the image. I like to post about the stuff I do there sometimes.)*

Clearly, upgrading to Windows 11 wasn't going to work. But Linus Tech Tips had done a fresh install. How about that?

After some looking around on Microsoft's website, it was clear that they hadn't yet released a full ISO for Windows 11 yet either. After some looking around *off* Microsoft's website, I found [UUP Dump](https://uupdump.net/), a service which generates a shell script for you to download Windows update packs from Microsoft update servers and package them into an ISO. 

Big disclaimer: I'm pretty sure this service is in a legal grey area at best, and I can't recommend you to run arbitrary code from an untrusted source without at least reading it. I read the script, it looked okay to me, and this laptop was network-isolated from my other devices and didn't have any important data on it. So I used it, and got a Windows 11 build 22000.1 ISO. Again, this does not mean you should do this. See the subtitle of this article on my stance on whether you should do this.

Anyway, I had a Windows 11 build 22000.1 ISO. I set my virtual machine to boot off that, and off we went. No compatibility problems! It just worked. For some definition of 'working' anyway. While I was able to install Windows 11 and use it, it was still very, very slow. And I was kinda annoyed. Clearly Secure Boot, TPM, and the new processors, weren't yet required. (Well, maybe additional CPU performance would help, but you can get adequate CPU *way* before 8th gen Intel Core/2nd gen AMD Ryzen.) There had to be a way to get this working on bare metal.

Turns out, there is! [Michael MJD on YouTube](https://www.youtube.com/watch?v=5rDJyMXbPdE) was able to install the leaked Windows 11 build on real hardware by simply replacing the contents of the `sources/` folder of the Windows 11 ISO (which contains the pre-boot environment and setup) with the stuff from a Windows 10 ISO. So I went ahead and replicated that -- made a bootable Windows 11 USB, then copied over the contents of the `sources/` folder (minus `install.esd/win/swm`, which contains the actual Windows image to be installed) from a Windows 10 ISO. And booting that onto the laptop itself, I was able to install Windows 11 with no problems, replacing Debian (a straight downgrade in all respects). 

<figure>
<img src="/img/windows11-unsupported/win11-on-x120e.jpg" class="vertical">
<figcaption>A follow-up to my previous Instagram story. Where the previous one seemed to indicate I couldn't run Windows 11, this one shows me running it with no problems. Overlaid text on the image says, "Apparently rules are fake and I get what I want. :D"</figcaption>
</figure>

I found it funny (and also convenient) that despite all my hardware technically being 'unsupported', Windows 11 was able to track down the right drivers from Windows Update and install them for me. I was especially surprised at the GPU drivers, since the Radeon HD 6310 requires an older driver version -- written specifically of AMD Fusion-powered laptops, none of which would meet the minimum hardware requirements for Windows 11. My best guess is that it's simply pulling in Windows 10 drivers for now, and I don't expect this to last into the final version of Windows 11.

Now, Windows setup does not by default set your operating system to receive beta builds via Windows Insider (yet). So I was in the unique position of running a build of Windows that was only avaiable on the Dev channel of Windows Insider, but having the Dev channel itself disabled. This led to the following screen on trying to enable it, which I found hilarious:

<figure>
<blockquote class="twitter-tweet" data-dnt="true" data-theme="light"><p lang="en" dir="ltr">Go home, Windows 11. You&#39;re drunk. <a href="https://t.co/aYoZjeqFjf">pic.twitter.com/aYoZjeqFjf</a></p>&mdash; Naman Sood (@tendstofortytwo) <a href="https://twitter.com/tendstofortytwo/status/1411480690449960960?ref_src=twsrc%5Etfw">July 4, 2021</a></blockquote> <script async src="https://platform.twitter.com/widgets.js" charset="utf-8"></script> 
<figcaption>A tweet, showing an image of a Settings app running on Windows 11, telling me I can't upgrade to Windows 11 since I don't meet the minimum hardware requirements.</figcaption>
</figure>

Despite that, I was actually able to enable the Dev channel, and upgrade to build 22000.51. If you read a launch-day review of the Windows 11 dev preview, it was most likely 22000.51. I'm not sure why UUP Dump made an ISO of 22000.1 instead. There were a few differences in the two builds, mostly in terms of UI features not being enabled -- 22000.1 still had the Ribbon UI for Windows Explorer, still had the old Windows 10-style Settings app (as can be seen in the tweet above), and a couple of context menus didn't look as pretty as they did for 22000.51. It was all fairly minor though, and I was able to upgrade to the new build without much issue.

Now, you might be thinking that since I was now running on bare metal, performance might be acceptable. If so, you would actually be wrong here. The AMD E-350 processor simply cannot handle modern Windows. It idles at 50-60% CPU usage, and shoots up to 100% if you do something as simple as opening the Start menu (or even doing nothing at all -- if Windows Updates run in the background, you will idle at 100% CPU usage and have no ability to do actual work). Because of this, I was not able to explore Windows 11 as much as I would have liked. Given that this is a CPU bottleneck, I can probably install it in a virtual machine on my main laptop and get decent performance, but that's annoying because it slows down the rest of the laptop and consumes battery. Oh well.

With that said, I don't think my experience represents the 'average' incompatible-with-Windows-11 PC. The problems I described in the paragraph above also exist in Windows 10, (and yet Windows 10 is not incompatible with the machine). That leads me to believe that on a PC where Windows 10 works fine, Windows 11 would probably work fine as well. As an example, radon, the desktop computer sitting at my home in India, has a Core i5 4440, 8 GB RAM, and a GTX 960 graphics card. It runs Windows 10 just fine, and will most likely also run Windows 11 just fine, if not for the 'minimum hardware requirements'. It meets the Secure Boot requirement, but not having TPM or being >=8th-gen. I can understand the security benefits of having the first two requirements, but the third one kinda baffles me.

But hey, I'm not Microsoft. They're probably smarter than I am. If they say these are the minimum hardware requirements, there's probably a good reason for that. After all, it's not like we can show that Windows 11 is perfectly capable of running on computers that don't meet these 'minimum' specifications -- oh, wait.

I kid, but that's just true today. I can definitely see features coming down the road that might make older CPUs incompatible -- like using some sort of a new feature or instruction set only available on these CPUs, or wanting to simplify their codebase by having to keep track of fewer hardware variations and features. And that's why even though I was able to get Windows 11 working on an 'incompatible' device now, I don't think this possibility will continue to exist, and I definitely wouldn't recommend using this method to upgrade one of your computers -- unless, of course, you just want to mess around like me.

Upgrade to Linux instead. :P
