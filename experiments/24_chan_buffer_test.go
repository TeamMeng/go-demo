package experiments

import (
	"testing"
	"time"
)

func TestBufferedChannelBasic(t *testing.T) {
	// 有缓冲 channel 可以先把值放进队列里，容量没满时发送方不会立刻阻塞。
	ch := make(chan int, 3)

	if len(ch) != 0 || cap(ch) != 3 {
		t.Fatalf("unexpected initial state: len=%d cap=%d", len(ch), cap(ch))
	}

	ch <- 10
	ch <- 20
	ch <- 30

	if len(ch) != 3 {
		t.Fatalf("unexpected buffered len: got %d want %d", len(ch), 3)
	}

	// 缓冲 channel 底层可以把它想成一个环形队列：
	//   buf:   [10][20][30]
	//   recvx   ^          sendx ^
	// 读取时仍然遵守 FIFO，所以出队顺序和入队顺序一致。
	values := []int{<-ch, <-ch, <-ch}
	want := []int{10, 20, 30}

	for i := range want {
		if values[i] != want[i] {
			t.Fatalf("unexpected order at %d: got %d want %d", i, values[i], want[i])
		}
	}
}

func TestBufferedChannelBlocksWhenFull(t *testing.T) {
	ch := make(chan string, 1)
	ch <- "first"

	done := make(chan struct{})

	go func() {
		// 缓冲区已满时，发送方会被挂起，直到有接收方腾出空间。
		ch <- "second"
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("sender should block when buffer is full")
	case <-time.After(20 * time.Millisecond):
	}

	if got := <-ch; got != "first" {
		t.Fatalf("unexpected first value: got %q want %q", got, "first")
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sender should continue after receiver consumes one slot")
	}

	if got := <-ch; got != "second" {
		t.Fatalf("unexpected second value: got %q want %q", got, "second")
	}
}

func TestBufferedChannelCloseAndRange(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	close(ch)

	// close 不会清空缓冲区，已经写进去的数据还能继续读出来。
	first, ok := <-ch
	if first != 1 || !ok {
		t.Fatalf("unexpected first receive after close: value=%d ok=%v", first, ok)
	}

	second, ok := <-ch
	if second != 2 || !ok {
		t.Fatalf("unexpected second receive after close: value=%d ok=%v", second, ok)
	}

	// 缓冲区读空后，再读取就会拿到元素零值和 ok=false。
	third, ok := <-ch
	if third != 0 || ok {
		t.Fatalf("unexpected zero receive after drain: value=%d ok=%v", third, ok)
	}

	rangeCh := make(chan int, 3)
	rangeCh <- 7
	rangeCh <- 8
	close(rangeCh)

	var got []int
	for v := range rangeCh {
		got = append(got, v)
	}

	want := []int{7, 8}
	if len(got) != len(want) {
		t.Fatalf("unexpected range result len: got %d want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected range result at %d: got %d want %d", i, got[i], want[i])
		}
	}
}

func TestBufferedChannelRuntimeDesignNotes(t *testing.T) {
	// 这个测试本身不关心输出，而是把几个底层直觉固定下来：
	//
	// 1. runtime 里的 hchan 会记录 qcount、dataqsiz、sendx、recvx 等状态。
	// 2. 缓冲区是一段固定大小的循环数组，不是无限增长的队列。
	// 3. 缓冲区满了以后，发送 goroutine 会进入发送等待队列。
	// 4. 缓冲区空了以后，接收 goroutine 会进入接收等待队列。
	//
	// 下面这个小例子对应的是“sendx/recvx 在环形缓冲区里推进”的效果。
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2

	if got := <-ch; got != 1 {
		t.Fatalf("unexpected receive: got %d want %d", got, 1)
	}

	ch <- 3

	if got := <-ch; got != 2 {
		t.Fatalf("unexpected receive: got %d want %d", got, 2)
	}

	if got := <-ch; got != 3 {
		t.Fatalf("unexpected receive: got %d want %d", got, 3)
	}
}
