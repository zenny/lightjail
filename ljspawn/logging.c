#ifndef _LJ_LOGGING_H
#define _LJ_LOGGING_H

#define ANSI_COLOR_RED     "\x1b[31m"
#define ANSI_BOLD_RED      "\x1b[1;31m"
#define ANSI_COLOR_GREEN   "\x1b[32m"
#define ANSI_COLOR_YELLOW  "\x1b[33m"
#define ANSI_COLOR_BLUE    "\x1b[34m"
#define ANSI_BOLD_BLUE     "\x1b[1;34m"
#define ANSI_COLOR_MAGENTA "\x1b[35m"
#define ANSI_BOLD_MAGENTA  "\x1b[1;35m"
#define ANSI_COLOR_CYAN    "\x1b[36m"
#define ANSI_COLOR_RESET   "\x1b[0m"

bool log_json = false;

typedef enum {
  EMERGENCY,
  ALERT,
  CRITICAL,
  ERROR,
  WARNING,
  NOTICE,
  INFO,
  DEBUG
} loglevel_t;

// Fuck C, that's supposed to be str.replace('"', '\\"')
char* escape_quotes(char* str) {
  unsigned long qcount = 0;
  unsigned long len = strlen(str);
  for (unsigned long i = 0; i < len; i++) {
    if (str[i] == '"') qcount++;
  }
  char* result = malloc(len + qcount + 1);
  unsigned long r = 0;
  for (unsigned long i = 0; i < len; i++) {
    if (str[i] == '"') {
      result[r] = '\\';
      r++;
    };
    result[r] = str[i];
    r++;
  }
  result[r] = '\0';
  return result;
}

char hostname[256] = "localhost";

void llog(FILE* outfile, const loglevel_t level, ...) {
  va_list vargs; va_start(vargs, level);
  time_t now; time(&now);
  char* current;
  int i = 0;
  if (log_json) {
    if (strcmp(hostname, "localhost") == 0) gethostname(hostname, 256);
    fprintf(outfile, "{\"version\":\"1.1\",\"host\":\"ljspawn@%s\",\"timestamp\":%ld,\"level\":%d", hostname, now, level);
    char* prev = malloc(1);
    while ((current = va_arg(vargs, char*)) != NULL) {
      i++;
      char* new_current = escape_quotes(current);
      if (i % 2 == 0) {
        if (strcmp(prev, "jid") == 0 || strcmp(prev, "status") == 0) {
          fprintf(outfile, "%s", new_current);
        } else {
          fprintf(outfile, "\"%s\"", new_current);
        }
      } else {
        if (strcmp(new_current, "message") == 0) {
          fprintf(outfile, ",\"short_message\":");
        } else {
          fprintf(outfile, ",\"_%s\":", new_current);
        }
      }
      free(prev);
      prev = malloc(strlen(new_current));
      memmove(prev, new_current, strlen(new_current) + 1);
      free(new_current);
    }
    if (i % 2 != 0) fprintf(outfile, "\"\"");
    fprintf(outfile, "}");
  } else {
    char* level_s = "UNKNOWN";
    switch(level) {
      case EMERGENCY:   level_s = "EMERGENCY"; break;
      case ALERT:       level_s = "ALERT"; break;
      case CRITICAL:    level_s = "CRITICAL"; break;
      case ERROR:       level_s = "ERROR"; break;
      case WARNING:     level_s = "WARNING"; break;
      case NOTICE:      level_s = "NOTICE"; break;
      case INFO:        level_s = "INFO"; break;
      case DEBUG:       level_s = "DEBUG"; break;
    }
    char time_s[sizeof "2014-14-14T14:14:14+04:00"];
    strftime(time_s, sizeof time_s, "%FT%T%z", localtime(&now));
    time_s[24] = time_s[23];
    time_s[23] = time_s[22];
    time_s[22] = ':';
    time_s[25] = '\0';
    char* app_color = "";
    char* msg_color = "";
    char* time_color = "";
    char* reset_color = "";
    char* level_color = "";
    if (isatty(1) && isatty(2)) {
      app_color = ANSI_BOLD_MAGENTA;
      msg_color = ANSI_BOLD_BLUE;
      time_color = ANSI_COLOR_YELLOW;
      reset_color = ANSI_COLOR_RESET;
      switch(level) {
        case EMERGENCY: 
        case ALERT:
        case CRITICAL:
        case ERROR:    level_color = ANSI_BOLD_RED; break;
        case WARNING:
        case NOTICE:   level_color = ANSI_COLOR_RED; break;
        case INFO:
        case DEBUG:    level_color = ANSI_COLOR_BLUE; break;
      }
    }
    fprintf(outfile, "=[%sljspawn%s]=[%s%s%s]=[%s%s%s]=> ", app_color, reset_color, level_color, level_s, reset_color, time_color, time_s, reset_color);
    while ((current = va_arg(vargs, char*)) != NULL) {
      i++;
      if (i % 2 == 0) {
        fprintf(outfile, "=%s  ", current);
      } else {
        fprintf(outfile, "%s%s%s", msg_color, current, reset_color);
      }
    }
  }
  fprintf(outfile, "\n");
  va_end(vargs);
}

#define die(...) if (1) { llog(stderr, ERROR, __VA_ARGS__, NULL); exit(1); }
#define die_errno(...) if (1) { llog(stderr, ERROR, __VA_ARGS__, "errno", errno, "strerror", strerror(errno), NULL); exit(errno); }
#define log(level, ...) llog(stdout, (level), __VA_ARGS__, NULL)
#endif
