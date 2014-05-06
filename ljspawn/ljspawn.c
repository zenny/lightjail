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
#include <sys/mman.h>
#include <sys/file.h>
#include <arpa/inet.h>
#include <pwd.h>
#include <signal.h>
#include <stdlib.h>
#include <stdarg.h>
#include <stdio.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>
#include <stdbool.h>

#define DEFAULT_NAME "lj"
#define DEFAULT_IFACE "lo0"
#define safe_snprintf(dst, len, ...) if ((snprintf((dst), (len), __VA_ARGS__)) >= (len)) die("Error: overflow!")

char *dest = NULL;
char *net_if = DEFAULT_IFACE;
char *ip_s = NULL;
struct in_addr ip;
char *name = DEFAULT_NAME;
char *proc = "echo 'Hello world'";
bool nobody = false;
pid_t p_fpid;

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
  kill(p_fpid, SIGTERM);
}

void handle_sigterm() {
  kill(p_fpid, SIGKILL);
}

void usage(char *pname) {
  printf("%s -- spawn a process in a lightweight jail environment\nusage: %s [-i <IP-address>] [-f <network-interface>] [-n <name>] [-0] -d <directory> -p <process>\nsee man %s for more info\n", pname, pname, pname);
}

void parse_options(int argc, char *argv[]) {
  int c;
  while ((c = getopt(argc, argv, "d:f:hi:0n:p:?")) != -1) {
    switch (c) {
      case 'd': dest   = optarg; break;
      case 'i': ip_s   = optarg; break;
      case 'f': net_if = optarg; break;
      case '0': nobody = true;   break;
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

static int *jail_id;

void wait_and_bleed(pid_t fpid) {
  signal(SIGQUIT, handle_sigint);
  signal(SIGHUP, handle_sigint);
  signal(SIGINT, handle_sigint);
  signal(SIGTERM, handle_sigterm);
  int status;
  waitpid(fpid, &status, 0);
  llog("Process exited with status %d", status);
  jail_remove(*jail_id); // Make sure there are no orphans in the jail
  munmap(jail_id, sizeof *jail_id);
  if (ip_s != NULL) execv("/sbin/ifconfig", (char *[]){ "ifconfig", net_if, "-alias", ip_s, 0 });
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
    pid_t ifconfig_pid = fork();
    if (ifconfig_pid == -1) die_errno("Could not fork");
    if (ifconfig_pid <= 0) { // Child
      return execv("/sbin/ifconfig", (char *[]){ "ifconfig", net_if, "alias", ip_s, 0 });
    }
  }
  int jresult = jail(&j);
  if (jresult == -1) die_errno("Could not start jail");
  *jail_id = jresult;
  chdir("/");
  llog("Running container %s in jail %d", dest, jresult);
  if (nobody) {
    system("pw useradd -n nobody -d /nonexistent -s /usr/sbin/nologin 2> /dev/null");
    struct passwd* pw = getpwnam("nobody");
    setgroups(1, &pw->pw_gid);
    setuid(pw->pw_uid);
  }
  return execve("/bin/sh", (char *[]){ "sh", "-c", proc, 0 },
      (char *[]){ "PATH=/usr/local/bin:/usr/local/sbin:/usr/games:/usr/bin:/usr/sbin:/bin:/sbin",
                  "LC_ALL=en_US.UTF-8",
                  "LANG=en_US.UTF-8",
                  "SHELL=/bin/sh", 0 });
}

int main(int argc, char *argv[]) {
  parse_options(argc, argv);
  llog("Going to start the command '%s' in %s\n", proc, dest);
  jail_id = (int*) mmap(NULL, sizeof *jail_id, PROT_READ | PROT_WRITE, MAP_SHARED | MAP_ANONYMOUS, -1, 0);
  if (jail_id == MAP_FAILED) die_errno("mmap failed");
  *jail_id = -1;
  p_fpid = fork();
  if (p_fpid == -1) die_errno("Could not fork");
  if (p_fpid > 0) { // Parent
    wait_and_bleed(p_fpid);
  } else { // Child
    return run();
  }
}
