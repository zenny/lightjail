#include <sys/param.h>
#include <sys/mount.h>
#include <sys/uio.h>

void build_iovec(struct iovec **iov, int *iovlen, const char *name, void *val, size_t len) {
  // copyright (c) 1994 The Regents of the University of California
  int i; 
  if (*iovlen < 0) return;
  i = *iovlen;
  *iov = realloc(*iov, sizeof **iov * (i + 2));
  if (*iov == NULL) {
    *iovlen = -1;
    return;
  }
  (*iov)[i].iov_base = strdup(name);
  (*iov)[i].iov_len = strlen(name) + 1;
  i++;
  (*iov)[i].iov_base = val;
  if (len == (size_t)-1) {
    if (val != NULL)
      len = strlen(val) + 1;
    else
      len = 0;
  }
  (*iov)[i].iov_len = (int)len;
  *iovlen = ++i;
}

int mount_nullfs_ro(char *from, char *to) {
  struct iovec *iov = NULL;
  int iovlen = 0;
  build_iovec(&iov, &iovlen, "fstype", "nullfs", (size_t)-1);
  build_iovec(&iov, &iovlen, "fspath", to, (size_t)-1);
  build_iovec(&iov, &iovlen, "target", from, (size_t)-1);
  build_iovec(&iov, &iovlen, "ro", "", (size_t)-1);
  return nmount(iov, iovlen, MNT_RDONLY);
}

int mount_unionfs_ro(char *from, char *to) {
  struct iovec *iov = NULL;
  int iovlen = 0;
  build_iovec(&iov, &iovlen, "fstype", "unionfs", (size_t)-1);
  build_iovec(&iov, &iovlen, "fspath", to, (size_t)-1);
  build_iovec(&iov, &iovlen, "target", from, (size_t)-1);
  build_iovec(&iov, &iovlen, "ro", "", (size_t)-1);
  return nmount(iov, iovlen, MNT_RDONLY);
}

int mount_unionfs_rw(char *from, char *to) {
  struct iovec *iov = NULL;
  int iovlen = 0;
  build_iovec(&iov, &iovlen, "fstype", "unionfs", (size_t)-1);
  build_iovec(&iov, &iovlen, "fspath", to, (size_t)-1);
  build_iovec(&iov, &iovlen, "target", from, (size_t)-1);
  build_iovec(&iov, &iovlen, "noatime", "", (size_t)-1);
  return nmount(iov, iovlen, MNT_NOATIME);
}

int mount_devfs(char *to) {
  struct iovec *iov = NULL;
  int iovlen = 0;
  build_iovec(&iov, &iovlen, "fstype", "devfs", (size_t)-1);
  build_iovec(&iov, &iovlen, "fspath", to, (size_t)-1);
  return nmount(iov, iovlen, 0);
}
