#include <stdio.h>
#include <stdbool.h>

int main() {
        char *line = NULL;
        size_t size;
        ssize_t len;
        while(true) {
                len = getline(&line, &size, stdin);
                if (len == -1) {
                        printf("No line");
                        break;
                } else {
                        printf("%s\n", line);
                }
        }
        return 0;
}
