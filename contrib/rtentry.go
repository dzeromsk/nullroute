//go:build ignore
// +build ignore

package ignore

// #include <net/route.h>
import "C"

type RTEntry C.struct_rtentry
