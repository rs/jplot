// +build freebsd,cgo

package terminal

/*
#include <unistd.h>
*/
import "C"

import "os"

func IsTerminal(file *os.File) bool {
        return int(C.isatty(C.int(file.Fd()))) != 0
}
