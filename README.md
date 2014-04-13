# lightjail

lightjail is an open-source modular system that allows you to use FreeBSD jails as lightweight containers for things like PaaS, CI, secure sandboxing, development environments.

Right now, the only thing that has been written is ljspawn, but there are plans and ideas...

## ljspawn

At the core of lightjail is the very lightweight `ljspawn` tool.
It allows you to use FreeBSD's `jail` syscall without using the stock FreeBSD jail infrastructure, which was designed for persistent virtual environments, not one-off ephemeral processes.
