# Chapter 15 — Channel

# Part I — Why Channel Exists

Trong các ngôn ngữ lập trình truyền thống, khi nhiều Thread cùng truy cập vào một vùng nhớ, lập trình viên phải sử dụng:

- Mutex
- Semaphore
- Condition Variable
- Read Write Lock

để tránh Race Condition.

Go lựa chọn một hướng tiếp cận khác.

Thay vì tập trung vào việc **chia sẻ dữ liệu**, Go khuyến khích **truyền dữ liệu** giữa các Goroutine.

Đó chính là lý do Channel ra đời.

Trong phần này chúng ta sẽ tìm hiểu:

- Vì sao Go cần Channel.
- Triết lý thiết kế của Channel.
- Các mô hình giao tiếp phổ biến.
- Buffered và Unbuffered Channel khác nhau như thế nào.

Đây sẽ là nền tảng trước khi chúng ta đi sâu vào source code của `runtime/chan.go`.

---

# 15.1 Why Channel

## Bài toán của Concurrent Programming

Giả sử có hai Thread cùng cập nhật một biến.

```go
counter++
```

Thoạt nhìn đây chỉ là một dòng code đơn giản.

Nhưng thực tế CPU thực hiện nhiều bước.

```text
Load counter

↓

Increment

↓

Store counter
```

Nếu hai Thread cùng thực hiện.

```text
Thread A

Load = 10

-------------------

Thread B

Load = 10

-------------------

Thread A

Store = 11

-------------------

Thread B

Store = 11
```

Kết quả cuối cùng.

```text
11
```

Thay vì:

```text
12
```

Đây chính là **Race Condition**.

---

# Cách tiếp cận truyền thống

Để giải quyết vấn đề trên.

Các ngôn ngữ như C/C++ hoặc Java thường sử dụng Mutex.

```go
mutex.Lock()

counter++

mutex.Unlock()
```

Ý tưởng rất đơn giản.

```text
Shared Memory

↓

Lock

↓

Modify

↓

Unlock
```

Cách làm này hoạt động.

Nhưng khi hệ thống lớn dần.

Bạn sẽ gặp rất nhiều vấn đề.

- Deadlock
- Lock Contention
- Priority Inversion
- Starvation
- Khó debug

---

# Go chọn một hướng khác

Go chịu ảnh hưởng mạnh từ ngôn ngữ **Newsqueak**, **Alef** và đặc biệt là **Limbo**, những ngôn ngữ do nhóm của Rob Pike và đồng nghiệp phát triển. Các ngôn ngữ này kế thừa tư tưởng của mô hình **Communicating Sequential Processes (CSP)** do :contentReference[oaicite:0]{index=0} đề xuất.

Triết lý nổi tiếng của Go là:

> **"Do not communicate by sharing memory; instead, share memory by communicating."**

Thay vì:

```text
Nhiều Goroutine

↓

Shared Memory

↓

Mutex
```

Go khuyến khích:

```text
Goroutine A

↓

Channel

↓

Goroutine B
```

Dữ liệu được truyền từ Goroutine này sang Goroutine khác.

Không còn nhiều Goroutine cùng sửa một Object.

---

# Message Passing

Có thể mô hình hóa như sau.

```text
+-------------+

 Goroutine A

+-------------+

       │

       │ Send

       ▼

   +---------+

   | Channel |

   +---------+

       │

       │ Receive

       ▼

+-------------+

 Goroutine B

+-------------+
```

Channel đóng vai trò là:

- đường truyền dữ liệu
- điểm đồng bộ (Synchronization Point)
- cơ chế giao tiếp giữa các Goroutine

---

# Ownership Transfer

Một cách hiểu rất quan trọng.

Giả sử.

```go
jobs <- task
```

Sau khi gửi.

```text
Task

↓

Ownership

↓

Receiver
```

Thay vì nhiều Goroutine cùng sửa một Object.

Go khuyến khích:

```text
Producer

↓

Transfer Ownership

↓

Consumer
```

Điều này giảm đáng kể khả năng xảy ra Race Condition.

---

# Channel không chỉ là Queue

Nhiều người nghĩ.

```text
Channel = Queue
```

Điều này chưa đúng.

Channel đồng thời là:

- Queue
- Synchronization Primitive
- Blocking Mechanism
- Wakeup Mechanism
- Memory Barrier

Trong các phần sau của chương.

Chúng ta sẽ thấy Runtime sử dụng:

- sudog
- gopark()
- goready()

để biến Channel thành một primitive đồng bộ hóa hoàn chỉnh.

---

# Khi nào nên dùng Channel?

Channel phù hợp khi:

- truyền dữ liệu giữa Goroutine
- xây dựng Pipeline
- Worker Pool
- Event Notification
- Streaming
- Task Queue

Ví dụ.

```go
jobs := make(chan Job)

results := make(chan Result)
```

---

# Khi nào không nên dùng Channel?

Không phải mọi Concurrent Program đều cần Channel.

Ví dụ.

```go
counter++
```

Nếu chỉ cần bảo vệ một biến.

```text
Mutex

↓

Đơn giản hơn
```

Trong nhiều trường hợp.

Mutex:

- nhanh hơn
- ít Allocation hơn
- ít Context Switch hơn

Đây là lý do Go vẫn cung cấp:

```go
sync.Mutex
```

---

# 15.2 Channel Model

Channel không chỉ dùng để truyền dữ liệu.

Nó còn là nền tảng của rất nhiều mô hình Concurrent Programming.

---

# Producer – Consumer

Đây là mô hình phổ biến nhất.

```text
        Producer

            │

            ▼

        +---------+

        | Channel |

        +---------+

            │

            ▼

        Consumer
```

Ví dụ.

```go
jobs := make(chan Job)

go producer(jobs)

go consumer(jobs)
```

Producer chỉ tạo dữ liệu.

Consumer chỉ xử lý dữ liệu.

Hai Goroutine hoàn toàn độc lập.

---

## Nhiều Producer

Một Channel có thể nhận dữ liệu từ nhiều Producer.

```text
Producer A

          \

Producer B ---->

             Channel

Producer C ---->
```

Điều này thường gặp trong:

- Logging
- Event Bus
- Metrics

---

## Nhiều Consumer

Ngược lại.

Một Channel cũng có nhiều Consumer.

```text
             Channel

        /      |      \

      C1      C2      C3
```

Runtime sẽ đánh thức một Goroutine phù hợp để nhận dữ liệu.

---

# Pipeline

Pipeline chia công việc thành nhiều giai đoạn.

```text
Read File

↓

Parse

↓

Validate

↓

Save Database
```

Mỗi Stage chạy trên một Goroutine riêng.

```text
Reader

↓

Channel

↓

Parser

↓

Channel

↓

Validator

↓

Channel

↓

Writer
```

Ưu điểm.

- Song song hóa.
- Dễ mở rộng.
- Dễ bảo trì.

---

# Fan-out

Một Producer.

Nhiều Worker.

```text
          Producer

              │

              ▼

          Job Channel

        /      |      \

     W1       W2      W3
```

Các Worker cùng lấy Job.

Đây là mô hình Worker Pool.

---

# Fan-in

Ngược lại.

Nhiều Producer.

Một Consumer.

```text
Worker 1

          \

Worker 2 ---->

             Result Channel

Worker 3 ---->

                 │

                 ▼

             Collector
```

Thường dùng khi:

- tổng hợp kết quả
- merge dữ liệu
- aggregation

---

# Pipeline kết hợp Fan-out và Fan-in

Đây là kiến trúc rất phổ biến.

```text
             Producer

                 │

                 ▼

            Job Channel

        /       |       \

      W1       W2       W3

        \       |       /

         \      |      /

          Result Channel

                 │

                 ▼

             Collector
```

Mô hình này xuất hiện rất nhiều trong:

- HTTP Server
- ETL
- Message Queue
- Streaming
- Distributed System

---

# 15.3 Buffered vs Unbuffered

Một trong những câu hỏi đầu tiên khi tạo Channel là:

Có cần Buffer hay không?

---

# Unbuffered Channel

```go
ch := make(chan int)
```

Capacity.

```text
0
```

Không có Buffer.

```text
Sender

↓

Channel

↓

Receiver
```

Muốn Send thành công.

Bắt buộc phải có Receiver.

---

## Minh họa

```go
ch <- 10
```

Nếu chưa có Receiver.

```text
Sender

↓

Waiting
```

Runtime sẽ:

```text
gopark()
```

đưa Goroutine vào trạng thái Waiting.

---

## Synchronization Point

Đây là đặc điểm quan trọng nhất.

```text
Sender

↓

Handshake

↓

Receiver
```

Hai Goroutine gặp nhau tại cùng một thời điểm.

Do đó.

Unbuffered Channel đồng thời là:

```text
Synchronization Primitive
```

---

# Buffered Channel

```go
ch := make(chan int, 4)
```

Capacity.

```text
4
```

Runtime tạo Buffer.

```text
+----------------+

10

20

30

40

+----------------+
```

---

# Send

Nếu Buffer chưa đầy.

```go
ch <- value
```

Runtime chỉ:

```text
Copy

↓

Buffer
```

Sender tiếp tục chạy.

Không cần chờ Receiver.

---

# Receive

Receiver lấy dữ liệu.

```text
Buffer

↓

Receiver
```

Không cần đợi Sender.

---

# Khi Buffer đầy

Ví dụ.

```text
Capacity = 4
```

Buffer.

```text
10

20

30

40
```

Sender tiếp theo.

```go
ch <- 50
```

Runtime sẽ:

```text
Park Sender
```

cho tới khi có Receiver lấy dữ liệu.

---

# So sánh

| Đặc điểm | Unbuffered | Buffered |
|----------|------------|----------|
| Capacity | 0 | > 0 |
| Send | Block cho đến khi có Receiver | Chỉ Block khi Buffer đầy |
| Receive | Block cho đến khi có Sender | Chỉ Block khi Buffer rỗng |
| Đồng bộ | Có | Ít hơn |
| Throughput | Thấp hơn | Cao hơn |
| Latency | Thấp | Có thể cao hơn nếu Buffer lớn |

---

# Khi nào dùng Unbuffered?

Phù hợp khi cần:

- Synchronization
- Handshake
- Signal
- Notification
- Ownership Transfer

Ví dụ.

```go
done := make(chan struct{})
```

---

# Khi nào dùng Buffered?

Phù hợp khi:

- Producer nhanh hơn Consumer.
- Pipeline.
- Worker Pool.
- Message Queue.
- Batch Processing.

Ví dụ.

```go
jobs := make(chan Job, 1024)
```

---

# Buffered có luôn tốt hơn?

Không.

Buffer quá lớn có thể:

- tăng Memory Usage
- tăng GC Pressure
- che giấu Bottleneck
- tăng Latency

Trong khi Buffer quá nhỏ lại làm tăng Blocking.

Việc lựa chọn Capacity phù hợp cần dựa trên Benchmark và đặc điểm của từng hệ thống.

---

# Tổng kết Part I

Cho đến thời điểm này, chúng ta có thể nhìn Channel dưới góc nhìn của Go Runtime như sau.

```text
                Goroutine

                    │

            Send / Receive

                    │

                    ▼

              +------------+

              |  Channel   |

              +------------+

                    │

         +----------+-----------+

         │                      │

         ▼                      ▼

 Synchronization          Data Transfer

         │                      │

         +----------+-----------+

                    │

                    ▼

             Runtime Scheduler

                    │

                    ▼

             Park / Wakeup G
```

Channel không chỉ là một hàng đợi dữ liệu.

Nó là một **primitive đồng bộ hóa** được xây dựng trực tiếp trong Go Runtime, kết hợp giữa:

- Message Passing
- Synchronization
- Scheduler
- Memory Model

Đây chính là nền tảng để chúng ta bước vào phần tiếp theo của chương, nơi sẽ phân tích chi tiết cấu trúc dữ liệu `hchan` và cách `runtime/chan.go` hiện thực hóa toàn bộ cơ chế này.

## Key Takeaways

- Channel được thiết kế dựa trên triết lý **message passing** thay vì **shared memory**.
- Channel vừa là cơ chế truyền dữ liệu, vừa là primitive đồng bộ hóa giữa các Goroutine.
- Các mô hình như **Producer–Consumer**, **Pipeline**, **Fan-out** và **Fan-in** đều được xây dựng tự nhiên trên Channel.
- **Unbuffered Channel** yêu cầu Sender và Receiver gặp nhau tại cùng một thời điểm, tạo thành một synchronization point.
- **Buffered Channel** bổ sung một hàng đợi trung gian, giúp tăng throughput nhưng cần lựa chọn kích thước buffer hợp lý.
- Hiểu rõ các mô hình và đặc tính của Channel là nền tảng trước khi phân tích cấu trúc `hchan`, `sudog` và mã nguồn `runtime/chan.go`.

# Part II — Runtime Data Structure

Ở Part I, chúng ta đã hiểu **Channel là gì** dưới góc nhìn của lập trình viên.

Trong phần này, chúng ta sẽ chuyển sang góc nhìn của **Go Runtime**.

Khi viết:

```go
ch := make(chan int, 10)
```

Runtime không tạo một "Channel" theo nghĩa trừu tượng.

Thay vào đó, nó cấp phát một cấu trúc dữ liệu tên là:

```go
runtime.hchan
```

Toàn bộ hoạt động:

- Send
- Receive
- Close
- Blocking
- Wakeup

đều được thực hiện trên `hchan`.

Nếu xem `Processor (P)` là trung tâm của Scheduler thì `hchan` chính là trung tâm của toàn bộ cơ chế Channel.

---

# 15.4 Anatomy of `hchan`

## `hchan` nằm ở đâu?

Định nghĩa của Channel nằm trong:

```text
src/runtime/chan.go
```

Cấu trúc thực tế (đã lược bỏ một số field):

```go
type hchan struct {
    qcount   uint
    dataqsiz uint
    buf      unsafe.Pointer
    elemsize uint16
    closed   uint32

    sendx uint
    recvx uint

    recvq waitq
    sendq waitq

    lock mutex
}
```

Đây là một trong những struct quan trọng nhất của Go Runtime.

---

# Kiến trúc tổng thể

Có thể hình dung `hchan` như sau.

```text
                hchan

    +----------------------------------+

    Metadata

    qcount
    capacity
    elemSize
    closed

    ------------------------------

    Ring Buffer

    ------------------------------

    sendq

    recvq

    ------------------------------

    mutex

    +----------------------------------+
```

Channel không chỉ chứa dữ liệu.

Nó còn chứa:

- Buffer
- Queue của Sender
- Queue của Receiver
- Lock
- Metadata

---

# qcount

```go
qcount uint
```

Lưu số lượng phần tử hiện có trong Buffer.

Ví dụ.

```go
ch := make(chan int, 8)
```

Sau khi gửi.

```go
ch <- 10
ch <- 20
```

Runtime có:

```text
Capacity = 8

qcount = 2
```

---

# dataqsiz

```go
dataqsiz uint
```

Dung lượng của Buffer.

Ví dụ.

```go
make(chan int, 100)
```

Runtime lưu:

```text
dataqsiz

↓

100
```

Nếu:

```go
make(chan int)
```

thì:

```text
dataqsiz

↓

0
```

Đây là Unbuffered Channel.

---

# buf

```go
buf unsafe.Pointer
```

Trỏ tới vùng nhớ chứa dữ liệu.

```text
buf

↓

+----------------------+

Object

Object

Object

+----------------------+
```

Nếu:

```go
make(chan int)
```

Runtime sẽ không cấp phát Buffer.

```text
buf = nil
```

---

# elemsize

```go
elemsize
```

Kích thước của từng phần tử.

Ví dụ.

```go
chan int64
```

```text
elemsize

↓

8
```

```go
chan User
```

```text
sizeof(User)
```

Runtime sử dụng giá trị này để tính offset trong Buffer.

---

# closed

```go
closed uint32
```

Lưu trạng thái Channel.

```text
0

↓

Open
```

```text
1

↓

Closed
```

Khi gọi.

```go
close(ch)
```

Runtime chỉ cần.

```text
closed = 1
```

Sau đó Wakeup toàn bộ Goroutine đang chờ.

---

# sendx

```go
sendx
```

Vị trí tiếp theo sẽ ghi dữ liệu.

Ví dụ.

```text
Capacity = 8

sendx = 3
```

Sender tiếp theo sẽ ghi vào.

```text
Index = 3
```

---

# recvx

```go
recvx
```

Vị trí Receiver sẽ đọc.

```text
recvx = 1
```

Receiver đọc:

```text
Buffer[1]
```

Sau đó.

```text
recvx++
```

---

# sendq

```go
sendq
```

Queue chứa các Sender đang Block.

```text
sendq

↓

G1

↓

G2

↓

G3
```

Nếu Buffer đầy.

Sender sẽ được đưa vào Queue này.

---

# recvq

```go
recvq
```

Queue chứa Receiver đang chờ.

```text
recvq

↓

G8

↓

G9

↓

G10
```

Nếu Buffer rỗng.

Receiver sẽ Park tại đây.

---

# mutex

Toàn bộ Channel được bảo vệ bởi:

```go
mutex
```

Điều này đảm bảo.

```text
Send

Receive

Close
```

không thể sửa đổi Channel đồng thời.

---

# Tổng quan `hchan`

```text
                 hchan

    +--------------------------------+

    qcount

    capacity

    closed

    sendx

    recvx

    --------------------------

    Buffer Pointer

    --------------------------

    sendq

    recvq

    --------------------------

    mutex

    +--------------------------------+
```

---

# 15.5 Ring Buffer

## Tại sao Runtime không dùng Slice?

Có thể nghĩ rằng.

```go
[]T
```

đã đủ.

Ví dụ.

```go
append()
```

```go
pop()
```

Nhưng.

Slice có nhiều vấn đề.

- append có thể Allocate.
- copy dữ liệu.
- di chuyển phần tử.

Scheduler không muốn điều này.

---

# Circular Buffer

Runtime sử dụng:

> Ring Buffer

```text
        +--------------------------+

        | 0 | 1 | 2 | 3 | 4 | 5 |

        +--------------------------+

          ↑                 ↑

        recvx            sendx
```

Sau khi gửi.

```text
sendx++
```

Nếu.

```text
sendx == capacity
```

Runtime quay về.

```text
sendx = 0
```

Đây là lý do gọi là Ring Buffer.

---

# Ví dụ

Capacity.

```text
4
```

Ban đầu.

```text
+----+----+----+----+

|    |    |    |    |

+----+----+----+----+

sendx = 0

recvx = 0
```

---

# Gửi

```go
1
```

```text
+----+----+----+----+

| 1  |    |    |    |

+----+----+----+----+

sendx = 1
```

---

```go
2
```

```text
+----+----+----+----+

| 1  | 2  |    |    |

+----+----+----+----+

sendx = 2
```

---

# Receive

Receiver lấy.

```text
Buffer[0]
```

```text
+----+----+----+----+

|    | 2  |    |    |

+----+----+----+----+

recvx = 1
```

---

# Tiếp tục Send

```go
3

4

5
```

Sau khi quay vòng.

```text
+----+----+----+----+

| 5  | 2  | 3  | 4  |

+----+----+----+----+

recvx = 1

sendx = 1
```

Không cần Allocate mới.

Không cần Copy.

---

# Tại sao Ring Buffer nhanh?

Độ phức tạp.

```
Push

O(1)
```

```
Pop

O(1)
```

Không có:

- append
- copy
- realloc

Đây là lý do Runtime lựa chọn Ring Buffer thay vì Slice.

---

# Luồng Send

```text
Send

↓

sendx

↓

Copy

↓

sendx++

↓

qcount++
```

---

# Luồng Receive

```text
Receive

↓

recvx

↓

Copy

↓

recvx++

↓

qcount--
```

---

# 15.6 `sudog`

## `sudog` là gì?

Đây là một trong những struct ít được biết đến nhất trong Go Runtime.

Nhưng lại xuất hiện ở:

- Channel
- Mutex
- Cond
- Semaphore
- Select

Có thể hiểu đơn giản.

> `sudog` là một đối tượng đại diện cho **một Goroutine đang chờ đồng bộ hóa**.

---

# Tại sao cần `sudog`?

Giả sử.

```go
<-ch
```

Nếu Channel chưa có dữ liệu.

Runtime phải lưu:

```text
Goroutine nào đang chờ?
```

Không thể đưa trực tiếp `g` vào Queue.

Vì.

Một Goroutine có thể đồng thời tham gia nhiều cơ chế Runtime khác nhau.

Runtime cần một lớp trung gian.

Đó chính là `sudog`.

---

# Quan hệ

```text
        Goroutine

             │

             ▼

          sudog

             │

             ▼

          Channel
```

Channel không lưu trực tiếp `g`.

Mà lưu.

```text
sudog

↓

g
```

---

# Cấu trúc

Đơn giản hóa.

```go
type sudog struct {

    g *g

    next *sudog

    prev *sudog

    elem unsafe.Pointer

}
```

Trong Runtime thực tế.

`sudog` còn nhiều field khác.

Nhưng đây là những field quan trọng nhất.

---

# sendq

Khi Sender Block.

```go
ch <- value
```

Runtime tạo.

```text
sudog
```

Sau đó.

```text
sendq

↓

sudog

↓

g
```

---

# recvq

Tương tự.

```go
<-ch
```

Nếu chưa có Sender.

```text
recvq

↓

sudog

↓

g
```

---

# Wakeup

Khi có dữ liệu.

Runtime.

```text
recvq

↓

Pop sudog

↓

goready(g)
```

Goroutine trở lại Runnable.

---

# sudog không phải Goroutine

Đây là điều rất nhiều người nhầm.

```text
sudog

≠

goroutine
```

`sudog` chỉ là:

```text
Waiting Descriptor
```

Một Goroutine.

Có thể tạo.

Nhiều `sudog`.

Ví dụ.

```go
select {
case <-ch1:

case <-ch2:

case <-ch3:

}
```

Runtime sẽ tạo.

```text
sudog

↓

ch1

----------------

sudog

↓

ch2

----------------

sudog

↓

ch3
```

Nhưng vẫn chỉ có:

```text
1 Goroutine
```

Đây cũng là lý do `select` phức tạp hơn Channel thông thường.

---

# sudog Cache

Việc tạo.

```text
new(sudog)
```

liên tục.

Rất tốn.

Runtime tạo.

```text
Sudog Cache
```

để tái sử dụng.

```text
Need sudog

↓

Cache

↓

Reuse
```

Giảm:

- Allocation
- GC
- Fragmentation

---

# Quan hệ giữa `hchan` và `sudog`

Có thể mô tả toàn bộ như sau.

```text
                    hchan

        +----------------------------------+

        Metadata

        ----------------------

        Ring Buffer

        ----------------------

        sendq

            │

            ▼

         sudog

            │

            ▼

         Goroutine

        ----------------------

        recvq

            │

            ▼

         sudog

            │

            ▼

         Goroutine

        ----------------------

             mutex

        +----------------------------------+
```

Có thể thấy:

- `hchan` chịu trách nhiệm quản lý trạng thái của Channel.
- Ring Buffer lưu dữ liệu.
- `sendq` và `recvq` quản lý các Goroutine đang chờ.
- `sudog` là cầu nối giữa Goroutine và Channel.
- `mutex` bảo vệ toàn bộ cấu trúc dữ liệu.

Đây chính là nền tảng để Runtime hiện thực hóa các thao tác `send`, `receive`, `close` và `select` mà chúng ta sẽ phân tích trong các phần tiếp theo.

---

## Key Takeaways

- `hchan` là cấu trúc dữ liệu cốt lõi hiện thực hóa Channel trong Go Runtime.
- `hchan` bao gồm ba nhóm thành phần chính: **Metadata**, **Ring Buffer** và **Wait Queue**.
- Ring Buffer cho phép Send và Receive diễn ra với độ phức tạp **O(1)** mà không cần cấp phát hay sao chép dữ liệu.
- `sendq` và `recvq` không lưu trực tiếp Goroutine mà lưu các đối tượng `sudog`.
- `sudog` là descriptor đại diện cho một Goroutine đang chờ đồng bộ hóa, được sử dụng không chỉ trong Channel mà còn trong Mutex, Semaphore và Select.
- Hiểu rõ `hchan`, Ring Buffer và `sudog` là điều kiện tiên quyết để đọc và phân tích các hàm `makechan()`, `chansend()`, `chanrecv()` và `closechan()` trong `runtime/chan.go`.

# Part III — Send Operation

Ở phần trước, chúng ta đã tìm hiểu cấu trúc dữ liệu của Channel:

- `hchan`
- Ring Buffer
- `sendq`
- `recvq`
- `sudog`

Trong phần này, chúng ta sẽ phân tích điều gì thực sự xảy ra khi chương trình thực hiện:

```go
ch <- value
```

Đây là một trong những thao tác quan trọng nhất của Go Runtime.

Một phép Send có thể diễn ra theo nhiều hướng khác nhau:

- Copy dữ liệu vào Buffer.
- Chuyển trực tiếp dữ liệu sang Receiver.
- Đánh thức Goroutine đang chờ.
- Block Goroutine hiện tại.

Runtime phải quyết định rất nhanh nên đi theo con đường nào.

---

# 15.7 Channel Send Flow

## Khi chương trình gọi

```go
ch <- value
```

Compiler sẽ không sinh ra lệnh ghi trực tiếp vào Channel.

Thay vào đó.

Nó gọi Runtime.

```text
User Code

↓

runtime.chansend()
```

Toàn bộ logic của Channel đều nằm trong:

```text
src/runtime/chan.go
```

---

# Tổng quan luồng Send

Có thể mô tả đơn giản như sau.

```text
                chansend()

                     │

                     ▼

            Receiver Waiting ?

              │            │

            Yes           No

              │            │

              ▼            ▼

      Direct Handoff   Buffer Full ?

                           │

                     ┌─────┴─────┐

                     │           │

                    No          Yes

                     │           │

                     ▼           ▼

              Copy vào Buffer   Park Sender
```

Đây là thuật toán tổng quát.

Trong Runtime thực tế còn nhiều bước tối ưu khác.

---

# Runtime thực hiện những gì?

Một lời gọi Send trải qua các bước.

```text
Acquire Lock

↓

Check Closed

↓

Check Receiver

↓

Check Buffer

↓

Copy

↓

Wakeup

↓

Unlock
```

Nếu không thể Send.

Runtime sẽ:

```text
Park Current Goroutine
```

---

# Pseudo Code

Có thể mô phỏng như sau.

```go
lock(ch)

if closed {

    panic()

}

if receiverWaiting {

    directSend()

}

else if bufferAvailable {

    enqueue()

}

else {

    parkSender()

}

unlock(ch)
```

Đây chính là ý tưởng cốt lõi của `runtime.chansend()`.

---

# Fast Path và Slow Path

Runtime cố gắng ưu tiên Fast Path.

```text
Fast Path

↓

Receiver Waiting

↓

Direct Send

-----------------------

Buffer Available

↓

Copy

-----------------------

Slow Path

↓

Park Sender
```

Phần lớn các ứng dụng thực tế đều hoạt động trên Fast Path.

---

# 15.8 Buffered Channel Send

Giả sử.

```go
ch := make(chan int, 4)
```

Buffer hiện tại.

```text
+----+----+----+----+

|    |    |    |    |

+----+----+----+----+
```

---

# Bước 1 — Acquire Lock

Runtime khóa Channel.

```text
lock(hchan)
```

Mọi thao tác:

- Send
- Receive
- Close

đều phải giữ Lock.

Điều này đảm bảo.

```text
Channel State

↓

Consistent
```

---

# Bước 2 — Kiểm tra Receiver

Runtime kiểm tra.

```text
recvq
```

Nếu:

```text
recvq

↓

Empty
```

Không có Receiver đang chờ.

Runtime chuyển sang Buffer.

---

# Bước 3 — Kiểm tra Buffer

Runtime kiểm tra.

```text
qcount

↓

capacity ?
```

Nếu.

```text
qcount < dataqsiz
```

Buffer còn chỗ.

---

# Bước 4 — Copy Memory

Đây là bước quan trọng.

Giả sử.

```go
type User struct {

    ID int

}
```

```go
u := User{ID:1}

ch <- u
```

Runtime thực hiện.

```text
memmove()

↓

Buffer
```

Không lưu:

```text
&u
```

Mà lưu:

```text
Copy(User)
```

Đây là lý do.

Sau khi Send.

```go
u.ID = 100
```

không ảnh hưởng tới dữ liệu đã nằm trong Channel.

---

# Minh họa

```text
Stack

User

↓

Copy

↓

Channel Buffer

User
```

Hai Object hoàn toàn độc lập.

---

# Bước 5 — Cập nhật Ring Buffer

Sau khi Copy.

Runtime cập nhật.

```text
sendx++

qcount++
```

Ví dụ.

```text
Capacity

4

-------------------

qcount

2

-------------------

sendx

1
```

Sau Send.

```text
qcount

3

-------------------

sendx

2
```

---

# Bước 6 — Wakeup Receiver

Sau khi Buffer có dữ liệu.

Runtime kiểm tra.

```text
recvq
```

Nếu có Receiver.

```text
recvq

↓

Pop

↓

goready()
```

Receiver trở thành:

```text
Runnable
```

Sau đó Scheduler quyết định khi nào Receiver chạy.

---

# Luồng Buffered Send

```text
             Sender

                │

                ▼

         Acquire Lock

                │

                ▼

      Buffer còn chỗ ?

          │         │

         No        Yes

          │         │

          ▼         ▼

     Park G     Copy Memory

                    │

                    ▼

             Ring Buffer

                    │

                    ▼

            Wake Receiver

                    │

                    ▼

                Unlock
```

---

# Khi Buffer đầy

Ví dụ.

```text
Capacity

4
```

```text
+----+----+----+----+

| A  | B  | C  | D  |

+----+----+----+----+
```

Sender tiếp theo.

```go
ch <- E
```

Runtime không thể Copy.

Thay vào đó.

```text
Create sudog

↓

sendq

↓

Park
```

Sender chuyển sang trạng thái:

```text
Waiting
```

---

# 15.9 Unbuffered Channel Send

Unbuffered Channel.

```go
make(chan int)
```

Capacity.

```text
0
```

Không tồn tại Ring Buffer.

Điều này làm thay đổi hoàn toàn thuật toán.

---

# Không có Buffer

Runtime không thể:

```text
Copy

↓

Buffer
```

Bởi vì.

```text
Buffer = nil
```

Chỉ còn hai khả năng.

- Có Receiver đang chờ.
- Không có Receiver.

---

# Receiver Waiting

Giả sử.

```go
<-ch
```

đã chạy trước.

Runtime có.

```text
recvq

↓

Receiver
```

Lúc này Sender.

```go
ch <- value
```

sẽ không đi qua Buffer.

---

# Direct Handoff

Runtime Copy trực tiếp.

```text
Sender Stack

↓

Receiver Stack
```

Không có vùng nhớ trung gian.

Đây gọi là:

> **Direct Handoff**

---

# Minh họa

```text
Sender

Value

↓

Copy

↓

Receiver
```

Không có.

```text
Ring Buffer
```

ở giữa.

---

# Runtime thực hiện

```text
Sender

↓

Receiver Waiting ?

↓

Yes

↓

Copy

↓

Wake Receiver

↓

Continue
```

Đây là Fast Path của Unbuffered Channel.

---

# Receiver chưa tồn tại

Nếu.

```text
recvq

↓

Empty
```

Runtime không thể Send.

Sender sẽ.

```text
Create sudog

↓

sendq

↓

Park
```

Đợi Receiver xuất hiện.

---

# Khi Receiver tới

Sau này.

```go
<-ch
```

Runtime sẽ.

```text
Pop Sender

↓

Direct Copy

↓

Wake Sender

↓

Wake Receiver
```

Hai Goroutine cùng tiếp tục.

---

# Synchronization Point

Đây là đặc điểm quan trọng nhất.

Sender.

```text
Waiting
```

Receiver.

```text
Waiting
```

Hai Goroutine phải gặp nhau.

```text
Sender

──────────────┐

              │

Synchronization Point

              │

──────────────┘

Receiver
```

Sau đó.

Runtime mới thực hiện.

```text
Copy Memory
```

Đây là lý do.

Unbuffered Channel đồng thời là.

```text
Synchronization Primitive
```

---

# Memory Visibility

Sau khi Send hoàn thành.

Go Memory Model đảm bảo.

```go
Sender

↓

Write X

↓

Send

↓

Receive

↓

Read X

Receiver luôn nhìn thấy giá trị mới nhất.
```

Điều này được Runtime đảm bảo bằng các cơ chế đồng bộ nội bộ và là một trong những lý do Channel được sử dụng như một công cụ đồng bộ hóa thay vì chỉ là một hàng đợi dữ liệu.

---

# So sánh Buffered và Unbuffered Send

| Buffered | Unbuffered |
|-----------|------------|
| Có Ring Buffer | Không có Buffer |
| Copy vào Buffer | Copy trực tiếp sang Receiver |
| Sender có thể tiếp tục ngay | Sender phải chờ Receiver (nếu chưa có Receiver) |
| Throughput cao | Đồng bộ mạnh |
| Có thể Queue nhiều phần tử | Không Queue dữ liệu |

---

# Toàn bộ Send Flow

```text
                    chansend()

                         │

                         ▼

                  Acquire Lock

                         │

                         ▼

                Channel Closed ?

                   │          │

                  Yes        No

                   │          │

                   ▼          ▼

                 Panic   Receiver Waiting ?

                              │

                    ┌─────────┴─────────┐

                    │                   │

                   Yes                 No

                    │                   │

                    ▼                   ▼

             Direct Handoff      Buffer Available ?

                                        │

                              ┌─────────┴─────────┐

                              │                   │

                             Yes                 No

                              │                   │

                              ▼                   ▼

                      Copy vào Buffer       Park Sender

                              │

                              ▼

                     Wake Receiver (nếu có)

                              │

                              ▼

                           Unlock

                              │

                              ▼

                            Return
```

---

# Key Takeaways

- Mọi phép `ch <- value` đều được Runtime xử lý thông qua `runtime.chansend()`.
- `chansend()` luôn ưu tiên **Fast Path**: Direct Handoff hoặc ghi trực tiếp vào Buffer.
- Với **Buffered Channel**, Runtime sao chép dữ liệu vào Ring Buffer bằng `memmove()`, sau đó cập nhật `sendx`, `qcount` và đánh thức Receiver nếu cần.
- Với **Unbuffered Channel**, dữ liệu được truyền trực tiếp từ Sender sang Receiver (**Direct Handoff**) mà không đi qua Buffer.
- Khi Channel không thể tiếp nhận dữ liệu, Sender được đóng gói thành một `sudog`, đưa vào `sendq` và bị `gopark()` cho đến khi có Receiver.
- Unbuffered Channel tạo ra một **Synchronization Point**, nơi Sender và Receiver phải gặp nhau trước khi dữ liệu được truyền. Đây cũng là cơ sở để Go Memory Model thiết lập quan hệ **happens-before** giữa hai Goroutine.

# Part IV — Receive Operation

Ở phần trước, chúng ta đã phân tích toàn bộ quá trình **Send** của Channel.

Trong phần này, chúng ta sẽ đi theo chiều ngược lại.

Khi chương trình thực hiện:

```go
v := <-ch
```

Runtime sẽ phải trả lời nhiều câu hỏi:

- Channel đã có dữ liệu chưa?
- Có Sender nào đang chờ không?
- Có cần Block Goroutine không?
- Có cần đánh thức Sender không?

Tất cả logic này đều được hiện thực trong:

```text
runtime.chanrecv()
```

---

# 15.10 Receive Flow

## Khi chương trình gọi

```go
v := <-ch
```

Compiler sẽ sinh lời gọi:

```text
runtime.chanrecv()
```

Không có phép đọc trực tiếp nào trên Channel.

Toàn bộ Receive đều đi qua Runtime.

---

# Tổng quan luồng Receive

```text
                chanrecv()

                     │

                     ▼

             Buffer có dữ liệu?

               │          │

             Yes         No

               │          │

               ▼          ▼

      Copy từ Buffer   Sender Waiting ?

                           │

                    ┌──────┴──────┐

                    │             │

                   Yes           No

                    │             │

                    ▼             ▼

             Direct Handoff    Park Receiver
```

Đây là thuật toán tổng quát của `chanrecv()`.

---

# Runtime thực hiện

Một phép Receive trải qua các bước:

```text
Acquire Lock

↓

Check Closed

↓

Check Buffer

↓

Check Sender

↓

Copy

↓

Wakeup

↓

Unlock
```

Nếu không có dữ liệu.

Runtime sẽ:

```text
Park Current Goroutine
```

---

# Fast Path

Runtime luôn ưu tiên:

```text
1. Buffer

↓

2. Waiting Sender

↓

3. Slow Path
```

Điều này giúp giảm Context Switch.

---

# 15.11 Buffered Receive

Giả sử.

```go
ch := make(chan int, 4)
```

Buffer.

```text
+----+----+----+----+

| 10 | 20 |    |    |

+----+----+----+----+
```

---

# Bước 1 — Acquire Lock

Runtime khóa Channel.

```text
lock(hchan)
```

---

# Bước 2 — Kiểm tra Buffer

Nếu.

```text
qcount > 0
```

Runtime biết.

```text
Có dữ liệu
```

Không cần Block.

---

# Bước 3 — Copy Memory

Runtime thực hiện.

```text
Buffer

↓

Receiver Stack
```

Ví dụ.

```go
v := <-ch
```

Runtime Copy.

```text
10

↓

v
```

Sau đó.

```text
recvx++

qcount--
```

---

# Minh họa

Trước Receive.

```text
+----+----+----+----+

|10  |20  |    |    |

+----+----+----+----+

recvx = 0
```

Sau Receive.

```text
+----+----+----+----+

|    |20  |    |    |

+----+----+----+----+

recvx = 1
```

---

# Bước 4 — Wakeup Sender

Sau khi Buffer có thêm chỗ.

Runtime kiểm tra.

```text
sendq
```

Nếu có Sender đang chờ.

```text
Pop sendq

↓

goready()
```

Sender tiếp tục chạy.

---

# Luồng Buffered Receive

```text
Receiver

↓

Acquire Lock

↓

Buffer Empty ?

│          │

No         Yes

│          │

▼          ▼

Copy      Park

↓

Update recvx

↓

Wake Sender

↓

Unlock
```

---

# Khi Buffer rỗng

Ví dụ.

```text
Capacity

4
```

```text
qcount = 0
```

Runtime không thể lấy dữ liệu.

Receiver sẽ:

```text
Create sudog

↓

recvq

↓

gopark()
```

---

# 15.12 Unbuffered Receive

Unbuffered Channel.

```go
make(chan int)
```

Không tồn tại Buffer.

```text
Buffer = nil
```

Runtime chỉ còn hai lựa chọn.

- Có Sender.
- Không có Sender.

---

# Sender Waiting

Nếu.

```text
sendq

↓

Sender
```

Runtime thực hiện.

```text
Direct Copy
```

---

# Direct Handoff

Không có Buffer.

Runtime Copy trực tiếp.

```text
Sender Stack

↓

Receiver Stack
```

Ví dụ.

```go
Sender

↓

100
```

```go
Receiver

↓

v
```

Không tồn tại Object trung gian.

---

# Wake Sender

Sau khi Copy.

Runtime.

```text
Pop sendq

↓

goready(sender)
```

Sender tiếp tục chạy.

---

# Sender chưa tồn tại

Nếu.

```text
sendq

↓

Empty
```

Receiver không thể lấy dữ liệu.

Runtime.

```text
Create sudog

↓

recvq

↓

gopark()
```

Đợi Sender xuất hiện.

---

# Khi Sender tới

Sau này.

```go
ch <- value
```

Runtime.

```text
Pop Receiver

↓

Copy

↓

Wake Receiver

↓

Wake Sender
```

Hai Goroutine tiếp tục.

---

# So sánh

Buffered.

```text
Buffer

↓

Receiver
```

Unbuffered.

```text
Sender

↓

Receiver
```

Không qua Buffer.

---

# 15.13 Zero Value

Một câu hỏi rất phổ biến.

Điều gì xảy ra nếu.

```go
close(ch)
```

sau đó.

```go
v := <-ch
```

---

# Receive từ Closed Channel

Ví dụ.

```go
ch := make(chan int)

close(ch)

v := <-ch
```

Runtime trả về.

```text
0
```

Không Panic.

---

# Tại sao?

Channel đã đóng.

Sẽ không còn dữ liệu.

Runtime trả về.

```text
Zero Value
```

của kiểu dữ liệu.

Ví dụ.

```go
int

↓

0
```

```go
bool

↓

false
```

```go
string

↓

""
```

---

# Dạng đầy đủ

Go cung cấp.

```go
v, ok := <-ch
```

Ví dụ.

```go
v, ok := <-ch
```

Nếu Channel còn mở.

```text
ok

↓

true
```

Nếu Channel đã đóng.

```text
ok

↓

false
```

---

# Minh họa

```go
for {

    v, ok := <-ch

    if !ok {

        break

    }

}
```

Đây là cách đọc Channel phổ biến nhất.

---

# Receive từ nil Channel

```go
var ch chan int

<-ch
```

Runtime.

```text
Forever Waiting
```

Không Panic.

Không Return.

---

# Receive từ Closed Channel có Buffer

Giả sử.

```text
Buffer

10

20
```

Sau khi.

```go
close(ch)
```

Runtime vẫn trả.

```text
10

↓

20

↓

Zero Value
```

Chỉ khi Buffer rỗng.

Mới trả về Zero Value.

---

# Tổng kết Receive

```text
Receive

↓

Buffer ?

│

├── Có

│      ↓

│   Copy

│

└── Không

       ↓

Sender Waiting ?

│

├── Có

│      ↓

│  Direct Copy

│

└── Không

       ↓

Park
```

---

# Part V — Blocking Mechanism

Channel không chỉ là nơi truyền dữ liệu.

Điều quan trọng hơn.

Channel là cơ chế:

```text
Blocking

↓

Wakeup

↓

Scheduling
```

Đây là nơi Channel kết nối trực tiếp với Scheduler.

---

# 15.14 Goroutine Parking

## Khi nào Runtime Park?

Ví dụ.

```go
<-ch
```

Nếu.

```text
No Sender
```

Runtime không thể tiếp tục.

Thay vì Busy Waiting.

Runtime.

```text
gopark()
```

---

# gopark()

Đây là hàm trung tâm của Scheduler.

Nó thực hiện.

```text
Current G

↓

Running

↓

Waiting
```

Sau đó.

```text
schedule()
```

chọn Goroutine khác.

---

# Minh họa

```text
Running

↓

gopark()

↓

Waiting

↓

Scheduler

↓

Next G
```

CPU không bị lãng phí.

---

# gopark() không Sleep Thread

Đây là điểm rất quan trọng.

Runtime không làm.

```text
Thread Sleep
```

Mà chỉ.

```text
Current Goroutine

↓

Waiting
```

Machine vẫn tiếp tục.

```text
Execute G khác
```

---

# 15.15 Wakeup

Khi dữ liệu tới.

Runtime cần đánh thức Goroutine.

Điều này được thực hiện bằng.

```text
goready()
```

---

# goready()

Runtime chuyển.

```text
Waiting

↓

Runnable
```

Không chạy ngay.

Chỉ chuyển trạng thái.

Sau đó.

```text
Local Queue
```

---

# Minh họa

```text
Waiting

↓

goready()

↓

Runnable

↓

Scheduler

↓

Running
```

---

# Wakeup không đồng nghĩa với Running

Đây là hiểu lầm phổ biến.

```text
Wakeup

≠

Running
```

Scheduler vẫn quyết định.

```text
Khi nào Execute
```

---

# 15.16 Send Queue (`sendq`)

Nếu Sender không thể Send.

Runtime tạo.

```text
sudog
```

Sau đó.

```text
sendq

↓

sudog

↓

goroutine
```

---

# Cấu trúc

```text
sendq

↓

G1

↓

G2

↓

G3
```

Đây là Queue của:

```text
Waiting Sender
```

---

# Khi Receiver tới

```text
Receiver

↓

Pop sendq

↓

Copy

↓

Wake Sender
```

---

# 15.17 Receive Queue (`recvq`)

Ngược lại.

Nếu Receiver không có dữ liệu.

Runtime.

```text
recvq

↓

sudog

↓

goroutine
```

---

# Minh họa

```text
recvq

↓

G8

↓

G9

↓

G10
```

---

# Khi Sender tới

```text
Sender

↓

Pop recvq

↓

Copy

↓

Wake Receiver
```

---

# Quan hệ giữa sendq và recvq

```text
                hchan

        +------------------------+

        Ring Buffer

        ------------------------

        sendq

            │

            ▼

        Waiting Sender

        ------------------------

        recvq

            │

            ▼

       Waiting Receiver

        +------------------------+
```

---

# Blocking Flow

Có thể mô tả toàn bộ cơ chế Block của Runtime như sau.

```text
Send / Receive

↓

Can Continue ?

│

├── Yes

│      ↓

│   Continue

│

└── No

       ↓

Create sudog

↓

sendq / recvq

↓

gopark()

↓

Waiting

↓

...

↓

goready()

↓

Runnable

↓

Scheduler

↓

Running
```

Đây là điểm kết nối giữa:

- Channel
- Scheduler
- Goroutine
- Processor

Mọi Goroutine Block trên Channel cuối cùng đều phải đi qua:

```text
gopark()

↓

Waiting

↓

goready()

↓

Runnable

↓

schedule()
```

---

# Key Takeaways

- Mọi phép Receive đều được Runtime xử lý thông qua `runtime.chanrecv()`.
- Với **Buffered Channel**, Receiver sao chép dữ liệu từ Ring Buffer, cập nhật `recvx` và `qcount`, sau đó có thể đánh thức Sender đang chờ.
- Với **Unbuffered Channel**, dữ liệu được truyền trực tiếp từ Sender sang Receiver (**Direct Handoff**) mà không cần vùng nhớ trung gian.
- Đọc từ **Closed Channel** trả về **Zero Value** của kiểu dữ liệu; sử dụng `v, ok := <-ch` để phân biệt Channel đã đóng hay chưa.
- Khi Channel không thể tiếp tục Send hoặc Receive, Runtime sử dụng `gopark()` để chuyển Goroutine sang trạng thái **Waiting** thay vì chặn OS Thread.
- Khi điều kiện đã thỏa mãn, Runtime gọi `goready()` để chuyển Goroutine sang **Runnable**; Scheduler sẽ quyết định thời điểm thực thi.
- `sendq` và `recvq` là hai hàng đợi lưu các `sudog` đại diện cho những Goroutine đang chờ gửi hoặc nhận dữ liệu.

# Part VI — Closing Channel

Ở các phần trước, chúng ta đã tìm hiểu cách Runtime xử lý:

- Send
- Receive
- Blocking
- Wakeup

Một câu hỏi quan trọng tiếp theo là:

> **Điều gì thực sự xảy ra khi gọi `close(channel)`?**

Rất nhiều lập trình viên cho rằng:

```go
close(ch)
```

sẽ:

- giải phóng bộ nhớ
- hủy Channel
- xóa Buffer

Thực tế hoàn toàn không phải như vậy.

Trong Go Runtime, `close()` chỉ là một thao tác **thay đổi trạng thái của Channel** và **đánh thức toàn bộ Goroutine đang chờ**.

Việc giải phóng bộ nhớ vẫn do Garbage Collector quyết định.

---

# 15.18 close()

## Hàm close()

Ví dụ.

```go
close(ch)
```

Compiler sẽ chuyển thành lời gọi Runtime.

```text
runtime.closechan()
```

Source code nằm tại:

```text
src/runtime/chan.go
```

---

# Runtime thực hiện những gì?

Có thể mô tả đơn giản.

```text
Acquire Lock

↓

Check nil

↓

Check Closed

↓

Mark Closed

↓

Wake Receiver

↓

Wake Sender

↓

Unlock
```

Toàn bộ thao tác chỉ diễn ra trong vài bước.

---

# Bước 1 — Acquire Lock

Đầu tiên Runtime khóa Channel.

```text
lock(hchan)
```

Điều này đảm bảo:

- không có Send
- không có Receive
- không có Close khác

được thực hiện đồng thời.

---

# Bước 2 — Kiểm tra nil

Ví dụ.

```go
var ch chan int

close(ch)
```

Runtime phát hiện.

```text
hchan == nil
```

Ngay lập tức.

```text
panic
```

Thông báo.

```text
panic: close of nil channel
```

---

# Bước 3 — Kiểm tra đã đóng chưa

Runtime kiểm tra.

```text
closed
```

Nếu.

```text
closed == 1
```

Điều này có nghĩa.

Channel đã đóng trước đó.

Runtime sẽ.

```text
panic
```

Thông báo.

```text
panic: close of closed channel
```

---

# Bước 4 — Đánh dấu Channel đã đóng

Nếu hợp lệ.

Runtime chỉ cần.

```text
closed = 1
```

Lưu ý.

Runtime **không**:

- xóa Buffer
- giải phóng bộ nhớ
- hủy `hchan`

Nó chỉ thay đổi một cờ trạng thái.

---

# Bước 5 — Đánh thức Receiver

Nếu.

```text
recvq

↓

G1

↓

G2

↓

G3
```

Runtime.

```text
Pop

↓

goready()

↓

Runnable
```

Toàn bộ Receiver đang chờ đều được đánh thức.

---

# Bước 6 — Đánh thức Sender

Tương tự.

```text
sendq

↓

G8

↓

G9
```

Runtime cũng.

```text
Wakeup
```

Tuy nhiên.

Các Sender này sẽ:

```text
panic
```

khi tiếp tục thực thi.

Chúng ta sẽ tìm hiểu ngay sau đây.

---

# Close Flow

```text
close()

↓

Acquire Lock

↓

Already Closed ?

│

├── Yes

│      ↓

│    panic

│

└── No

       ↓

closed = 1

↓

Wake recvq

↓

Wake sendq

↓

Unlock
```

---

# 15.19 Close Semantics

Đóng Channel không có nghĩa là "Channel biến mất".

Nó chỉ thay đổi cách Runtime xử lý các thao tác tiếp theo.

---

# Send vào Closed Channel

Ví dụ.

```go
ch := make(chan int)

close(ch)

ch <- 10
```

Runtime.

```text
panic
```

Thông báo.

```text
panic: send on closed channel
```

---

## Tại sao lại Panic?

Giả sử Runtime cho phép.

```text
Closed Channel

↓

Send
```

Receiver sẽ không biết.

```text
10

↓

Từ đâu?
```

Điều này phá vỡ ý nghĩa của việc đóng Channel.

Do đó.

Go Runtime lựa chọn.

```text
panic
```

để phát hiện lỗi lập trình càng sớm càng tốt.

---

# Receive từ Closed Channel

Ngược lại.

```go
v := <-ch
```

không Panic.

Runtime kiểm tra.

```text
Buffer còn dữ liệu?
```

---

# Buffer chưa rỗng

Ví dụ.

```text
10

20

30
```

Sau khi.

```go
close(ch)
```

Receiver vẫn nhận.

```text
10

↓

20

↓

30
```

Hoàn toàn bình thường.

---

# Buffer đã rỗng

Sau khi đọc hết.

Runtime trả về.

```text
Zero Value
```

Ví dụ.

```go
chan int
```

↓

```text
0
```

---

```go
chan bool
```

↓

```text
false
```

---

```go
chan string
```

↓

```text
""
```

---

# Multiple Receiver

Giả sử.

```text
recvq

↓

G1

↓

G2

↓

G3

↓

G4
```

Runtime gọi.

```go
close(ch)
```

Điều gì xảy ra?

Toàn bộ Receiver.

```text
Wakeup
```

Sau đó.

```text
Receive

↓

Zero Value
```

Không Goroutine nào bị bỏ sót.

---

# Multiple Sender

Nếu.

```text
sendq

↓

G1

↓

G2

↓

G3
```

Sau khi Wakeup.

Mỗi Sender.

```text
panic
```

Đây là lý do.

Thông thường chỉ nên có:

```text
Một Goroutine

↓

close()
```

---

# Vì sao Receiver không Panic?

Giả sử.

```go
for {

    v, ok := <-ch

}
```

Nếu Receiver Panic.

Sẽ rất khó xây dựng:

- Pipeline
- Worker Pool
- Broadcast

Do đó Runtime chọn.

```text
Receive Zero Value
```

để kết thúc luồng dữ liệu một cách tự nhiên.

---

# Memory Model

Sau khi.

```go
close(ch)
```

Runtime đảm bảo.

```text
close()

↓

happens-before

↓

Receive Zero Value
```

Điều này được Go Memory Model định nghĩa rõ ràng.

---

# 15.20 Closed Channel

Sau khi đóng.

Channel vẫn tồn tại.

```text
hchan

↓

closed = 1
```

Mọi thao tác sau đó đều dựa trên trạng thái này.

---

# Read

Ví dụ.

```go
v := <-ch
```

Có ba trường hợp.

---

## Trường hợp 1

Buffer còn dữ liệu.

```text
Read

↓

Buffer
```

---

## Trường hợp 2

Buffer rỗng.

```text
Read

↓

Zero Value
```

---

## Trường hợp 3

Dùng.

```go
v, ok := <-ch
```

Runtime trả.

```text
ok

↓

false
```

---

# Write

Ví dụ.

```go
ch <- 10
```

Runtime.

```text
panic
```

Không có ngoại lệ.

---

# Close lần hai

Ví dụ.

```go
close(ch)

close(ch)
```

Runtime.

```text
panic
```

Thông báo.

```text
close of closed channel
```

---

# nil Channel

nil Channel khác hoàn toàn Closed Channel.

---

## Read

```go
var ch chan int

<-ch
```

Runtime.

```text
Forever Waiting
```

---

## Write

```go
ch <- 1
```

Runtime.

```text
Forever Waiting
```

---

## Close

```go
close(ch)
```

Runtime.

```text
panic
```

---

# So sánh

| Thao tác | Open | Closed | nil |
|----------|------|---------|-----|
| Read | Có thể Block | Zero Value | Block mãi mãi |
| Write | Có thể Block | Panic | Block mãi mãi |
| Close | Thành công | Panic | Panic |

---

# Vì sao nil Channel Block mãi?

Một nil Channel.

```text
hchan

↓

nil
```

Runtime không có:

- Buffer
- sendq
- recvq

Không có nơi để Park hay Wakeup.

Do đó Goroutine sẽ chờ vô thời hạn.

Đây cũng là một tính năng được tận dụng rất nhiều trong `select`, chúng ta sẽ tìm hiểu ở Chapter 16.

---

# Sơ đồ trạng thái của Channel

```text
                 make(chan)

                     │

                     ▼

                 Open Channel

                     │

          +----------+----------+

          │                     │

          ▼                     ▼

      Send/Receive          close()

                                   │

                                   ▼

                             Closed Channel

                                   │

                     +-------------+-------------+

                     │                           │

                     ▼                           ▼

                 Read Zero Value             Write Panic

                                   │

                                   ▼

                              GC thu hồi
```

Lưu ý rằng:

- `close()` **không hủy Channel**.
- `close()` chỉ thay đổi trạng thái từ **Open** sang **Closed**.
- Bộ nhớ của `hchan` chỉ được giải phóng khi **không còn tham chiếu** và Garbage Collector thu hồi.

---

# Tổng kết Closing Channel

Toàn bộ quá trình đóng Channel có thể mô tả như sau.

```text
                close(ch)

                     │

                     ▼

              runtime.closechan()

                     │

                     ▼

              Acquire Lock

                     │

                     ▼

             Already Closed ?

               │             │

              Yes           No

               │             │

               ▼             ▼

            panic       closed = 1

                            │

            +---------------+---------------+

            │                               │

            ▼                               ▼

      Wake recvq                     Wake sendq

            │                               │

            ▼                               ▼

 Receive Zero Value                 Send Panic

            │

            ▼

          Unlock
```

---

# Key Takeaways

- `close(ch)` được Runtime hiện thực thông qua `runtime.closechan()`.
- Đóng Channel **không giải phóng bộ nhớ**; Runtime chỉ đánh dấu `closed = 1` và đánh thức các Goroutine đang chờ.
- Gửi dữ liệu vào **Closed Channel** luôn gây **panic** (`send on closed channel`).
- Đọc từ **Closed Channel** không panic; Runtime trả về dữ liệu còn lại trong Buffer, sau đó trả về **Zero Value**.
- Sử dụng `v, ok := <-ch` là cách chuẩn để phát hiện Channel đã đóng.
- Khi Channel đóng, **tất cả Receiver đang chờ đều được đánh thức** và có thể tiếp tục thực thi.
- `nil` Channel khác hoàn toàn `Closed Channel`: thao tác Send và Receive trên `nil` Channel sẽ Block vô thời hạn, còn `close(nil)` sẽ gây panic.
- `close()` là một cơ chế **broadcast** của Runtime: nó đồng thời đánh thức mọi Goroutine đang chờ trên Channel.

# Part VII — Memory Model

Trong các phần trước, chúng ta đã tìm hiểu:

- Cấu trúc của `hchan`
- Send
- Receive
- Blocking
- Wakeup
- Close

Tuy nhiên, vẫn còn một câu hỏi rất quan trọng.

Giả sử.

```go
var x int

go func() {

    x = 100

    ch <- struct{}{}

}()

<-ch

fmt.Println(x)
```

Tại sao chương trình **luôn luôn** in ra:

```text
100
```

Trong khi:

```go
x
```

không hề được bảo vệ bởi:

- Mutex
- RWMutex
- Atomic

Điều gì đảm bảo CPU không đọc giá trị cũ?

Điều gì ngăn Compiler reorder instruction?

Điều gì ngăn CPU cache giữ dữ liệu cũ?

Câu trả lời nằm ở:

> **Go Memory Model**

Channel không chỉ truyền dữ liệu.

Nó còn truyền **thứ tự thực thi (Execution Ordering)**.

Đây là một trong những khía cạnh quan trọng nhất của Go Runtime.

---

# 15.21 Happens-Before

## Memory Model là gì?

Trong hệ thống đa CPU.

Các thao tác:

```go
a = 1

b = 2
```

không nhất thiết được CPU thực hiện đúng theo thứ tự lập trình.

CPU có thể:

- Reorder Instruction
- Store Buffer
- Cache Write
- Out-of-order Execution

Ví dụ.

```go
a = 1

b = 2
```

CPU có thể nhìn thấy.

```text
Store b

↓

Store a
```

Điều này hoàn toàn hợp lệ ở mức phần cứng.

---

# Một ví dụ

Thread A.

```go
data = 100

ready = true
```

Thread B.

```go
if ready {

    fmt.Println(data)

}
```

Thoạt nhìn.

Ta nghĩ.

```text
100
```

sẽ luôn được in.

Thực tế.

CPU hoàn toàn có thể.

```text
ready = true

↓

data = 100
```

Thread B có thể thấy.

```text
ready == true

data == 0
```

Đây chính là nguyên nhân của rất nhiều Race Condition.

---

# Happens-Before

Để giải quyết vấn đề này.

Go định nghĩa khái niệm:

> **Happens-Before**

Nếu.

```text
A Happens-Before B
```

thì Runtime đảm bảo.

- mọi ghi (Write) trước A
- sẽ nhìn thấy bởi mọi đọc (Read) sau B

Nói cách khác.

```text
Write

↓

A

↓

B

↓

Read
```

Read luôn nhìn thấy dữ liệu mới nhất.

---

# Channel Synchronization

Go Memory Model định nghĩa.

> **Một phép Send trên Channel Happens-Before phép Receive tương ứng.**

Đây là một trong những quy tắc quan trọng nhất.

Ví dụ.

```go
var x int

go func() {

    x = 10

    ch <- struct{}{}

}()

<-ch

fmt.Println(x)
```

Memory Model đảm bảo.

```text
x = 10

↓

Send

↓

Receive

↓

Read x
```

Receiver **luôn** nhìn thấy:

```text
10
```

---

# Minh họa

```text
Goroutine A

x = 100

↓

Send

==========================

Channel

==========================

↓

Receive

↓

Read x

Goroutine B
```

Send và Receive tạo thành một điểm đồng bộ hóa.

---

# Buffered Channel

Điều này vẫn đúng.

Ví dụ.

```go
ch := make(chan int, 10)
```

```go
x = 200

ch <- 1
```

Sau này.

```go
<-ch

fmt.Println(x)
```

Runtime vẫn đảm bảo.

```text
Write x

↓

Send

↓

Receive

↓

Read x
```

---

# close()

Memory Model còn định nghĩa.

```text
close()

↓

Receive Zero Value
```

là một Happens-Before Relation.

Ví dụ.

```go
done = true

close(ch)
```

Sau này.

```go
<-ch

fmt.Println(done)
```

Receiver luôn nhìn thấy.

```text
true
```

---

# Happens-Before không chỉ dành cho dữ liệu Channel

Một hiểu lầm phổ biến.

Nhiều người nghĩ.

```text
Channel

↓

Chỉ đồng bộ dữ liệu gửi.
```

Không.

Runtime đồng bộ.

```text
Toàn bộ Memory trước Send.
```

Đó là lý do.

Ví dụ trên luôn đúng.

---

# Những quan hệ Happens-Before quan trọng

Go Memory Model định nghĩa nhiều quy tắc.

Ví dụ.

```text
Mutex Unlock

↓

Mutex Lock
```

---

```text
Once.Do

↓

Return
```

---

```text
Channel Send

↓

Channel Receive
```

---

```text
close()

↓

Receive Closed
```

---

Đây là nền tảng để xây dựng chương trình Concurrent đúng.

---

# 15.22 Memory Barrier

## Memory Barrier là gì?

CPU hiện đại sử dụng rất nhiều kỹ thuật tối ưu.

Ví dụ.

```text
Store Buffer

CPU Cache

Out-of-order Execution

Instruction Reorder
```

Những tối ưu này làm chương trình nhanh hơn.

Nhưng lại gây khó khăn cho Concurrent Programming.

---

# Ví dụ

CPU có thể.

```go
a = 1

b = 2
```

thành.

```text
Store b

↓

Store a
```

Điều này hoàn toàn hợp lệ.

---

# Runtime Insert Fence

Để ngăn việc Reorder gây sai lệch.

Go Runtime chèn:

> **Memory Fence (Memory Barrier)**

vào những vị trí quan trọng.

Ví dụ.

```text
Sender

↓

Store Data

↓

Memory Fence

↓

Wake Receiver
```

Receiver.

```text
Wakeup

↓

Memory Fence

↓

Load Data
```

Nhờ đó.

CPU không được phép:

- reorder qua Fence
- giữ dữ liệu cũ sau Fence

---

# Acquire và Release

Có hai loại Barrier phổ biến.

## Release Barrier

Trước khi Sender đánh thức Receiver.

Runtime đảm bảo.

```text
All Previous Write

↓

Visible
```

---

## Acquire Barrier

Sau khi Receiver được đánh thức.

Runtime đảm bảo.

```text
All Future Read

↓

Read Latest Data
```

---

# Minh họa

```text
Sender

Write Data

↓

Release Fence

↓

Wake Receiver

====================

Acquire Fence

↓

Read Data

Receiver
```

Đây chính là lý do Receiver luôn nhìn thấy dữ liệu mới nhất.

---

# Runtime đặt Barrier ở đâu?

Lập trình viên không cần tự viết.

Runtime tự động chèn Barrier trong:

- `chansend()`
- `chanrecv()`
- `closechan()`
- `Mutex`
- `RWMutex`
- `Once`
- Atomic Operation

Đó là lý do Channel trở thành một primitive đồng bộ hóa hoàn chỉnh.

---

# Chi phí của Memory Barrier

Barrier không miễn phí.

CPU phải:

- Flush Store Buffer.
- Đồng bộ Cache.
- Ngăn một số tối ưu Out-of-order Execution.

Do đó.

Barrier chỉ được Runtime đặt ở những vị trí cần thiết.

---

# 15.23 Atomic

## Channel có cần Lock ngoài không?

Đây là câu hỏi rất phổ biến.

Ví dụ.

```go
var ch = make(chan int)

go func() {

    ch <- 10

}()

go func() {

    ch <- 20

}()
```

Có cần.

```go
mutex.Lock()
```

không?

Câu trả lời.

> **Không.**

---

# Vì sao?

Toàn bộ trạng thái của Channel được bảo vệ bên trong Runtime.

```text
hchan

↓

mutex
```

Runtime đảm bảo.

- Send
- Receive
- Close

là Thread-safe.

---

# Lock nội bộ

Mỗi lần.

```go
ch <- value
```

Runtime.

```text
Acquire Lock

↓

Modify hchan

↓

Release Lock
```

Lập trình viên không cần Lock thêm.

---

# Atomic bên trong Runtime

Ngoài Mutex.

Runtime còn sử dụng rất nhiều Atomic Operation.

Ví dụ.

```text
atomic.Load

atomic.Store

atomic.CAS

atomic.Xadd
```

Những Atomic này được dùng để.

- kiểm tra trạng thái
- tối ưu Fast Path
- giảm Lock Contention
- đồng bộ Scheduler

---

# Tại sao không dùng Atomic cho toàn bộ Channel?

Một Channel không chỉ sửa:

```text
Một Integer
```

Nó còn sửa.

```text
Buffer

Queue

sendx

recvx

qcount

sudog

Wakeup
```

Đây là nhiều thao tác liên quan chặt chẽ với nhau.

Atomic đơn lẻ không thể đảm bảo tính nhất quán cho toàn bộ cấu trúc.

Do đó Runtime sử dụng:

- Mutex để bảo vệ trạng thái tổng thể của `hchan`.
- Atomic cho một số biến và đường đi tối ưu (Fast Path).

---

# Không cần Lock ngoài

Ví dụ.

```go
jobs <- job
```

Hoặc.

```go
job := <-jobs
```

Bạn không cần.

```go
mutex.Lock()
```

Nếu làm như vậy.

```text
Lock

↓

Channel Lock

↓

Unlock
```

Chỉ làm tăng chi phí.

---

# Khi nào vẫn cần Mutex?

Channel chỉ đồng bộ.

```text
Send

Receive

Close
```

Nếu nhiều Goroutine cùng sửa.

```go
counter++
```

Channel không giúp.

Ví dụ.

```go
counter++

jobs <- task
```

Channel chỉ đảm bảo.

```text
Send Happens-Before Receive
```

Không bảo vệ.

```text
counter
```

Nếu nhiều Goroutine cùng sửa `counter`, bạn vẫn cần:

- `sync.Mutex`
- `sync.RWMutex`
- Atomic Operation

---

# Channel Synchronization không thay thế mọi Synchronization

Đây là điểm rất quan trọng.

Channel rất mạnh.

Nhưng không phải.

```text
Universal Synchronization Primitive
```

Hãy dùng Channel khi:

- truyền Ownership
- Pipeline
- Event
- Notification

Hãy dùng Mutex khi:

- bảo vệ Shared State
- cập nhật biến dùng chung
- thao tác trên cấu trúc dữ liệu chia sẻ

---

# Memory Model của Channel

Có thể mô tả toàn bộ như sau.

```text
             Goroutine A

        Write Shared Data

                │

                ▼

         Channel Send

                │

        Release Fence

===============================

             Channel

===============================

        Acquire Fence

                │

                ▼

        Channel Receive

                │

                ▼

        Read Shared Data

             Goroutine B
```

Đây là điều khiến Channel không chỉ là một hàng đợi dữ liệu.

Nó còn là một **Memory Synchronization Primitive** được Runtime hiện thực trực tiếp.

---

# Tổng kết Memory Model

```text
              Write Data

                  │

                  ▼

          Channel Send

                  │

          Release Barrier

                  │

===============================

             Channel

===============================

          Acquire Barrier

                  │

                  ▼

         Channel Receive

                  │

                  ▼

             Read Data
```

Memory Barrier và Happens-Before là hai nền tảng giúp Go Runtime đảm bảo rằng các Goroutine nhìn thấy dữ liệu nhất quán mà không yêu cầu lập trình viên tự chèn các lệnh đồng bộ ở mức CPU.

---

# Key Takeaways

- **Channel Send Happens-Before Channel Receive** là một trong những quy tắc quan trọng nhất của Go Memory Model.
- Quan hệ Happens-Before đảm bảo mọi ghi bộ nhớ trước Send sẽ hiển thị cho Receiver sau Receive.
- Runtime tự động chèn **Memory Barrier (Acquire/Release Fence)** tại các điểm đồng bộ như `chansend()`, `chanrecv()` và `closechan()`.
- `close(ch)` cũng tạo ra quan hệ Happens-Before với các phép Receive nhận trạng thái Channel đã đóng.
- Channel đã được Runtime bảo vệ bằng cơ chế đồng bộ nội bộ; lập trình viên **không cần** bọc Send hoặc Receive bằng `sync.Mutex`.
- Runtime kết hợp **Mutex** và **Atomic Operation** để hiện thực Channel: Mutex bảo vệ trạng thái phức tạp của `hchan`, còn Atomic tối ưu một số đường đi nhanh và thao tác đồng bộ.
- Channel đồng bộ hóa việc **truyền dữ liệu**, nhưng **không tự động bảo vệ mọi Shared State**. Với các biến được nhiều Goroutine cùng truy cập và cập nhật, vẫn cần sử dụng Mutex hoặc Atomic phù hợp.

# Part VIII — Performance

Sau khi hiểu toàn bộ cách Channel hoạt động bên trong Runtime, chúng ta có thể trả lời một câu hỏi mà rất nhiều lập trình viên Go quan tâm:

> **Channel có nhanh không?**

Câu trả lời là:

> **Có, nhưng không phải miễn phí.**

Mỗi lần thực hiện:

```go
ch <- value
```

Runtime có thể phải thực hiện:

- Acquire Lock
- Memory Copy
- Memory Barrier
- Wakeup Goroutine
- Context Switch
- Scheduler

Nếu hiểu rõ những chi phí này, chúng ta sẽ biết khi nào nên:

- truyền Value
- truyền Pointer
- dùng Buffered Channel
- dùng Unbuffered Channel
- dùng Mutex thay vì Channel

---

# 15.24 Copy Cost

## Channel luôn Copy dữ liệu

Đây là điều cực kỳ quan trọng.

Giả sử.

```go
type User struct {
    ID   int64
    Name [128]byte
}
```

```go
u := User{}

ch <- u
```

Runtime **không lưu địa chỉ của `u`**.

Runtime thực hiện.

```text
Sender Stack

↓

memmove()

↓

Channel Buffer
```

Sau đó.

```text
Channel Buffer

↓

memmove()

↓

Receiver Stack
```

Như vậy.

Một phép Send thông thường phải trải qua:

```text
Copy lần 1

↓

Buffer
```

Sau đó.

```text
Copy lần 2

↓

Receiver
```

---

# Hai lần Copy

Giả sử.

```go
chan User
```

Luồng dữ liệu.

```text
User

↓

Copy

↓

Channel Buffer

↓

Copy

↓

Receiver
```

Đây là lý do.

Channel không phải:

```text
Zero Copy
```

---

# Big Struct

Ví dụ.

```go
type User struct {

    Name [1024]byte

    Address [1024]byte

    Email [1024]byte

}
```

Kích thước.

```text
≈ 3 KB
```

Nếu.

```go
ch <- user
```

Runtime phải Copy.

```text
3 KB

↓

Buffer

↓

3 KB

↓

Receiver
```

Tổng cộng.

```text
6 KB
```

cho mỗi lần truyền.

---

# Ví dụ Benchmark

```go
type Big struct {
    Data [4096]byte
}
```

```go
ch <- big
```

Runtime phải Copy.

```text
4096 Bytes

↓

4096 Bytes
```

Nếu truyền:

```text
1 triệu Message
```

Runtime phải Copy.

```text
≈ 8 GB dữ liệu
```

Đây là một chi phí rất lớn.

---

# Vì sao Runtime phải Copy?

Giả sử Runtime chỉ lưu Pointer.

```text
Sender

↓

Pointer

↓

Receiver
```

Sau khi Send.

Sender sửa.

```go
user.Name = "ABC"
```

Receiver sẽ nhìn thấy dữ liệu thay đổi.

Điều này phá vỡ:

- Ownership
- Memory Model
- Isolation

Do đó Runtime luôn Copy Value.

---

# Copy Cost tăng theo kích thước Object

Có thể hình dung.

```text
Object Size

↓

Copy Cost
```

| Kích thước | Chi phí |
|------------|---------|
| 8 Bytes | Rất nhỏ |
| 32 Bytes | Nhỏ |
| 128 Bytes | Trung bình |
| 1 KB | Lớn |
| 4 KB | Rất lớn |
| 64 KB | Cực lớn |

---

# Rule of Thumb

Nếu Object lớn.

Đừng truyền.

```go
chan BigStruct
```

Hãy cân nhắc.

```go
chan *BigStruct
```

---

# 15.25 Pointer vs Value

Một trong những quyết định quan trọng nhất khi thiết kế API.

Là truyền:

```go
chan User
```

hay.

```go
chan *User
```

---

# Value Channel

Ví dụ.

```go
ch := make(chan User)
```

Luồng dữ liệu.

```text
User

↓

Copy

↓

Buffer

↓

Copy

↓

Receiver
```

Ưu điểm.

- Không chia sẻ dữ liệu.
- An toàn.
- Không Race.
- Ownership rõ ràng.

---

# Pointer Channel

```go
ch := make(chan *User)
```

Runtime chỉ Copy.

```text
Pointer

8 Bytes
```

Không Copy toàn bộ Struct.

```text
Pointer

↓

Buffer

↓

Pointer

↓

Receiver
```

Chi phí nhỏ hơn rất nhiều.

---

# So sánh

Giả sử.

```go
User

4 KB
```

Value.

```text
Copy

4096 Bytes
```

Pointer.

```text
Copy

8 Bytes
```

Khác biệt rất lớn.

---

# Nhưng Pointer có nhược điểm

Giả sử.

```go
go func() {

    ch <- user

}()
```

Receiver.

```go
u := <-ch

u.Name = "ABC"
```

Sender.

```go
user.Name = "XYZ"
```

Hai Goroutine.

```text
Shared Object
```

Có thể Race.

---

# Value Ownership

Value.

```text
Sender

↓

Copy

↓

Receiver
```

Hai Object độc lập.

---

# Pointer Ownership

```text
Sender

↓

Pointer

↓

Receiver
```

Hai Goroutine cùng nhìn một Object.

---

# Khi nào dùng Value?

Nên dùng khi.

Object nhỏ.

Ví dụ.

```go
int

bool

time.Time

struct < 64 Bytes
```

---

# Khi nào dùng Pointer?

Khi.

- Object lớn.
- Immutable Object.
- Chỉ đọc.
- Cache.
- Database Model.

---

# Quy tắc thực tế

| Object | Nên dùng |
|----------|-----------|
| int | Value |
| bool | Value |
| string | Value *(header nhỏ, dữ liệu dùng chung)* |
| time.Time | Value |
| Struct nhỏ | Value |
| Struct vài KB | Pointer |
| Struct vài MB | Pointer |

Lưu ý rằng `string` là một kiểu đặc biệt: giá trị `string` chỉ gồm một header nhỏ (con trỏ + độ dài), còn dữ liệu chuỗi được dùng chung và bất biến.

---

# 15.26 Buffer Size

Một câu hỏi rất phổ biến.

```go
make(chan Job, N)
```

Nên chọn:

```text
N = ?
```

Không có câu trả lời chung.

---

# Buffer nhỏ

Ví dụ.

```go
make(chan Job, 1)
```

Ưu điểm.

- Memory nhỏ.
- Latency thấp.
- Back Pressure tốt.

Nhược điểm.

Producer dễ Block.

---

# Buffer lớn

```go
make(chan Job, 100000)
```

Ưu điểm.

Producer gần như không Block.

---

Nhược điểm.

- Memory lớn.
- GC nhiều hơn.
- Queue dài.
- Latency tăng.

---

# Throughput

Buffer lớn.

```text
Producer

↓↓↓↓↓

Buffer

↓↓↓↓↓

Consumer
```

Producer ít phải chờ.

Throughput tăng.

---

# Latency

Giả sử.

```text
10000 Job

↓

Buffer
```

Job cuối cùng.

Có thể phải chờ rất lâu.

Latency tăng.

---

# Minh họa

Buffer nhỏ.

```text
Producer

↓

Consumer
```

Latency thấp.

---

Buffer lớn.

```text
Producer

↓

Queue

↓

Consumer
```

Job phải chờ.

---

# Back Pressure

Buffer nhỏ.

```text
Producer

↓

Blocked
```

Đây gọi là.

```text
Back Pressure
```

Giúp hệ thống.

- ổn định
- không tạo quá nhiều dữ liệu

---

# Buffer không phải Cache

Nhiều người dùng.

```go
make(chan Job, 1000000)
```

như Queue.

Đây là thiết kế không tốt.

Channel không phải Message Queue.

Nếu Queue quá lớn.

Hãy dùng.

- Kafka
- NATS
- RabbitMQ

Hoặc Queue trong Database.

---

# Chọn Buffer như thế nào?

Không có công thức cố định.

Thông thường.

```text
CPU Worker

↓

2x

↓

4x

↓

Benchmark
```

Luôn Benchmark.

Không đoán.

---

# 15.27 False Sharing

Đây là chủ đề rất ít tài liệu Go đề cập.

Nhưng lại ảnh hưởng lớn tới hiệu năng.

---

# CPU Cache Line

CPU không đọc.

```text
1 Byte
```

CPU đọc.

```text
Cache Line

64 Bytes
```

---

# Ví dụ

```text
Cache Line

+--------------------------------------+

Counter A

Counter B

+--------------------------------------+
```

Hai Goroutine.

```text
CPU1

↓

Counter A
```

```text
CPU2

↓

Counter B
```

Mặc dù.

```text
A

↓

Khác

↓

B
```

Nhưng chúng nằm chung một Cache Line.

---

# Điều gì xảy ra?

CPU1.

```text
Write A
```

CPU2.

```text
Write B
```

Cache phải.

```text
Invalidate

↓

Sync

↓

Invalidate

↓

Sync
```

Hiệu năng giảm mạnh.

Đây gọi là.

> **False Sharing**

---

# Channel có gặp không?

Bản thân Channel đã được Runtime tối ưu khá tốt.

Tuy nhiên.

Nếu nhiều Goroutine cùng thao tác trên:

```text
Một Channel
```

Thì.

```text
hchan.lock

↓

Cache Line

↓

Contention
```

vẫn có thể xảy ra.

Đặc biệt.

```text
100 CPU

↓

1 Channel
```

---

# Ví dụ

```text
100 Producer

↓

1 Channel

↓

1 Consumer
```

Toàn bộ Producer.

```text
Acquire Lock

↓

Release Lock
```

trên cùng một `hchan`.

Cache Line liên tục bị đồng bộ giữa các CPU.

---

# Cách giảm False Sharing

## Chia nhỏ Channel

Thay vì.

```text
100 Producer

↓

1 Channel
```

Có thể.

```text
25 Producer

↓

Channel 1

----------------

25 Producer

↓

Channel 2

...
```

---

## Worker Local Queue

Hoặc.

```text
Worker

↓

Local Queue

↓

Merge
```

Giảm Lock Contention.

---

## Sharding

Ví dụ.

```go
channels[hash(id)%N]
```

Thay vì.

```text
One Global Channel
```

---

# Tổng kết Performance

Có thể mô hình hóa chi phí của một phép Send như sau.

```text
Sender

↓

Acquire Lock

↓

Memory Copy

↓

Memory Barrier

↓

Update Buffer

↓

Wakeup

↓

Scheduler

↓

Receiver
```

Nếu Object lớn.

```text
Memory Copy

↑↑↑
```

Nếu nhiều CPU.

```text
Lock Contention

↑↑↑
```

Nếu Buffer quá lớn.

```text
Latency

↑↑↑
```

Nếu Buffer quá nhỏ.

```text
Blocking

↑↑↑
```

Hiệu năng của Channel luôn là sự cân bằng giữa:

- Throughput
- Latency
- Memory
- Synchronization

---

# Key Takeaways

- Channel **luôn sao chép dữ liệu** khi truyền giá trị; với Buffered Channel thường có hai lần sao chép: Sender → Buffer và Buffer → Receiver.
- Truyền `BigStruct` qua Channel có thể rất tốn chi phí do lượng dữ liệu phải sao chép tăng theo kích thước Object.
- `chan Value` an toàn và rõ ràng về Ownership nhưng tốn chi phí Copy; `chan *Value` giảm Copy nhưng cần cẩn thận với Shared State và Race Condition.
- Kích thước Buffer ảnh hưởng trực tiếp đến **Throughput**, **Latency**, **Memory Usage** và **Back Pressure**; không có kích thước tối ưu cho mọi trường hợp.
- Channel không phải là một Message Queue dung lượng lớn; Buffer quá lớn có thể làm tăng độ trễ và áp lực lên Garbage Collector.
- Khi nhiều Goroutine cùng truy cập một Channel, Lock Contention và **False Sharing** có thể trở thành điểm nghẽn hiệu năng.
- Hãy Benchmark và Profile trước khi tối ưu; lựa chọn Value hay Pointer, Buffer lớn hay nhỏ nên dựa trên đặc điểm workload thực tế thay vì các quy tắc cứng nhắc.

# Part IX — Engineering

Sau khi hoàn thành toàn bộ chương này, chúng ta đã hiểu Channel từ nhiều góc độ:

- Triết lý thiết kế
- Runtime Data Structure
- `hchan`
- Ring Buffer
- `sudog`
- Send
- Receive
- Blocking
- Wakeup
- Close
- Memory Model
- Performance

Trong phần cuối cùng, chúng ta sẽ chuyển từ góc nhìn **Runtime Engineer** sang **Software Engineer**.

Mục tiêu không còn là:

> "Runtime hoạt động như thế nào?"

mà là:

> "Làm thế nào để sử dụng Channel đúng trong Production?"

Đây cũng là những kiến thức thường xuyên xuất hiện trong các buổi phỏng vấn Middle, Senior và Staff Backend Engineer.

---

# 15.34 Best Practices

## Best Practice 1 — Dùng Channel để truyền Ownership

Triết lý quan trọng nhất của Go là:

> **Share Memory by Communicating**

Ví dụ.

```go
jobs <- job
```

Sau khi gửi.

```text
Producer

↓

Ownership

↓

Consumer
```

Thay vì:

```text
Producer

↓

Shared Object

↓

Consumer
```

Điều này giúp giảm đáng kể khả năng xảy ra Race Condition.

---

## Best Practice 2 — Không dùng Channel để thay thế mọi Mutex

Đây là sai lầm phổ biến.

Ví dụ.

```go
counter++
```

Nhiều người tạo.

```go
counterChan := make(chan int)
```

để bảo vệ biến.

Điều này thường làm chương trình:

- phức tạp hơn
- chậm hơn
- khó đọc hơn

Nếu chỉ cần bảo vệ Shared State.

```go
sync.Mutex
```

là lựa chọn phù hợp hơn.

---

## Best Practice 3 — Chọn đúng Primitive

Có thể tham khảo quy tắc sau.

| Bài toán | Nên dùng |
|----------|----------|
| Shared Counter | Mutex |
| Cache | RWMutex |
| Worker Pool | Channel |
| Pipeline | Channel |
| Event Notification | Channel |
| Ownership Transfer | Channel |
| Concurrent Map | Mutex hoặc `sync.Map` |

Đừng cố gắng giải mọi bài toán bằng Channel.

---

## Best Practice 4 — Thiết kế Pipeline rõ ràng

Ví dụ.

```text
Reader

↓

Parser

↓

Validator

↓

Writer
```

Mỗi Stage.

- chỉ có một trách nhiệm.
- giao tiếp bằng Channel.
- không truy cập dữ liệu nội bộ của Stage khác.

Pipeline rõ ràng sẽ:

- dễ mở rộng.
- dễ kiểm thử.
- dễ tối ưu.

---

## Best Practice 5 — Định nghĩa Channel Ownership

Một nguyên tắc rất quan trọng.

> **Ai tạo Channel thì người đó chịu trách nhiệm đóng Channel.**

Ví dụ.

```go
func producer() <-chan Job
```

Producer.

- tạo Channel.
- gửi dữ liệu.
- gọi `close()`.

Consumer.

Chỉ:

```go
range ch
```

Không bao giờ.

```go
close(ch)
```

Điều này tránh rất nhiều lỗi.

---

## Best Practice 6 — Luôn dùng Context để hủy Goroutine

Ví dụ.

```go
ctx, cancel := context.WithCancel(...)
```

Worker.

```go
select {

case <-ctx.Done():

    return

}
```

Không nên.

```go
for {

    job := <-jobs

}
```

Nếu Channel không bao giờ đóng.

Worker sẽ tồn tại mãi mãi.

---

## Best Practice 7 — Tránh Goroutine Leak

Ví dụ.

```go
go func() {

    <-ch

}()
```

Nếu.

```text
Không ai Send

Không ai Close
```

Goroutine.

```text
Waiting Forever
```

Đây gọi là.

```text
Goroutine Leak
```

Trong Production.

Leak vài nghìn Goroutine có thể gây:

- tăng Memory.
- Scheduler chậm.
- GC nhiều hơn.

---

## Best Practice 8 — Đóng Channel để phát tín hiệu hoàn thành

Ví dụ.

```go
close(done)
```

thay vì.

```go
done <- true
```

Lợi ích.

Toàn bộ Receiver.

```text
Wakeup
```

Đây là Broadcast miễn phí.

---

## Best Practice 9 — Thiết kế Buffer dựa trên Benchmark

Không nên.

```go
make(chan Job, 1000000)
```

chỉ vì.

```text
"Sợ Block"
```

Hãy đo:

- Throughput
- Latency
- Memory

Sau đó chọn Buffer phù hợp.

---

# 15.35 Common Mistakes

## Sai lầm 1 — Gửi vào Channel đã đóng

Ví dụ.

```go
close(ch)

ch <- 10
```

Runtime.

```text
panic
```

Đây là lỗi phổ biến nhất.

---

## Sai lầm 2 — Đóng Channel nhiều lần

Ví dụ.

```go
close(ch)

close(ch)
```

Runtime.

```text
panic
```

Nếu nhiều Goroutine có thể đóng Channel, hãy thiết kế lại Ownership hoặc sử dụng cơ chế đồng bộ phù hợp.

---

## Sai lầm 3 — Đọc từ nil Channel

Ví dụ.

```go
var ch chan int

<-ch
```

Runtime.

```text
Forever Waiting
```

Đây là nguyên nhân gây Deadlock rất khó debug.

---

## Sai lầm 4 — Buffer quá lớn

Ví dụ.

```go
make(chan Job, 1000000)
```

Tăng:

- Memory.
- GC.
- Latency.

Không cải thiện hiệu năng trong nhiều trường hợp.

---

## Sai lầm 5 — Buffer quá nhỏ

Ví dụ.

```go
make(chan Job, 1)
```

Producer.

```text
Block liên tục
```

Throughput giảm.

---

## Sai lầm 6 — Dùng Channel để bảo vệ Shared State

Ví dụ.

```go
counterChan
```

để thay cho.

```go
sync.Mutex
```

Đây thường không phải thiết kế tốt.

---

## Sai lầm 7 — Quên đóng Channel

Ví dụ.

```go
for v := range ch {

}
```

Nếu.

```text
Không Close
```

Receiver.

```text
Waiting Forever
```

---

## Sai lầm 8 — Không xử lý Context Cancellation

Worker.

```go
for {

    job := <-jobs

}
```

Nếu Service Shutdown.

Worker vẫn tồn tại.

---

## Sai lầm 9 — Deadlock do thiếu Receiver

Ví dụ.

```go
ch <- 10
```

Không có Receiver.

Runtime.

```text
fatal error:

all goroutines are asleep
```

---

## Sai lầm 10 — Deadlock do thiếu Sender

Ví dụ.

```go
<-ch
```

Không có Sender.

Runtime.

```text
Deadlock
```

---

## Sai lầm 11 — Quên kiểm tra `ok`

Ví dụ.

```go
v := <-ch
```

Sau khi.

```go
close(ch)
```

Bạn không biết.

```text
0
```

là dữ liệu thật.

Hay.

```text
Zero Value
```

Hãy dùng.

```go
v, ok := <-ch
```

---

# 15.36 Interview Questions

Dưới đây là những câu hỏi rất phổ biến trong các buổi phỏng vấn Backend Engineer, Go Developer và Distributed Systems.

---

## Nhóm cơ bản

### 1.

Channel là gì?

---

### 2.

Buffered Channel và Unbuffered Channel khác nhau như thế nào?

---

### 3.

Channel có Thread-safe không?

---

### 4.

Tại sao Send có thể Block?

---

### 5.

Receive Block khi nào?

---

### 6.

Điều gì xảy ra khi Buffer đầy?

---

### 7.

Điều gì xảy ra khi Buffer rỗng?

---

### 8.

Điều gì xảy ra khi đóng Channel?

---

### 9.

Điều gì xảy ra khi Send vào Closed Channel?

---

### 10.

Điều gì xảy ra khi Receive từ Closed Channel?

---

## Nhóm Runtime

### 11.

`hchan` gồm những field nào?

---

### 12.

`qcount` dùng để làm gì?

---

### 13.

`sendx` và `recvx` là gì?

---

### 14.

Tại sao Channel dùng Ring Buffer?

---

### 15.

`sudog` là gì?

---

### 16.

Tại sao Runtime cần `sendq` và `recvq`?

---

### 17.

Khi nào Runtime gọi `gopark()`?

---

### 18.

`goready()` làm gì?

---

### 19.

Tại sao Wakeup không có nghĩa là Running?

---

### 20.

Tại sao Unbuffered Channel không cần Buffer?

---

## Nhóm Memory Model

### 21.

Channel thiết lập Happens-Before như thế nào?

---

### 22.

Memory Barrier được đặt ở đâu?

---

### 23.

Tại sao Receiver luôn nhìn thấy dữ liệu mới nhất?

---

### 24.

Tại sao Channel không cần Lock ngoài?

---

## Nhóm Performance

### 25.

Tại sao `chan BigStruct` chậm?

---

### 26.

Khi nào nên dùng `chan User`?

Khi nào nên dùng.

```go
chan *User
```

---

### 27.

Buffer lớn có luôn nhanh hơn không?

---

### 28.

False Sharing ảnh hưởng thế nào đến Channel?

---

### 29.

Tại sao nhiều Producer dùng chung một Channel có thể làm giảm hiệu năng?

---

### 30.

Làm thế nào để Benchmark Channel?

---

## Nhóm Design

### 31.

Khi nào nên dùng Channel thay vì Mutex?

---

### 32.

Khi nào không nên dùng Channel?

---

### 33.

Ai nên gọi `close()`?

---

### 34.

Làm sao tránh Goroutine Leak?

---

### 35.

Thiết kế Worker Pool bằng Channel như thế nào?

---

### 36.

Thiết kế Pipeline bằng Channel như thế nào?

---

### 37.

Làm thế nào để Graceful Shutdown với Context và Channel?

---

### 38.

Tại sao `close(done)` thường tốt hơn `done <- struct{}{}` trong cơ chế broadcast?

---

### 39.

Tại sao `range ch` tự động kết thúc khi Channel đóng?

---

### 40.

Nếu phải tối ưu một hệ thống sử dụng hàng triệu phép Send/Receive mỗi giây, bạn sẽ đo những chỉ số nào và tối ưu ở đâu?

---

# 15.37 Chapter Summary

Trong chương này, chúng ta đã nghiên cứu Channel từ góc nhìn của **Go Runtime**, thay vì chỉ dừng lại ở cú pháp của ngôn ngữ.

Chúng ta bắt đầu bằng câu hỏi:

> **Tại sao Go cần Channel?**

Thông qua triết lý:

> **Share Memory by Communicating**

Go khuyến khích truyền Ownership giữa các Goroutine thay vì cùng chia sẻ và sửa đổi một vùng nhớ.

Tiếp theo, chúng ta đi sâu vào cấu trúc dữ liệu quan trọng nhất của Runtime:

```text
hchan
```

và phân tích vai trò của từng thành phần:

- Ring Buffer
- `sendq`
- `recvq`
- `sudog`
- Lock nội bộ
- Metadata

Sau đó, chúng ta lần lượt phân tích toàn bộ vòng đời của một phép:

- Send
- Receive
- Block
- Wakeup
- Close

dưới góc nhìn của các hàm:

- `chansend()`
- `chanrecv()`
- `closechan()`
- `gopark()`
- `goready()`

Tiếp theo là phần **Memory Model**, nơi chúng ta hiểu rằng Channel không chỉ truyền dữ liệu mà còn thiết lập quan hệ **Happens-Before** thông qua Memory Barrier do Runtime tự động chèn.

Ở phần **Performance**, chúng ta thấy rằng hiệu năng của Channel chịu ảnh hưởng bởi nhiều yếu tố:

- Copy Cost
- Kích thước Object
- Buffer Size
- Lock Contention
- False Sharing

Từ đó, chúng ta có thể đưa ra các quyết định thiết kế hợp lý hơn khi lựa chọn giữa:

- `chan Value`
- `chan *Value`
- Channel
- Mutex

Cuối cùng, chương kết thúc bằng các Best Practices, Common Mistakes và hệ thống câu hỏi phỏng vấn giúp chuyển hóa kiến thức Runtime thành kỹ năng áp dụng trong thực tế.

---

# Những điều quan trọng cần ghi nhớ

Sau khi hoàn thành chương này, bạn nên nắm vững những nguyên tắc sau:

- Channel là một **message passing primitive**, không chỉ là một Queue.
- `hchan` là cấu trúc dữ liệu cốt lõi hiện thực hóa Channel trong Go Runtime.
- Buffered Channel sử dụng **Ring Buffer**, còn Unbuffered Channel sử dụng **Direct Handoff** giữa Sender và Receiver.
- `sendq`, `recvq` và `sudog` là nền tảng cho cơ chế Blocking và Wakeup.
- `gopark()` chuyển Goroutine sang trạng thái **Waiting**, còn `goready()` chuyển Goroutine sang **Runnable**.
- `close()` chỉ thay đổi trạng thái của Channel và đánh thức các Goroutine đang chờ; việc giải phóng bộ nhớ vẫn do Garbage Collector đảm nhiệm.
- Channel thiết lập quan hệ **Happens-Before**, đảm bảo tính nhất quán của dữ liệu giữa các Goroutine.
- Hiệu năng của Channel phụ thuộc vào Copy Cost, Buffer Size, Lock Contention và thiết kế hệ thống.
- Không phải mọi bài toán đồng bộ đều nên sử dụng Channel; hãy chọn đúng primitive cho đúng bài toán.

---

# Chuẩn bị cho Chapter 16

Trong Chapter 15, chúng ta tập trung vào **một Channel**.

Nhưng trong các hệ thống thực tế, Goroutine thường phải chờ **nhiều sự kiện cùng lúc**.

Ví dụ.

- Nhận Job.
- Timeout.
- Context Cancellation.
- Shutdown Signal.
- Nhiều Channel đầu vào.

Go giải quyết bài toán này bằng:

```go
select
```

Chapter 16 sẽ phân tích toàn bộ cơ chế hoạt động của `select` từ góc nhìn Runtime, bao gồm:

- cách Runtime xây dựng danh sách các case,
- thuật toán lựa chọn ngẫu nhiên (pseudo-random selection),
- cơ chế đăng ký nhiều `sudog` trên nhiều Channel,
- cách tránh Starvation,
- và phân tích chi tiết mã nguồn trong `runtime/select.go`.

Đây sẽ là chương kết nối trực tiếp với những kiến thức về `Channel` vừa học và là bước tiếp theo để hiểu sâu cơ chế đồng bộ hóa trong Go Runtime.