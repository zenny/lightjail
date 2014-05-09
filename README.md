# lightjail

lightjail is a modular system for using FreeBSD jails as lightweight ephemeral containers.
Projects like PaaS, CI, secure sandboxing, development environments will find it useful.

## ljspawn

At the core of lightjail is the very lightweight `ljspawn` tool.
It allows you to use FreeBSD's `jail` syscall without using the stock FreeBSD jail infrastructure, which was designed for persistent virtual environments, not one-off ephemeral processes.
It's installed as a `setuid` binary, so you can run jails as non-root.
Hopefully, it's secure :-)

## ljbuild

`ljbuild` is the overlay build script runner for lightjail.
It takes care of mounting the world and the directory of what you're building (using `nullfs`/`unionfs`) and running the installation script using `ljspawn`.
The `Jailfile` build script format is simple:

```bash
# Example: Jailfile for installing Ruby
# Comments start with a # symbol, like in the shell
name lang/ruby # Required. This is the directory where it will be built. MUST contain a slash (subdirectory)
version 2.1.1 # Optional, a random string will be used by default
world 10.0-RELEASE # Techically optional, `uname -r` of the host will be used by default
from lang/common 1.0.0 # Optional. This is the parent overlay.
                       # Prefix search is used for the version.
                       # If you specify 2.1, it will use 2.1.1 if [2.1.0, 2.1.1] exist
copy / # Optional. Copies all files in the Jailfile's directory to the specified directory inside the overlay
---
# After the separator, we have a shell script
pkg install -y ruby21
pkg clean -aqy
```

You can run ljbuild as non-root:

1. add `vfs.usermount=1` to `/etc/sysctl.conf` and reload sysctl: `sysctl -f /etc/sysctl.conf`
2. chown the worlds directory and all worlds you have to the user who will run ljbuild
3. make sure `ljspawn` is `setuid` (the makefile does it)
4. run it!

There is a problem though -- building something as root makes it impossible to mount as non-root... so you can't build something `from` the container you built as non-root.

## The directory structure

There is a root directory for all lightjail related things.
It MUST be read from the `LIGHTJAIL_ROOT` environment variable by all `lj*` tools, `/usr/local/lj` MUST be used if the variable is empty.

Under the root, there MUST be a `worlds` directory.
Under `worlds`, directories MUST contain FreeBSD base system installations.
The directories MUST be named with whatever version they contain (result of `uname -r`.)
The installations MUST have ports and pkg configured.

The overlays MUST be stored under the root.
Each overlay MUST be stored in a subdirectory, eg. `namespace / name`.

### Example world installation

```bash
W=/usr/local/lj/worlds/$(uname -r)
IFACE=eth0 # your default gateway network interface, could also be vtnet0 or something
mkdir -p $W
cd /usr/src
make buildworld
make installworld DESTDIR=$W
make distribution DESTDIR=$W
cp /etc/resolv.conf $W/etc
mount -t devfs devfs $W/dev
ljspawn -f $IFACE -d $W -p 'portsnap fetch extract' 
ljspawn -f $IFACE -d $W -p 'make -C/usr/ports/ports-mgmt/pkg install clean' 
ljspawn -f $IFACE -d $W -p 'pkg' 
umount $W/dev
```

Running this as root should install a world to eg. `/usr/local/lj/worlds/10.0-RELEASE`.
You don't have to do this on every server.
Do it once, archive it with `cpio`, compress with `xz`, copy to all servers (*with the same CPU architecture, of course*) and extract there.

## Requirements

1. You must rebuild your kernel with RCTL ([14.13.2. Enabling and Configuring Resource Limits](http://www.freebsd.org/doc/handbook/security-resourcelimits.html#idp76631728)), otherwise limiting containers' RAM won't work
2. You must set up devfs rules:
  - `sudo cp /etc/defaults/devfs.rules /etc/devfs.rules`
  - `sudo sysrc devfs_load_rulesets="YES"`
  - `sudo service devfs restart`
