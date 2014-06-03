// ljspawn -- spawn a process in a lightweight jail environment

#include <sys/param.h>
#include <sys/jail.h>
#include <sys/wait.h>
#include <sys/socket.h>
#include <sys/mman.h>
#include <sys/file.h>
#include <sys/stat.h>
#include <sys/resource.h>
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
#include "logging.c"

#define DEFAULT_NAME "lj"
#define DEFAULT_IFACE "lo0"
#define safe_snprintf(dst, len, ...) if ((unsigned long)(snprintf((dst), (len), __VA_ARGS__)) >= (len)) die("message", "Buffer overflow")

char *dest = NULL;
char *net_if = DEFAULT_IFACE;
char *ip_s = NULL;
struct in_addr ip;
char *default_name = DEFAULT_NAME; // don't compare with literal
char *name = DEFAULT_NAME;
char *proc = "#!/bin/sh\necho 'Hello world'";
char *redir_stdout = "/dev/stdout";
char *redir_stderr = "/dev/stderr";
long long memory_limit_mb = 256;
bool nobody = false;
pid_t p_fpid;

void handle_sigint() {
  kill(p_fpid, SIGTERM);
}

void handle_sigterm() {
  kill(p_fpid, SIGKILL);
}

void usage(char *pname) {
  printf("%s -- spawn a process in a lightweight jail environment\nusage: %s [-i <ip-address>] [-f <network-interface>] [-n <name>] [-O <stdout>] [-E <stderr>] [-0] [-j] -d <directory> -p <process>\nsee man %s for more info\n", pname, pname, pname);
}

void parse_options(int argc, char *argv[]) {
  int c;
  while ((c = getopt(argc, argv, "0d:f:hi:jm:n:p:O:E:?")) != -1) {
    switch (c) {
      case 'd': dest              = optarg; break;
      case 'i': ip_s              = optarg; break;
      case 'j': log_json          = true; break;
      case 'm': memory_limit_mb   = strtoll(optarg, 0, 10); break;
      case 'f': net_if            = optarg; break;
      case '0': nobody            = true;   break;
      case 'n': name              = optarg; break;
      case 'p': proc              = optarg; break;
      case 'O': redir_stdout      = optarg; break;
      case 'E': redir_stderr      = optarg; break;
      case '?':
      case 'h':
      default: usage(argv[0]); exit(1); break;
    }
  }
  if (dest == NULL) { usage(argv[0]); die("message", "Cannot run without directory (-d)"); }
  if (memory_limit_mb == 0) die_errno("message", "Could not read memory limit (-m)");
  if (ip_s == NULL) log(WARNING, "message", "Running without an IP address (-i)");
  if (ip_s != NULL && inet_pton(AF_INET, ip_s, &ip) <= 0) die("message", "Cannot parse IPv4 address", "ip", ip_s);
  if (name == default_name) log(WARNING, "message", "Running with default name (-n)");
}

static int *jail_id;

#define INT_STR_LEN ((CHAR_BIT * sizeof(int) - 1) / 3 + 2)

void wait_and_bleed(pid_t fpid) {
  signal(SIGQUIT, handle_sigint);
  signal(SIGHUP, handle_sigint);
  signal(SIGINT, handle_sigint);
  signal(SIGTERM, handle_sigterm);
  int status;
  waitpid(fpid, &status, 0);
  char status_s[INT_STR_LEN];
  safe_snprintf(status_s, INT_STR_LEN, "%d", status);
  log(INFO, "message", "Process exited", "status", status_s);
  jail_remove(*jail_id); // Make sure there are no orphans in the jail
  munmap(jail_id, sizeof *jail_id);
  if (ip_s != NULL) execv("/sbin/ifconfig", (char *[]){ "ifconfig", net_if, "-alias", ip_s, 0 });
}

void run() {
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
    if (ifconfig_pid == -1) die_errno("message", "Could not fork");
    if (ifconfig_pid <= 0) { // Child
      execv("/sbin/ifconfig", (char *[]){ "ifconfig", net_if, "alias", ip_s, 0 });
    }
  }
  freopen(redir_stdout, "w", stdout);
  freopen(redir_stderr, "w", stderr);
  int jresult = jail(&j);
  if (jresult == -1) die_errno("message", "Could not start jail");
  *jail_id = jresult;
  char jail_id_s[INT_STR_LEN];
  safe_snprintf(jail_id_s, INT_STR_LEN, "%d", jresult);
  log(INFO, "message", "Process starting", "path", dest, "jid", jail_id_s, "stdout", redir_stdout, "stderr", redir_stderr, "net_if", net_if, "net_ip", ip_s);
  if (chdir("/") != 0) die_errno("message", "Could not chdir to jail");
  struct rlimit ramlimit;
  ramlimit.rlim_cur = ramlimit.rlim_max = 1024*1024*memory_limit_mb;
  if (setrlimit(RLIMIT_AS, &ramlimit) != 0) die_errno("message", "Could not limit memory");
  char *tmpname = malloc(sizeof "/tmp/ljspawn.XXXXXXXX");
  strcpy(tmpname, "/tmp/ljspawn.XXXXXXXX");
  int scriptfd = mkstemp(tmpname); // tmpname IS REPLACED WITH ACTUAL NAME HERE
  if (scriptfd == -1) die_errno("message", "Could not open temp file");
  write(scriptfd, proc, strlen(proc));
  if (fchmod(scriptfd, (unsigned short) strtol("0755", 0, 8)) != 0) die_errno("message", "Could not chmod temp file");
  close(scriptfd);
  if (nobody) {
    system("pw useradd -n nobody -d /nonexistent -s /usr/sbin/nologin 2> /dev/null");
    struct passwd* pw = getpwnam("nobody");
    if (setgroups(1, &pw->pw_gid) != 0) die_errno("message", "Could not drop groups privileges");
    if (setgid(pw->pw_gid) != 0) die_errno("message", "Could not drop group privileges");
    if (setuid(pw->pw_uid) != 0) die_errno("message", "Could not drop user privileges");
  }
  if (execve(tmpname, (char *[]){ tmpname, 0 },
      (char *[]){ "PATH=/usr/local/bin:/usr/local/sbin:/usr/games:/usr/bin:/usr/sbin:/bin:/sbin",
                  "LC_ALL=en_US.UTF-8",
                  "LANG=en_US.UTF-8", 0 }) == -1) die_errno("message", "Could not spawn process");
}

int main(int argc, char *argv[]) {
  parse_options(argc, argv);
  jail_id = (int*) mmap(NULL, sizeof *jail_id, PROT_READ | PROT_WRITE, MAP_SHARED | MAP_ANONYMOUS, -1, 0);
  if (jail_id == MAP_FAILED) die_errno("message", "Could not mmap");
  *jail_id = -1;
  p_fpid = fork();
  if (p_fpid == -1) die_errno("message", "Could not fork");
  if (p_fpid > 0) { // Parent
    wait_and_bleed(p_fpid);
  } else { // Child
    run();
  }
  return 0;
}
