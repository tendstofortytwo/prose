---
title: Trying to build a NAS that's as good as Synology
summary: feat. a person who has never actually used a Synology NAS
time: 1688317476
---

A few months ago, I finally decided to build myself a NAS. I was pushed to do this by a few different factors:

* I started regularly using two computers and realized that not being able to work on the stuff that's on the other computer is incredibly frustrating
* I realized that "it's on my laptop and possibly in my email somewhere" is probably not the ideal backup strategy
* A friend of mine upgraded his Synology NAS from a $300 unit with 256MB RAM, to a $500 unit with 512MB RAM, or something silly like that, prompting me to think, "Well, that's a stupidly large amount of money for not a lot of compute... I can do better with commodity hardware!"
* SSD prices crashed and [r/bapcsalescanada](https://reddit.com/r/bapcsalescanada) went wild trying to keep up with all the discounted drives
* My tax refund came in

And so with my fate sealed, I got to work.

### Hardware

I may be slightly addicted to buying used hardware. Given that, and my "I can do better" attitude to the Synology, I certainly wasn't going to buy new, expensive hardware. That said, I also didn't want to buy unreliable stuff, so I ended up going with an HP EliteDesk 705 G3, that I bought on Kijiji for $50[^1].

It came with AMD Generic Pre-Ryzen Quad Core APU, 8GB of RAM, and a Generic 500GB Spinning Hard Drive. It has three SATA ports, which would be less than what I needed for this purpose. So I bought some more RAM, three Crucial MX500 1TB SSDs for the actual storage part of the NAS, and a PCIe SATA card to plug in more drives. 

All in all, my costs were around $320 for the whole thing, storage included. Not bad, I think?

You may be wondering whether the 3TB of raw capacity is actually going to be the entire capacity of my NAS. My answer to that is: yes! I don't consume a lot of storage, so I don't mind this -- I have never even managed to fill up a terabyte across all my devices. That said, should I need more capacity, there's a couple of ports still free on that PCIe SATA card. And in the worst case that even that isn't enough, I'll cave and buy Real NAS Hard Drives that store 10TB or whatever, and use these as boot drives in computers. You can always use more SSDs. This is also why I went for SSDs over HDDs -- the price difference was about $10 per drive, and I wouldn't have been able to use those HDDs as boot drives if I had to.

### Software attempt 1 [failure]

Trying to pursue the same turnkey-ness as Synology, I looked for similar software solutions that you can install and expect to Just Work. I came across iX Systems's [TrueNAS](https://www.truenas.com/), which is supposed to be basically that. Great! I followed the instructions, set it up, and... rebooted my computer to realize that I don't actually have internet.

I'm currently a University of Waterloo student, living on campus, and it turns out that the Ethernet plugs in my room are broken. I hadn't realized this for months because all my other devices (including the desktop computer) connect to the internet over WiFi -- it's reliable and fast enough that I never felt like setting up Ethernet. 

Oh well. Just use WiFi, right? Turns out TrueNAS does not officially mention WiFi anywhere in their documentation, and community threads on the TrueNAS forums and Reddit will either [argue with you about your usecase being invalid](https://www.reddit.com/r/truenas/comments/sz06df/has_anyone_tried_to_install_wifi_on_truenas_scale/), or [tell you that it is unsupported without anything to back it up](https://www.truenas.com/community/threads/supported-wifi.49693/).

Fine. It seems that even if this were somehow supported, it'll be a pain. I won't bother. Instead, I set up my NAS using USB tethering, and submitted a ticket to IT to fix my Ethernet port.

As part of setting up, I tried to install [Tailscale](https://tailscale.com) on it, as I do on all my devices. Since university WiFi is shared with all university students and staff, they don't let devices talk to each other over LAN. So if I want my devices to be able to talk to my NAS (or each other), I need to connect them all some other way -- and Tailscale is the most convenient way that I know of. I've been using them since I [interned with them](https://prose.nsood.in/tailscale-github-action) in 2021, and they haven't let me down yet.

Since I'm using TrueNAS Scale, which is based on Debian, I expected to be able to just add Tailscale's PPA and `sudo apt install` my merry way. But iX had too many people messing around with Linux internals and then complaining, so they disabled `apt` in the shell and made any modifications to the system done via the shell explicitly unsupported.

The way around this is through "charts", which are TrueNAS's equivalent of apps, and best I can tell, run some sort of a Kubernetes cluster locally (?!) with each app inside a container. So I performed all the steps to enable Kubernetes, installed a third party charts repo ([TrueCharts](https://truecharts.org/)), and installed Tailscale from a mildly broken GUI with an overwhelming amount of options from there. It somehow worked, so I sighed and put the computer away until Ethernet worked.

Fast forward a few months, Ethernet was still broken. IT came in to test the port a couple times, determined that the problem was on their end somewhere, said they'd fix it in a week or so, and have not responded to my emails since that week passed. Oh well.

I'm currently an intern at the FreeBSD Foundation, and as part of that, I learned that you can use your FreeBSD system as a gateway for other computers connected through an Ethernet cable to access the internet through it. Could I just get internet through my work computer, conveniently sitting next to the NAS?

Yes, yes, I could. It was pretty easy[^2].

With that out of the way, my NAS finally had internet access! I tried to log into my NAS using the Tailscale IP address, and... nothing. I looked at the Tailscale admin panel, and the NAS showed as offline. That was annoying. I could ping IPs and domain names from the NAS, so internet and DNS worked. But not Tailscale...

At this point I realized that I had another computer -- the FreeBSD "router" -- that had a direct link to the NAS. So I typed into it the IP address that the DHCP server had given to the NAS, and the admin interface showed right up. I logged in, and after a bit of digging around, realized that Kubernetes had _somehow_ nuked itself and all the apps installed with it were gone.

I could troubleshoot it at this point... but I didn't do anything wrong to my knowledge, and if the system is _this_ fragile, and has had all the complications I dealt with above, I honestly don't want to deal with it. TrueNAS had to go.

### Software attempt 2 [success]

It seemed like turnkey solutions weren't really working for me, so I decided to just install my own operating system and set it up in a configuration that worked. The idea now was to install a standard server OS, setup my RAID system, and then point some kind of self-hosted cloud storage solution at it. Unfortunately that wouldn't be able to do hardware monitoring for me, but I'll cross that bridge when I get to it.

#### The operating system: FreeBSD

Since I was still feeling quite adventurous, I decided to go with FreeBSD 13.2 as the operating system. In my usage so far it seems quite reliable, simple to configure, should have great support for ZFS (the filesystem I intended to use for storage), and has _fantastic_ documentation. Have you read the [FreeBSD Handbook](https://docs.freebsd.org/en/books/handbook/)? It's hosted and maintained online like the Rust book, and it's a much more beginner-friendly introduction into the depths of a Unix-y OS than the Arch wiki or anything else I've seen on the Linux side. Highly recommend reading through it.

Installation was a breeze -- quite reminiscent of Debian's setup, if you've done that before. I installed the OS on the Generic Spinning Hard Drive, since I don't particularly care for the OS -- if it dies, I'll just reinstall it. The important bit is the data, and that would be stored in a ZFS pool consisting of a RAID-Z1 vdev that had my three SSDs in it.

#### The storage: ZFS w/ RAID-Z1

"A ZFS pool consisting of a RAID-Z1 vdev that had my three SSDs in it" is a lot of words, so to explain:

* One ZFS "system" is a pool. Different pools are completely independent and unrelated to each other, the pool is highest individual abstraction of ZFS (that I'm aware of).
* A pool consists of vdevs -- virtual devices, I would guess? Data is striped across vdevs and losing any one vdev means losing all the data in the pool. Data redundancy is at the vdev level, not the pool level.
* A vdev can be any one of:
    - A RAID-ZN array of drives, for some value of N. 
        - This means that any if you have K drives (for K > N) of the same size, the vdev is going to have capacity (K-N) times one drive, and it can tolerate N drives failing before losing all your data. My case is RAID-Z1 with 3 drives of 1TB each, so I have 2TB of storage and any one drive can fail and I would still have all my data.
    - A block device, such as a single physical drive or partition, or even a file mounted as one. This approach is not recommended for production, because you're putting the data in your entire pool (not just this drive!) in the hands of this drive. If it fails, you're done.[^3]

Once I made the pool, I can partition it up into volumes. By default a volume doesn't use a set amount of space, it just uses space from the pool as needed, but you can set "quotas" on volumes if you miss the good old days of fixed-size partitions. I just made two volumes, one for all my stuff and one for the SQL database.

SQL database? About that:

#### The interface: Nextcloud

Nextcloud is a PHP+SQL cloud storage solution -- you install it on your FAMP[^4] stack server, and it provides you an easy way to sync files across devices, access files on the go easily, and much more. Exactly what I need!

Despite Nextcloud really trying to push you to using Docker containers for a self-hosted setup, [they have instructions for setting it up on a bare Linux computer](https://docs.nextcloud.com/server/latest/admin_manual/installation/source_installation.html), which also work for a bare FreeBSD computer, pretty much unaltered. Just install the package `php82-whatever` whenever they tell you to install the PHP module for `whatever`, and you should be almost good to go.

Note the "almost". I did that, provided DB credentials and an admin username/password to Nextcloud, and was greeted with a white screen. Looking at the PHP logs and googling around a bit on DuckDuckGo, it turns out FreeBSD has some [further unlisted requirements](https://help.nextcloud.com/t/server-migration-issues/56609) in terms of PHP modules. Installing `php82-mysqli`, `php82-filter`, `php82-xsl` and `php82-wddx`, as mentioned in that link (and adapted for my PHP version), fixed it.

Once it was set up, Nextcloud was a breeze. I did go to the admin settings and do a whole bunch of "recommended" extra things that it told me to, like enabling caching, HTTPS, and some other things. Conveniently, it links to the documentation page for all the things it suggests. For HTTPS, since I only expose my NAS over Tailscale, I was able to generate an HTTPS certificate for use with Nextcloud using `tailscale cert`. If you expose your NAS over the web, you'll probably want to use Let's Encrypt or such. I just personally feel a bit icky about exposing a PHP web app over the public internet, so internal-network-only it is.

After all that, it was just a matter of setting up the Nextcloud mobile and desktop apps for syncing, and I was good to go! These apps are the main reason I went for Nextcloud over just setting up an NFS/SMB server -- I value the automatic syncing of camera photos from my phone, local access to schoolwork files across my various computers, and other such things, a lot more than just being able to remote into a server to occasionally access files as necessary.

### Conclusion: after a lot of work and troubleshooting and headbanging, It Just Works

I enjoyed this process a lot! I generally like configuring computers, the NAS has been pretty useful and reliable in the few weeks that I've been using it, and it did come in a lot cheaper than anything that Synology seems to sell. It did require a lot of manual intervention to setup (and occasionally requires more intervention still), so I would not call this a turnkey solution at the Synology level, but I'm quite happy with what I got.

Would I recommend that you do the same thing? Well, it works for me, but I am not a sysadmin and this is not backup advice. If you feel comfortable tinkering with second-hand hardware and open source software to craft your own solution, that's great! But if you want the peace of mind of, "Someone has figured everything out for me and I never have to worry about my data disappearing on me," without the effort of doing the figuring-out yourself, maybe Synology etc are worth the money.

### Epilogue: things to do

I still need offsite backups -- the NAS just lives under my desk for now. The idea for now is to buy Backblaze B2 or something, and setup incremental backups. 

I'm also still not sold 100% on FreeBSD, and may move to Linux (possibly NixOS or Fedora Silverblue so I can store my entire OS config in a git repo and truly never worry about the Generic Spinning Hard Drive dying on me). I should do this before my data grows too much, or after inventing and checking the veracity of my offsite backups. It works for now though.

Finally, once [RAID-Z expansion](https://freebsdfoundation.org/blog/raid-z-expansion-feature-for-zfs/) lands, I would like to be able to expand my array without destroying and recreating it (or more likely, adding another RAID-Z1 vdev).

But that's for the future! For now, I'm more backed up and peace-of-minded than ever before. Yay!

[^1]: All prices in this article are CAD.

[^2]: For those following along at home, I used the networking setup from [this guide](https://freebsdfoundation.org/wp-content/uploads/2014/07/Using-bhyve-for-FreeBSD-Development.pdf) and just hooked up my actual real-life Ethernet port instead of a virtual machine's TUN device.

[^3]: I wonder if this is just implemented as a RAID-Z0 internally. 🤔

[^4]: Haha, geddit, FreeBSD-Apache-MySQL-PHP? Like LAMP but with FreeBSD instead of Linux? Apparently this is the canonical name for this setup -- though some people substitute MariaDB for M, and Perl for P. I used the vanilla kind.
