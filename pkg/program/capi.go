package native

// #cgo LDFLAGS: -L./capi/libs -lchia_capi_aarch64-apple-darwin20
// #include <stdlib.h>
// #include "capi/include/chia_capi.h"
import "C"

import "fmt"

func Random() int {
	var r C.long = C.random()
	return int(r)
}

func Seed(i int) {
	C.srandom(C.uint(i))
}

func Add(left int, right int) int32 {
	result := C.add(C.int32_t(left), C.int32_t(right))
	fmt.Println("result: ", int(result))
	return int32(result)
}
