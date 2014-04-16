// ljspawn -- spawn a process in a lightweight jail environment
// 
// Copyright 2014 Greg V <floatboth@me.com>
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions
// and limitations under the License.

#include <sys/param.h>
#include <sys/jail.h>
#include <sys/wait.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <signal.h>
#include <stdlib.h>
#include <stdarg.h>
#include <stdio.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>

#define DEFAULT_NAME "lj"
#define DEFAULT_IFACE "lo0"
#define M_BUF 256
#define safe_snprintf(dst, len, ...) if ((snprintf((dst), (len), __VA_ARGS__)) >= (len)) die("Error: args are too long!")

char *dest = NULL;
char *net_if = DEFAULT_IFACE;
char *ip_s = NULL;
struct in_addr ip;
char *name = DEFAULT_NAME;
char *proc = "echo 'Hello world'";

void die(const char *format, ...) {
  va_list vargs;
  va_start(vargs, format);
  fprintf(stderr, "=[ljspawn]=> "); vfprintf(stderr, format, vargs); fprintf(stderr, "\n");
  exit(1);
}

void die_errno(const char *format, ...) {
  va_list vargs;
  va_start(vargs, format);
  fprintf(stderr, "=[ljspawn]=> "); vfprintf(stderr, format, vargs); fprintf(stderr, ". Error %d: %s\n", errno, strerror(errno));
  exit(errno);
}

void llog(const char* format, ...) {
  va_list vargs;
  va_start(vargs, format);
  printf("=[ljspawn]=> "); vprintf(format, vargs); printf("\n");
}

void handle_sigint() {
  // Just need to handle somehow, otherwise the finish/cleanup section is not called at all
}

void usage(char *pname) {
  puts("ljspawn -- spawn a process in a lightweight jail environment");
  printf("usage: %s [options]\noptions:\n", pname);
  puts("  -d /path/to/dir -- path to the temporary folder where you have your jail environment");
  puts("  -i IP.AD.DR.ESS -- the jail's IPv4 address (optional)");
  puts("  -f interface0 -- the network interface for aliasing the jail's IPv4 address to (optional, 'lo0' by default)");
  puts("  -n name -- the value for setting jailname and hostname (optional, 'lj' by default)");
  puts("  -p '/path/to/process args' -- the shell command to execute in the jail");
}

void parse_options(int argc, char *argv[]) {
  int c;
  while ((c = getopt(argc, argv, "d:f:hi:n:p:?")) != -1) {
    switch (c) {
      case 'd': dest   = optarg; break;
      case 'i': ip_s   = optarg; break;
      case 'f': net_if = optarg; break;
      case 'n': name   = optarg; break;
      case 'p': proc   = optarg; break;
      case '?':
      case 'h':
      default: usage(argv[0]); exit(1); break;
    }
  }
  if (dest == NULL) { usage(argv[0]); die("Arg -d (directory) not found"); }
  if (ip_s == NULL) llog("Warning: running without IP address (-i)");
  if (ip_s != NULL && inet_pton(AF_INET, ip_s, &ip) <= 0) die("Could not parse IP address %s", ip_s);
  if (name == DEFAULT_NAME) llog("Warning: running with default name, specify with -n");
}

void wait_and_bleed(pid_t fpid) {
  signal(SIGINT, handle_sigint);
  int status;
  waitpid(fpid, &status, 0);
  llog("Process exited with status %d", status);
}

int run() {
  struct jail j;
  j.version = JAIL_API_VERSION;
  j.path = dest;
  j.hostname = name;
  j.jailname = name;
  j.ip4s = 0;
  j.ip6s = 0;
  if (ip_s != NULL) {
    j.ip4s++;
    j.ip4 = malloc(sizeof(struct in_addr) * j.ip4s);
    j.ip4[0] = ip;
    char shellstr[M_BUF];
    safe_snprintf(shellstr, M_BUF, "ifconfig %s alias '%s'", net_if, ip_s);
    system(shellstr);
  }
  int jresult = jail(&j);
  if (jresult == -1) die_errno("Could not start jail");
  chdir("/");
  llog("Running container %s in jail %d", dest, jresult);
  system("echo 'nobody:*:65534:65534:Unprivileged user:/nonexistent:/usr/sbin/nologin' >> /etc/passwd");
  return execve("/bin/sh", (char *[]){ "sh", "-c", proc, 0 },
      (char *[]){ "PATH=/usr/local/bin:/usr/local/sbin:/usr/games:/usr/bin:/usr/sbin:/bin:/sbin",
                  "LC_ALL=en_US.UTF-8",
                  "LANG=en_US.UTF-8",
                  "SHELL=/bin/sh", 0 });
}

int main(int argc, char *argv[]) {
  parse_options(argc, argv);
  llog("Going to start the command '%s' in %s\n", proc, dest);
  pid_t fpid = fork();
  if (fpid == -1) die_errno("Could not fork");
  if (fpid > 0) { // Parent
    wait_and_bleed(fpid);
  } else { // Child
    return run();
  }
}
