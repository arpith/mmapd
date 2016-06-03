#include <stdio.h>
#include <stdbool.h>

int main() {
        char *line = NULL;
        size_t size;
        while(true) {
                if (getline(&line, &size, stdin) == -1) {
                        printf("No line");
                } else {
                        printf("%s\n", line);
                }
        }
        return 0;
}
