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
  ERROR,
  WARN,
  INFO
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

void llog(FILE* outfile, const loglevel_t level, ...) {
  va_list vargs; va_start(vargs, level);
  char* level_s = "UNKNOWN";
  switch(level) {
    case ERROR:   level_s = "ERROR"; break;
    case WARN:    level_s = "WARN"; break;
    case INFO:    level_s = "INFO"; break;
  }
  time_t now; time(&now);
  char* current;
  int i = 0;
  if (log_json) {
    fprintf(outfile, "{\"application\":\"ljspawn\",\"timestamp\":%ld,\"level\":\"%s\"", now, level_s);
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
        fprintf(outfile, ",\"%s\":", new_current);
      }
      free(prev);
      prev = malloc(strlen(new_current));
      memmove(prev, new_current, strlen(new_current) + 1);
      free(new_current);
    }
    if (i % 2 != 0) fprintf(outfile, "\"\"");
    fprintf(outfile, "}");
  } else {
    char time_s[sizeof "2014-14-14T14:14:14Z"];
    strftime(time_s, sizeof time_s, "%FT%TZ", gmtime(&now));
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
        case ERROR:   level_color = ANSI_BOLD_RED; break;
        case WARN:    level_color = ANSI_COLOR_RED; break;
        case INFO:    level_color = ANSI_COLOR_BLUE; break;
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
