package pool_test

import (
	"bytes"
	"fmt"

	"github.com/Vitaly898/metricsCollector/pkg/pool"
)

func Example() {
	buffers := pool.New(func() *bytes.Buffer { return new(bytes.Buffer) })
	buffer := buffers.Get()
	buffer.WriteString("metrics")
	fmt.Println(buffer.String())
	buffers.Put(buffer)

	buffer = buffers.Get()
	fmt.Println(buffer.Len())
	buffers.Put(buffer)

	// Output:
	// metrics
	// 0
}
