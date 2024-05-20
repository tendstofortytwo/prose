---
title: "FreeBSD pf: forward traffic from one interface to a particular server on another interface"
summary: I can't believe nobody ever made a post about how to port-forward with pf
time: 1716243156
---

Enable `pf`, if not already enabled, and also enable IP forwarding, by running these commands:

```
// enable the firewall
# sysrc pf_enable=YES
// optional: enable logging
# sysrc pflog_enable=YES
// enable IP forwarding for future boots
# sysrc gateway_enable=YES
# sysrc ipv6_gateway_enable=YES
// enable IP forwarding for current boot
# sysctl net.inet.ip.forwarding=1
# sysctl net.inet6.ip6.forwarding=1
```

Add the following code block to your `/etc/pf.conf`, setting:

* `ext_if` to the interface on which traffic will arrive
* `int_if` to the interface on which the server you want to forward traffic to is accessible
* `printer_ip` to the IP address of the server you want to forward the traffic to
* `printer_port` to the port number (or in this case, well-known protocol name) of the port you want to forward

```
ext_if = "tailscale0"
int_if = "bge0"
printer_ip = "192.168.0.35"
printer_port = "ipp"
# translate packets going out of int_if to the IP of int_if
nat on $int_if -> ($int_if)
# redirect TCP packets coming into ext_if on printer_port to printer_ip
rdr on $ext_if proto tcp from any to any port $printer_port -> $printer_ip port $printer_port
```

You may also have to add a `pass` rule if you block traffic by default. Reload the firewall rules by running:

```
# pfctl -F all -f /etc/pf.conf
```

And that should be it.

This guide is adapted from instructions in [this blog post](https://sporks.space/2021/02/15/redirecting-privileged-ports-to-unprivileged-ports-on-the-same-system-with-pf-on-freebsd/) and [this Server Fault answer](https://serverfault.com/a/792463).
