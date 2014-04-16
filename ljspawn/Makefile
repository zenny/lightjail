CC ?= cc
CFLAGS += -g -Ofast -fPIC -fPIE -Wall -Wextra -Wformat-security -fstack-protector-all -fsanitize=bounds
RONN ?= ronn
PREFIX ?= /usr/local

all: ljspawn ljspawn.1

install: all
	install ljspawn $(PREFIX)/bin/
	install ljspawn.1 $(PREFIX)/man/man1/

clean:
	rm -f ljspawn ljspawn.1 ljspawn.1.html

ljspawn: ljspawn.c
	$(CC) $(CFLAGS) -o ljspawn ljspawn.c

ljspawn.1: ljspawn.1.ronn
	$(RONN) --organization="Greg V" -r ljspawn.1.ronn

.PHONY: all install clean
