CC ?= cc
CFLAGS += -g -Ofast
CFLAGS += -Wall -Wextra -Wconversion -Wno-unused-parameter -Wstrict-overflow -Wformat-security -Wformat=2 -Wno-format-nonliteral
CFLAGS += -fstack-protector-all -fsanitize=bounds
CFLAGS += -fPIC -fPIE
LDFLAGS += -pie -Wl,z,noexecstack -Wl,z,noexecheap -Wl,z,now -Wl,-z,relro
PREFIX ?= /usr/local

all: ljspawn ljspawn.1

install: all
	install ljspawn $(PREFIX)/bin/
	chmod u+s $(PREFIX)/bin/ljspawn
	install ljspawn.1 $(PREFIX)/man/man1/

clean:
	rm -f ljspawn ljspawn.1 ljspawn.1.html

ljspawn: ljspawn.c
	$(CC) $(CFLAGS) -o ljspawn ljspawn.c

.PHONY: all install clean
