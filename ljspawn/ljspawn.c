#include <sys/stat.h>
#include <sys/param.h>
#include <sys/jail.h>
#include <sys/wait.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <signal.h>
#include <stdlib.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdbool.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>
#include "mounts.c"

#define DEFAULT_NAME "lj"
#define M_BUF 1024
#define safe_snprintf(dst, len, ...) if ((snprintf((dst), (len), __VA_ARGS__)) >= (len)) die("Error: args are too long!")

char *app = NULL;
char *dest = NULL;
char dest_dev[M_BUF];
char *ip_s = NULL;
struct in_addr ip;
char *name = DEFAULT_NAME;
char *proc = "echo 'Hello world'";
char *stor = NULL;
char *world = "/usr/worlds/10.0-RELEASE";
bool mount_started = false;

void unmount_dirs();

void die(const char *format, ...) {
  va_list vargs;
  va_start(vargs, format);
  fprintf(stderr, "=[ljspawn]=> ");
  vfprintf(stderr, format, vargs);
  fprintf(stderr, "\n");
  if (mount_started) unmount_dirs();
  exit(1);
}

void die_errno(const char *format, ...) {
  va_list vargs;
  va_start(vargs, format);
  fprintf(stderr, "=[ljspawn]=> ");
  vfprintf(stderr, format, vargs);
  fprintf(stderr, ". Error %d: %s\n", errno, strerror(errno));
  if (mount_started) unmount_dirs();
  exit(errno);
}

void llog(const char* format, ...) {
  va_list vargs;
  va_start(vargs, format);
  printf("=[ljspawn]=> ");
  vprintf(format, vargs);
  printf("\n");
}

void handle_sigint() {
  // Just need to handle somehow, otherwise the umount section is not called at all
}

void usage(char *pname) {
  puts("ljspawn -- spawn a process in a lightweight jail environment powered by a union mount");
  printf("usage: %s [options]\noptions:\n", pname);
  puts("  -a /path/to/app -- path to the application you want to run (mounted over world, under storage and devfs)");
  puts("  -d /path/to/dest -- path to the destination -- a temporary folder that will be created (must not exist!) as the mount point");
  puts("  -i IP.AD.DR.ESS -- the jail's IPv4 address (optional)");
  puts("  -n name -- the value for setting jailname and hostname (optional, 'lj' by default)");
  puts("  -p '/path/to/process args' -- the shell command to execute in the jail");
  puts("  -s /path/to/storage -- path to the storage (mounted over app, under dev)");
  puts("  -w /path/to/world -- path to the world, ie. FreeBSD userland you installed from /usr/src using installworld (mounted first)");
}

void parse_options(int argc, char *argv[]) {
  int c;
  while ((c = getopt(argc, argv, "a:d:ehi:n:p:s:w:?")) != -1) {
    switch (c) {
      case 'a': app   = optarg; break;
      case 'd': dest  = optarg; break;
      case 'i': ip_s  = optarg; break;
      case 'n': name  = optarg; break;
      case 'p': proc  = optarg; break;
      case 's': stor  = optarg; break;
      case 'w': world = optarg; break;
      case '?':
      case 'h':
      default: usage(argv[0]); exit(1); break;
    }
  }
  if (app == NULL) { usage(argv[0]); die("Arg -a (app directory) not found"); }
  if (dest == NULL) { usage(argv[0]); die("Arg -d (destination directory) not found"); }
  if (stor == NULL) llog("Warning: running without storage (-s)");
  if (ip_s == NULL) llog("Warning: running without IP address (-i)");
  if (ip_s != NULL && inet_pton(AF_INET, ip_s, &ip) <= 0) die("Could not parse IP address %s", ip_s);
  if (name == DEFAULT_NAME) llog("Warning: running with default name, specify with -n");
}

void mount_dirs() {
  mount_started = true;
  if (mkdir(dest, 0600) == -1) die_errno("Could not mkdir %s", dest);
  if (stor != NULL && mkdir(stor, 0600) == -1) {
    if (errno != EEXIST) {
      die_errno("Could not mkdir %s", stor);
    }
  }
  if (mount_nullfs_ro(world, dest) < 0) die_errno("Could not mount world %s to %s", world, dest);
  if (mount_unionfs_ro(app, dest) < 0)  die_errno("Could not mount app %s to %s", app, dest);
  if (stor != NULL) if (mount_unionfs_rw(stor, dest) < 0) die_errno("Could not mount storage %s to %s", stor, dest);
  safe_snprintf(dest_dev, M_BUF, "%s/dev", dest);
  if (mount_devfs(dest_dev) < 0) die_errno("Could not mount devfs to %s", dest_dev);
}

void unmount_dirs() {
  if (unmount(dest_dev, MNT_FORCE) == -1) llog("Could not unmount %s, error %d: %s", dest_dev, errno, strerror(errno));
  for (int i = 0; i < 3; i++) if (unmount(dest, MNT_FORCE) == -1) llog("Could not unmount %s, error %d: %s", dest, errno, strerror(errno));
  if (stor != NULL && rmdir(dest) == -1) llog("Could not rmdir %s, error %d: %s", dest, errno, strerror(errno));
}

int main(int argc, char *argv[]) {
  parse_options(argc, argv);
  llog("Going to start container %s with world %s and app %s\n", dest, world, app);
  mount_dirs();

  pid_t fpid = fork();
  if (fpid == -1) die_errno("Could not fork");
  if (fpid > 0) { // Parent
    signal(SIGINT, handle_sigint);
    int status;
    waitpid(fpid, &status, 0);
    llog("Process exited with status %d", status);
    unmount_dirs();
  } else { // Child
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
      safe_snprintf(shellstr, M_BUF, "ifconfig lo0 alias '%s'", ip_s); system(shellstr);
    }
    int jresult = jail(&j);
    if (jresult == -1) die_errno("Could not start jail");
    chdir("/app");
    llog("Running container %s in jail %d", dest, jresult);
    if (stor != NULL) system("echo 'nobody:*:65534:65534:Unprivileged user:/nonexistent:/usr/sbin/nologin' >> /etc/passwd");
    return execve("/bin/sh", (char *[]){ "sh", "-c", proc, 0 }, (char *[]){ 0 });
  }
}
