CFLAGS = -g -O3 -Wall -Wextra -Wformat-security -fstack-protector-all -fsanitize=bounds

ljspawn: ljspawn.c mounts.c
	cc $(CFLAGS) -o ljspawn ljspawn.c

install: ljspawn
	cp ljspawn /usr/local/bin/

clean:
	rm ljspawn
