CC ?= cc
CFLAGS += -g -O3 -Wall -Wextra -Wformat-security -fstack-protector-all -fsanitize=bounds

ljspawn: ljspawn.c
	$(CC) $(CFLAGS) -o ljspawn ljspawn.c

install: ljspawn
	install ljspawn /usr/local/bin/

clean:
	rm ljspawn
