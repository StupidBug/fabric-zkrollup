package utils

import (
	"fmt"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
)

func PrintSlice[T types.Account | types.Transaction](slice []T) {
	for i := 0; i < len(slice); i++ {
		fmt.Printf("%#v\n", slice[i])
	}
}
