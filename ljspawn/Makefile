CC ?= cc
CFLAGS = -pipe
CFLAGS += -Ofast
#CFLAGS += -g -O0
CFLAGS += -Wall -Wextra -Wconversion -Wno-unused-parameter -Wstrict-overflow -Wformat-security -Wformat=2 -Wno-format-nonliteral
CFLAGS += -fstack-protector-all -fsanitize=bounds
CFLAGS += -fPIC -fPIE
LDFLAGS += -pie
PREFIX ?= /usr/local

all: ljspawn ljspawn.1

install: all
	install ljspawn $(PREFIX)/bin/
	install ljspawn.1 $(PREFIX)/man/man1/

clean:
	rm -f ljspawn ljspawn.1 ljspawn.1.html

ljspawn: ljspawn.c logging.c
	$(CC) $(CFLAGS) -o ljspawn ljspawn.c $(LDFLAGS)

.PHONY: all install clean
