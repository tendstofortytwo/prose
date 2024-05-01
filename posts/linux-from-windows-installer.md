---
title: How to install Linux from a Windows installer
summary: I'm not sure why you would want to do this. Presumably for the same reason as me? I also don't know what that was.
time: 9999999999999
---

A few days ago, I made this relatively-popular post on my Fediverse account, which had a picture of a Windows installer offering to install Alpine Linux:

<iframe src="https://social.treehouse.systems/@tendstofortytwo/112352451918495313/embed" class="mastodon-embed" allowfullscreen="allowfullscreen"></iframe><script src="https://social.treehouse.systems/embed.js" async="async"></script>

And a day or two after that, I posted a YouTube video showing the entire install process from start to finish:

<iframe src="https://www.youtube.com/embed/mJwueWvZRG8?si=ykHkDfNNBdYmpSzm" class="youtube-embed" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

Countless[^1] people have suggested that I should write a blog post about how I did this, so here we go.

[^1]: two

### Step 1: Install Linux on an NTFS partition

This is actually pretty easy these days! You want to pick a Linux distribution that will let you mount partitions by hand and then uncritically install to those mountpoints without questioning your decisions. Two Linux distributions that I'm aware of that will let you do this are Arch Linux and Alpine Linux.

#### The partition scheme

This is approximately what the partitioning scheme looks like:

1. We are using UEFI/GPT. Presumably you could make this work with legacy boot/MBR, but that would require you to somehow install the grub boot sector from Windows. I don't know how to do that.

2. The first partition is an EFI partition, about 100MB in size. Your Linux distributions may want you to create bigger partitions, but the Windows installer will only create a 100MB partition so you're stuck with this size anyway. This is formatted as FAT32, and mounted at `/boot` the usual way.

3. The second partition is a 16MB partition that we don't need to use, but need to correctly offset the partition numbers, since the Windows installer will create a 16MB partition between the EFI partition and the system partition for some reason.

4. The third partition is your root filesystem, filling up the rest of the disk. This is formatted as NTFS, mounted at `/` using the `ntfs3` partition type and the `acl` and `windows_names` mount options. The `acl` flag enables the driver to use POSIX ACLs (think: file permissions), and the `windows_names` flag prevents the creation of files that would not work on a Windows computer (say, anything that contains a `:` or is named `con`).

#### Making the partitions and installing

Using the [Alpine semi-automatic installation guide](https://docs.alpinelinux.org/user-handbook/0.1a/Installing/manual.html) as a reference, you might create your partitions like so:

```
// ntfs-3g-progs is required for mkfs.ntfs, which does not ship on 
// the alpine live iso for some reason
# apk add ntfs-3g-progs parted
# alias p="parted -sa optimal /dev/sda"
# p mklabel gpt
# p mkpart p 0M 100M
# p mkpart p 100M 116M
# p mkpart p 116M 100%
# p set 1 esp
// required so Windows will detect this as a valid NTFS partition
// later
# p set 3 msftdata
# mkfs.vfat -F32 /dev/sda1
# mkfs.ntfs -Q /dev/sda3
# mount -t ntfs3 -o acl,windows_names /dev/sda3 /mnt
# mkdir /mnt/boot
# mount /dev/sda1 /mnt/boot
```

After that, you simply proceed with the installation as normal. At this point, if you're using Arch Linux, your `pacstrap` command will fail, because `pacman` expects to be able to create files with `:`s in the name, and as we discussed, that's illegal. So, uh, try not to use a Linux distro that relies on the ability to do that. The rest of this guide will assume you're installing Alpine.

#### Getting the system to boot

While Alpine correctly adds `rootfstype=ntfs3` to your kernel cmdline, for some reason, as of writing, the initramfs it generates does not include the `ntfs3` kernel module. Also, the Alpine installer will generate an `fstab` that uses UUIDs to identify partitions. Since we'll be relying on the partitions created by the Windows installer, which will have fresh UUIDs, we need to change the `fstab` to use the ol' unreliable drive names instead.

So after you're done installing, open `/mnt/etc/fstab` and change the first column of the EFI partition to say `/dev/sda1`, and the first column of the root partition to say `/dev/sda3`.

Next, we need to add `ntfs3` to the initramfs. For that, install the `mkinitfs` package on the live environment, create a file `/mnt/etc/mkinitfs/features.d/ntfs3.modules` that contains the text `kernel/fs/ntfs3`, and make sure that `/mnt/etc/mkinitfs/mkinitfs.conf`'s `features` line contains the `ntfs3` feature.

> Side note: At the time of writing, you need to open `/sbin/mkinitfs` in a text editor and prefix the `sysconfdir` variable at the top of the file with `/mnt`, since `mkinitfs`'s `-b` flag doesn't apply to that variable for some reason. [I've filed a bug report for this](https://gitlab.alpinelinux.org/alpine/mkinitfs/-/issues/60) so hopefully that can be fixed and we can all start installing Alpine Linux to NTFS partitions as ~~God~~ Gates intended.

After that, running `mkinitfs -b /mnt -f /mnt/etc/fstab 6.6.28-0-lts` should regenerate your initramfs in a way that works with our shenanigans. Replace the kernel version string with whatever's there in `/mnt/lib/modules`.

Finally, we need to tell grub to stop using UUIDs as well. Open `/mnt/boot/grub/grub.cfg`, find the Alpine Linux menu entry, replace the line `search ... --set root <UUID of the EFI partition>` with `set root='hd0,gpt1'` (that line may already exist, so let it be if so), and change the `linux ... root=UUID=<UUID of root partition> ...` line to say `root=/dev/sda3` instead.

There's probably a better way to do this by changing `/mnt/etc/default/grub` and regenerating the config using `grub-mkconfig` but I don't understand how that works and so this is what I did.

#### Behold the monster

Now, unmount `/mnt/boot` and `/mnt`, and you should be ready to reboot into your new Alpine installation! Unmounting is very important since if the dirty bit is set on an NTFS partition (which it is while it is mounted), `ntfs3` will refuse to mount it read-write until you run `chkdsk` on it from a Windows system or `ntfsfix` on it from a Linux system with `ntfs-3g-progs`. You do not want your root partition to be mounted read-only, it makes life hard.

After verifying that it does boot, we can move on to the hard part.

### Step 2: Stuff Linux into a WIM file

This section (and transitively, the rest of this article) was motivated by the following exchange on Discord about the Windows installer:

<figure>
<img src="/img/linux-from-windows-installer/discord-exchange.png" class="vertical" alt="A screenshot of a Discord conversation. chompxchg asks if you could shove a Linux distro into a WIM. Luna laughs. tendstofortytwo wonders if the USB boot stage of the installer does anything other than extracting the WIM, and if you could install Linux on an NTFS partition. chompxchg points out that at the very least the WIM could contain a Linux squashfs, and that nothing since the 2000s has used the Rock Ridge extensions. Luna notes that the Windows installer UI basically just runs DISM to extract the WIM file and bcdboot to make the partition bootable, on top of the actual disk partitioning." style="max-width: 30rem">
</figure>

For context, a WIM (Windows Image) file is a compressed storage format used by Microsoft in the Windows installer. Think of it as a fancy ZIP file that the installer extracts onto your hard drive. A single WIM file can contain multiple images, and the Windows installer can pick which image it wants to install, either based on your product key, or what edition you choose during setup.

Our goal, then, is to add an image to the Windows installer's WIM file (`sources\install.wim` in your Windows ISO/USB) whose contents are our Linux installation.

#### Laying out the plan

Creating a WIM file with our Linux installation is easy --- Microsoft themselves provide tools for you to create your own WIM files, using a tool called DISM[^2] (Deployment Image Service and Management tool). I've been using a convenient GUI frontend for DISM called [GImageX](https://www.autoitconsulting.com/site/software/gimagex/), but if you're a Microsoft purist you could just use DISM directly for the same effect.

[^2]: formerly ImageX

The hard part is making Linux bootable. `bcdboot` does not know what a Linux is. It only knows how to make Windows bootable. I thought of solving this problem by replacing `bcdboot` in the Windows setup with something that installs GRUB, but that felt against the spirit of the whole thing and also kinda hard. Instead, I came up with the following strategy:

1. Install a copy of the Windows Preinstallation Environment (WinPE) on the same rootfs as Linux, putting them both in the same WIM image
2. Have the Windows installer make WinPE bootable and the computer reboot into WinPE after the first phase of Windows setup
3. Have WinPE make the necessary adjustments to make Linux bootable, and reboot into Linux

#### Meet the Windows Preinstallation Environment

WinPE is made freely available by Microsoft as part of the Windows Assessment and Deployment Kit, and [Microsoft even provides instructions](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/winpe-intro?view=windows-11) on how to download WinPE, customize it for your purposes, and use it in your custom applications. The only thing they disallow you from doing is using it as a general-purpose OS, but that's fine, since we're not doing that.

> Side note: You will need to use Windows to follow the steps below, since the Windows ADK is not available for other platforms. You may have luck with obtaining a copy of WinPE through some other means and using `wimlib-imagex` to capture and apply your WIM files from within Linux, but I found that that tool would happily generate files from partitions that cause Windows to choke --- which are useless since we'll be using them in the Windows setup or in WinPE, both of which would also choke on them.

After installing the Windows ADK and the WinPE addons, open the Deployment and Imaging Tools Environment from the Start menu as an administrator, and run the following command to get a working copy of WinPE.

```
> copype amd64 c:\pe_amd64
```

This will dump a copy of a WIM file containing Windows PE in `C:\pe_amd64\media\sources\boot.wim`. Apply this image to the drive containing your Linux rootfs, either using the [`Dism /Apply-Image`](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/dism-image-management-command-line-options-s14?view=windows-11#apply-image) command or using the Apply tab in GImageX. The Linux rootfs drive is `E:\` in this screenshot:

<figure>
<img src="/img/linux-from-windows-installer/gimagex-apply-winpe.png" class="vertical" alt="GImageX open at the Apply tab, with the source pointing to the WinPE boot.wim and the destination as the E: drive">
</figure>

Now that we have WinPE, we can make it do our bidding.

#### Making Linux bootable

The great thing about UEFI boot is that there doesn't need to be any magic bootcode in the first 512 bytes of the hard drive. You *can* inject magic UEFI boot entries into the firmware (Windows will inject one called "Windows Boot Manager", and grub will inject one with whatever name you pass in the `--bootloader-id` flag while running `grub-install`), but you don't have to. If you boot from a hard drive directly, without the help of any magic boot entry, the UEFI will simply read `efi\boot\bootx64.efi` from the EFI partition in that drive and run it.

`bcdboot` will put a file there that will load Windows, in addition to creating the UEFI entry. And if you specify `--removable` while running `grub-install`, grub will also put a file there that will load grub and read `grub/grub.cfg` from the EFI partition. Conveniently, Alpine specifies `--removable`, so no magic UEFI entries to worry about creating.

Recall that when we installed Linux, we made `/boot` our EFI partition. So all we need to do to make Linux bootable from WinPE is to drop into the EFI partition the files that Alpine put in `/boot`. It'll drop the `efi\boot\bootx64.efi` in place, and that will make it so that booting from the hard drive will boot grub, which will load Alpine.

To do this, first, we need to save a copy of the Linux EFI partition. In Windows, run the `diskpart` command[^3], and run the following commands in the `diskpart` prompt:

[^3]: `diskpart` is like `fdisk` for Windows people

```
> list disk
// note N, the disk number of your Linux drive
> select disk N
> select partition 1
> assign letter=S
```

This will make the Linux EFI partition available as `S:\` on Windows. Once that's done, you can use the [`Dism /Capture-Image`](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/dism-image-management-command-line-options-s14?view=windows-11#capture-image) command or the "Capture" tab in GImageX to capture the `S:\` drive as `efi.wim` on your Linux rootfs.

<figure>
<img src="/img/linux-from-windows-installer/gimagex-capture-efi.png" class="vertical" alt="GImageX open at the Capture tab, with the source pointing to the S: drive and the destination as efi.wim in the E: drive">
<figcaption>The Name and Description fields are not important, for now.</figcaption>
</figure>

Now that we have the contents of the EFI partition, we just need to get WinPE to apply them. Conveniently, the default build of WinPE simply executes `\Windows\System32\startnet.cmd` on startup, so we can make that file do what we want.

Open `E:\Windows\System32\startnet.cmd` in Notepad as administrator. It should already have a line that says `wpeinit` in it --- as the name suggests, that initializes a WinPE environment. I have no idea if it's necessary, but it doesn't hurt so I just left it in there. Add the following lines:

```
diskpart /s mount-efi.txt
dism /Apply-Image /ImageFile:"C:\efi.wim" /Index:1 /ApplyDir:S:\
bcdedit /delete {bootmgr} /f
exit
```

The second line should be fairly self-explanatory --- we're applying the EFI image we captured to the `S:\` drive, which presumably is where our EFI partition is. The first line is what we use to put it there. By default, EFI partitions are not assigned a drive letter, but as we did above, you can use `diskpart` to give them one. The `/s` flag allows you to give a script to execute to `diskpart` rather than it reading commands from the terminal. Create a file called `mount-efi.txt` in the `System32` folder, and give it the following contents to mount the EFI partition:

```
select disk 0
select partition 1
assign letter=S
exit
```

The third line we added in `startnet.cmd` deletes the magic UEFI boot entry that the Windows installer creates to boot Windows. This entry has a higher precedence than booting directly from the hard drive, so we want to delete it to boot into grub like we want.

And that should be it! Now, booting into our Linux rootfs will start a copy of WinPE, which will render itself unbootable and make Linux bootable in its stead, then reboot.

#### One more thing

This took me three days of headbanging, trial and error, giving up, retrying, and more, to figure out. You could capture the Linux rootfs into a WIM file at this point, and put it into a Windows ISO and boot from it. The installer will appear to run successfully, until the installer asks you for a product key and you say you don't have one. Then you'll see this screen:

<figure>
<img src="/img/linux-from-windows-installer/windows-setup-eula-error.png" class="vertical" alt="The Windows installer, running in a virtual machine titled, 'Can a Windows ISO and a Linux Installation Really'. It's displaying an error message: 'Windows cannot find the Microsoft Software License terms. Make sure the installation sources are valid and restart the installation.'">
<figcaption>Typical Microsoft behavior.</figcaption>
</figure>

It turns out that the Windows installer will read the EULA out from within the WIM file, and from within the specific image you chose. It will look for this EULA in `\Windows\System32\[Locale]\Licenses\[Channel]\[Edition]\license.rtf`. The locale is the language and country code, like `en-US`, the channel is one of `_Default`, `OEM`, or `Volume`, and the `Edition` is set for the image by the software that captured your WIM file. I *thought* that this corresponded to the SKU dropdown in GImageX, but that seems to not be the case, it seems to instead be auto-detected by DISM. The channel for WinPE is, well, `WindowsPE`.

So download a copy of the [GNU General Public License v2.0 in RTF format](https://www.gnu.org/licenses/old-licenses/gpl-2.0.rtf), open it in WordPad to set the font to Segoe UI so that it renders properly, and then save it as `E:\Windows\System32\en-US\Licenses\WindowsPE\_Default\license.rtf`, where `E:` is your Linux rootfs. You may also have to populate the `OEM` and `Volume` directories --- I'm not sure which one the installer read, since I populated all three in my testing.

#### The capture

Now, we can use `Dism /Capture-Image` or GImageX's Capture tab to capture the Linux rootfs into your Windows setup. Copy the contents of a Windows ISO into a convenient folder (I used `C:\win10` since I was using a Windows 10 ISO), and capture your Linux rootfs drive into the `sources\install.wim` in that folder. Use the "append" mode if you want to keep the existing Windows images in the WIM file, or "create" mode if you want to overwrite the file entirely and only be able to install Linux.

<figure>
<img src="/img/linux-from-windows-installer/gimagex-capture-efi.png" class="vertical" alt="GImageX open at the Capture tab, with the source pointing to the E: drive and the destination as install.wim in the Windows installer directory. The name and description fields are all set to 'Alpine Linux'.">
<figcaption>This is where the Name and Description fields become important.</figcaption>
</figure>

And that's it! The Windows installer in that directory will now be able to install Alpine Linux. Or rather, it will be able to install a copy of Windows PE that incidentally has an entire installation of Alpine Linux embedded in it, and whose only job is to make that copy of Alpine Linux bootable. Same thing.

You can turn this folder into a bootable ISO using the `oscdimg.exe` command provided in the Windows ADK. Instructions for doing that are in [this wonderful ElevenForum post](https://www.elevenforum.com/t/create-custom-windows-11-iso-file.443/). I'm so glad I found that post, because I have no idea how anyone would be expected to figure out those magic value in that command. In case the forum post goes down, there's an archive of that URL in the Wayback Machine.

### Step 3: Profit?

At this point, you should be able to boot the ISO and install Alpine Linux using the Windows installer onto a blank hard drive. Note that I do not recommend using this installer for anything important, for various reasons:

* The installer is *incredibly* fragile. I basically hardcoded the disk and partition layouts into the Linux installation, as well as the `diskpart` commands that WinPE runs. This will only work for installing to the first hard drive in a computer when it is completely blank before the Windows installer touches it. And if the installer changes the partition layout it creates on a blank hard drive, it's over. For the record, the installer I used was the Windows 10 22H2 ISO from the Microsoft website.
* The WIM capture does not preserve UNIX file permissions, so while the installed copy of Linux boots, you will need to fix permissions for the entire installation afterwards somehow.
* The grub changes are very hacky and I don't know if they'll survive an update.
* Using NTFS for your Linux rootfs is honestly just a really bad time. As I mentioned before, if the dirty bit is set that will make your entire rootfs read-only, making normal use of the computer impossible until you boot from a live CD to clear it. Also, while the base Alpine install works, there's no guarantee that any software you install won't expect to be able to create files with names prohibited by Windows.
* You literally already installed Alpine the normal way as a sub-procedure of this. Just do that. Please.

But you have to admit, it's quite funny.

Please redirect any hate mail to `"spam\x40tends\x2eto"`. :)

