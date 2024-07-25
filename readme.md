# Goche

### Cache implemented in Go(LRU, LFU)

## Example

```go
package main

import (
	"context"
	"fmt"
	"github.com/alserov/goche"
)

func main() {
	cache := goche.New(goche.LRU)

	cache.Set(context.Background(), "key", "value")

	val, ok := cache.Get(context.Background(), "key")
	fmt.Printf("Found: %v Value: %v", ok, val)
}
```

## Benchmarks

### LRU

#### Get
```text
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkLRUGet
BenchmarkLRUGet-12      70945466                16.89 ns/op
```


#### Set
```text
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkLRUSet
BenchmarkLRUSet-12      40799257                29.68 ns/op
```

### LFU

#### Get
```text
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkLFUGet
BenchmarkLFUGet-12      76725852                15.04 ns/op
```


#### Set
```text
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkLFUSet
BenchmarkLFUSet-12      43526202                27.68 ns/op
```

