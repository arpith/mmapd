#include <stdio.h>

int main() {
        char *line = NULL;
        size_t size;
        if (getline(&line, &size, stdin) == -1) {
                printf("No line");
        } else {
                printf("%s\n", line);
        }
        return 0;
}
