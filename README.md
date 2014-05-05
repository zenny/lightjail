# lightjail

lightjail is an open-source modular system that allows you to use FreeBSD jails as lightweight containers for things like PaaS, CI, secure sandboxing, development environments.

## ljspawn

At the core of lightjail is the very lightweight `ljspawn` tool.
It allows you to use FreeBSD's `jail` syscall without using the stock FreeBSD jail infrastructure, which was designed for persistent virtual environments, not one-off ephemeral processes.

## ljbuild

`ljbuild` is the build script runner for lightjail.
It takes care of mounting the world and the directory of what you're building (using `nullfs`/`unionfs`) and running the installation script using `ljspawn`.
It will handle dependencies in the future.

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

```shell
W=/usr/local/lj/worlds/$(uname -r)
IP=192.168.1.11 # example
cd /usr/src
make buildworld
make installworld DESTDIR=$W
make distribution DESTDIR=$W
cp /etc/resolv.conf $W/etc
mount -t devfs devfs $W/dev
ljspawn -i $IP -d $W -p 'portsnap fetch extract' 
ljspawn -i $IP -d $W -p 'make -C/usr/ports/ports-mgmt/pkg install clean' 
ljspawn -i $IP -d $W -p 'pkg' 
umount $W/dev
```
