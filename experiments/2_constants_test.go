package experiments

import (
	"fmt"
	"testing"
)

const s string = "constant"

func TestConstant(t *testing.T) {
	fmt.Println(s)

	const n string = "web3"
	fmt.Println(n)

	const r = 50000
	fmt.Println(r)
}
