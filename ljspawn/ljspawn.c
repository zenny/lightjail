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
#include <errno.h>
#include <string.h>
#include <unistd.h>

#define M_BUF 2048
#define safe_snprintf(dst, len, ...) if ((snprintf((dst), (len), __VA_ARGS__)) >= (len)) die("Error: args are too long!")

char *app = NULL;
char *dest = NULL;
char *eph = NULL;
char *ip_s = NULL;
struct in_addr ip;
char *proc = "echo 'Hello world'";
char *world = "/usr/worlds/10.0-RELEASE";
char mountstr[M_BUF];

void die(const char *format, ...) {
  va_list vargs;
  va_start(vargs, format);
  fprintf(stderr, "=[ljspawn]=> ");
  vfprintf(stderr, format, vargs);
  fprintf(stderr, "\n");
  exit(1);
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

void parse_options(int argc, char *argv[]) {
  int c;
  while ((c = getopt(argc, argv, "a:d:e:i:p:w:")) != -1) {
    switch (c) {
      case 'a': app   = optarg; break;
      case 'd': dest  = optarg; break;
      case 'e': eph   = optarg; break;
      case 'i': ip_s  = optarg; break;
      case 'p': proc  = optarg; break;
      case 'w': world = optarg; break;
    }
  }
  if (app == NULL) die("Arg -a (app directory) not found");
  if (dest == NULL) die("Arg -d (destination directory) not found");
  if (eph == NULL) llog("Warning: running without ephemeral storage");
  if (ip_s == NULL) llog("Warning: running without IP address");
  if (ip_s != NULL && inet_pton(AF_INET, ip_s, &ip) <= 0) die("Could not parse IP address %s", ip_s);
}

void mount_dirs() {
  if (mkdir(dest, 0600) == -1) die("Could not mkdir %s, error %d: %s", dest, errno, strerror(errno));
  if (eph != NULL && mkdir(eph, 0600) == -1) die("Could not mkdir %s, error %d: %s", eph, errno, strerror(errno));
  safe_snprintf(mountstr, M_BUF, "mount_nullfs -o ro '%s' '%s'", world, dest); system(mountstr);
  safe_snprintf(mountstr, M_BUF, "mount_unionfs -o ro '%s' '%s'", app, dest); system(mountstr);
  if (eph != NULL) { safe_snprintf(mountstr, M_BUF, "mount_unionfs -o noatime '%s' '%s'", eph, dest); system(mountstr); }
  safe_snprintf(mountstr, M_BUF, "mount -t devfs devfs '%s/dev'", dest); system(mountstr);
}

void unmount_dirs() {
  safe_snprintf(mountstr, M_BUF, "umount '%s/dev'", dest); system(mountstr);
  safe_snprintf(mountstr, M_BUF, "umount '%s'", dest); system(mountstr);
  safe_snprintf(mountstr, M_BUF, "umount '%s'", dest); system(mountstr);
  safe_snprintf(mountstr, M_BUF, "umount '%s'", dest); system(mountstr);
  safe_snprintf(mountstr, M_BUF, "rmdir '%s'", dest); system(mountstr);
  if (eph != NULL) { safe_snprintf(mountstr, M_BUF, "rm -r '%s'", eph); system(mountstr); }
}

int main(int argc, char *argv[]) {
  parse_options(argc, argv);
  llog("Going to start container %s with world %s and app %s\n", dest, world, app);
  mount_dirs();

  pid_t fpid = fork();
  if (fpid == -1) die("Could not fork, error %d: %s", errno, strerror(errno));
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
    j.hostname = "";
    j.jailname = "lj";
    j.ip4s = 0;
    j.ip6s = 0;
    if (ip_s != NULL) {
      j.ip4s++;
      j.ip4 = malloc(sizeof(struct in_addr) * j.ip4s);
      j.ip4[0] = ip;
    }
    int jresult = jail(&j);
    if (jresult == -1) die("Could not start jail, error %d: %s", errno, strerror(errno));
    chdir("/app");
    llog("Running container %s in jail %d", dest, jresult);
    if (eph != NULL) system("echo 'nobody:*:65534:65534:Unprivileged user:/nonexistent:/usr/sbin/nologin' >> /etc/passwd");
    return execve("/bin/sh", (char *[]){ "sh", "-c", proc, 0 }, (char *[]){ 0 });
  }
}
