# Part I — Tại sao Go cần Processor (`P`)

> "Nếu Goroutine (`G`) đại diện cho công việc và Machine (`M`) đại diện cho OS Thread, thì tại sao Go Runtime lại cần thêm một thành phần thứ ba là Processor (`P`)?"

Đây là một trong những quyết định thiết kế quan trọng nhất trong lịch sử phát triển của Go Runtime. Nếu không có `P`, Go sẽ rất khó đạt được hiệu năng và khả năng mở rộng (Scalability) như hiện nay.

Trong phần này, chúng ta sẽ tìm hiểu:

- Vì sao mô hình chỉ gồm `G` và `M` không đủ.
- Những vấn đề mà Go Runtime gặp phải trước Go 1.1.
- Cách `Processor (P)` giải quyết các vấn đề đó.
- Lịch sử phát triển của Scheduler từ mô hình GM sang GMP.

---

# 13.1 Why P Exists

## Một câu hỏi rất tự nhiên

Nếu nhìn vào mô hình ban đầu của Go Runtime, chúng ta có thể nghĩ rằng chỉ cần hai thành phần là đủ:

```text
Goroutine (G)
      │
      ▼
Machine (M)
      │
      ▼
OS Thread
      │
      ▼
CPU
```

Trong mô hình này:

- Goroutine là đơn vị thực thi.
- Machine (`M`) là một OS Thread.
- Scheduler chỉ cần gán một Goroutine cho một Thread để chạy.

Thoạt nhìn, mô hình này khá đơn giản và có vẻ hợp lý.

Vậy tại sao Go lại phải tạo thêm một thành phần mới?

---

## Mô hình GM hoạt động như thế nào?

Giả sử Runtime chỉ có `G` và `M`.

```text
G1 ─────────► M1

G2 ─────────► M2

G3 ─────────► M3

G4 ─────────► M4
```

Khi một Goroutine kết thúc, Thread sẽ lấy Goroutine tiếp theo từ Scheduler.

Nếu chỉ có vài chục Goroutine thì mô hình này hoạt động khá tốt.

Tuy nhiên, khi số lượng Goroutine tăng lên hàng trăm nghìn hoặc hàng triệu, hàng loạt vấn đề bắt đầu xuất hiện.

---

# Vấn đề thứ nhất: Global Scheduler Lock

Giả sử Scheduler chỉ có một hàng đợi toàn cục.

```text
          Global Run Queue

+---------------------------------------+
| G1 G2 G3 G4 G5 G6 G7 G8 ...           |
+---------------------------------------+

        ▲        ▲        ▲
        │        │        │
       M1       M2       M3
```

Mỗi Thread muốn lấy Goroutine đều phải thực hiện:

```go
lock(globalQueue)

g := pop(globalQueue)

unlock(globalQueue)
```

Điều gì xảy ra nếu máy có:

- 8 CPU Core
- 8 Thread

Khi đó cả 8 Thread đều phải tranh chấp cùng một mutex.

```text
CPU 1

lock()

------------------------

CPU 2

wait...

------------------------

CPU 3

wait...

------------------------

CPU 4

wait...
```

Thay vì chạy song song, Scheduler lại bị tuần tự hóa (Serialized) bởi một mutex duy nhất.

Đây là nút thắt hiệu năng đầu tiên.

---

# Vấn đề thứ hai: Thread không ổn định

Thread không phải là tài nguyên ổn định.

Một Thread có thể:

- bị block bởi System Call
- bị Runtime tạo mới
- bị Runtime thu hồi
- bị OS Scheduler chuyển CPU

Nếu Runtime lưu toàn bộ trạng thái Scheduler trên Thread thì mọi thứ sẽ trở nên rất phức tạp.

Ví dụ:

```text
Thread

├── Run Queue
├── Memory Cache
├── Timer
├── GC State
└── Scheduler State
```

Khi Thread bị block:

```text
Thread
    │
    ▼

System Call

    │
    ▼

Blocked
```

Toàn bộ trạng thái trên Thread cũng bị "đóng băng".

Điều này khiến Scheduler mất khả năng điều phối linh hoạt.

---

# Vấn đề thứ ba: Memory Cache

Go Runtime có một bộ cấp phát bộ nhớ (Memory Allocator) rất tối ưu.

Nếu Memory Cache được gắn với Thread:

```text
Thread A

└── Memory Cache
```

Khi Thread chết:

```text
Thread

↓

Destroy

↓

Memory Cache mất theo
```

Điều này dẫn đến:

- Fragmentation
- Cache Miss
- Tăng áp lực lên Memory Allocator

Go Runtime cần một nơi ổn định hơn để lưu Memory Cache.

---

# Vấn đề thứ tư: Work Stealing

Giả sử có hai Thread.

```text
Thread 1

100 Goroutine

----------------------

Thread 2

0 Goroutine
```

Nếu Queue gắn với Thread thì:

```text
Thread 2

↓

Muốn lấy việc

↓

Phải lấy từ Thread 1
```

Nhưng Thread 1 có thể:

- đang Block
- đang Sleep
- đang bị OS chuyển CPU

Scheduler rất khó cân bằng tải.

---

# Vấn đề thứ năm: Cache Locality

Giả sử Goroutine tạo thêm nhiều Goroutine con.

```go
func worker() {

    go task1()

    go task2()

    go task3()
}
```

Thông thường:

- các Goroutine này sử dụng cùng dữ liệu
- cùng Memory
- cùng Cache

Nếu Scheduler luôn đưa chúng lên Global Queue:

```text
CPU 1

↓

Global Queue

↓

CPU 7
```

CPU khác sẽ phải đọc lại toàn bộ dữ liệu từ RAM.

Kết quả là:

- Cache Miss tăng
- Hiệu năng giảm

Go Runtime muốn các Goroutine liên quan tiếp tục chạy trên cùng một CPU nếu có thể.

---

# Ý tưởng của Go Team

Sau nhiều thử nghiệm, nhóm phát triển Go nhận ra một điều quan trọng.

**Thread không phải nơi phù hợp để lưu trạng thái của Scheduler.**

Thread chỉ nên làm đúng nhiệm vụ của nó:

> Thực thi mã máy (Machine Instructions).

Mọi trạng thái của Scheduler nên được tách riêng.

Từ đó, một khái niệm hoàn toàn mới ra đời:

> **Processor (`P`)**

---

# Processor là gì?

Processor (`P`) không phải là CPU.

Processor cũng không phải là Thread.

Processor là một **Execution Context** do Go Runtime tạo ra để chứa toàn bộ tài nguyên cần thiết cho việc thực thi Goroutine.

```text
           Goroutine
               │
               ▼
        +--------------+
        | Processor(P) |
        +--------------+
               │
               ▼
         Machine (M)
               │
               ▼
          OS Thread
               │
               ▼
              CPU
```

Có thể hiểu đơn giản:

- `M` chịu trách nhiệm **chạy code**.
- `P` chịu trách nhiệm **quản lý tài nguyên để chạy code**.

---

# Những gì được chuyển sang Processor

Thay vì lưu trên Thread, Runtime chuyển rất nhiều tài nguyên sang `P`.

Ví dụ:

```text
Processor

├── Local Run Queue
├── mcache
├── Timer Heap
├── GC Assist State
├── Sudog Cache
├── Defer Pool
└── Scheduler State
```

Nhờ đó:

- Thread có thể bị thay thế.
- Processor vẫn tồn tại.
- Toàn bộ tài nguyên không bị mất.

---

# Processor giúp Runtime linh hoạt hơn

Giả sử Thread bị block bởi System Call.

```text
M1

↓

read()

↓

Blocked
```

Runtime sẽ thực hiện:

```text
P

↓

Detach khỏi M1

↓

Attach sang M2

↓

Tiếp tục chạy Goroutine
```

Điều này là không thể nếu toàn bộ trạng thái nằm trên Thread.

---

# Một Processor gần tương ứng với một CPU Logic

Go Runtime mặc định tạo số lượng `P` bằng:

```go
runtime.GOMAXPROCS()
```

Ví dụ:

Máy có:

```text
8 Logical CPU
```

Runtime sẽ tạo:

```text
P0

P1

P2

P3

P4

P5

P6

P7
```

Mỗi Processor sở hữu:

- Local Queue
- Memory Cache
- Timer
- GC State

Điều này giúp giảm đáng kể việc chia sẻ tài nguyên giữa các CPU.

---

## Key Takeaways

Sau khi bổ sung `Processor (P)`:

- Scheduler không còn phụ thuộc vào Thread.
- Giảm tranh chấp Global Lock.
- Tăng Cache Locality.
- Hỗ trợ Work Stealing hiệu quả.
- Cho phép Thread bị block mà Scheduler vẫn hoạt động bình thường.
- Tăng khả năng mở rộng trên hệ thống nhiều CPU.

Có thể nói:

> `P` chính là trái tim của Go Scheduler.

---

# 13.2 Historical Evolution (Before Go 1.1 → GMP)

Để hiểu rõ giá trị của `Processor`, chúng ta cần nhìn lại lịch sử phát triển của Go Scheduler.

---

# Go 1.0 — Mô hình GM

Phiên bản Go đầu tiên sử dụng mô hình khá đơn giản.

```text
Goroutine
     │
     ▼
Machine
```

Không tồn tại `Processor`.

Mỗi Machine chịu trách nhiệm:

- lấy Goroutine
- chạy Goroutine
- quản lý Scheduler

```text
G1 ─────────► M1

G2 ─────────► M2

G3 ─────────► M3
```

Lúc đó Scheduler chủ yếu dựa trên một Global Run Queue.

```text
Global Queue

↓

Machine lấy Goroutine

↓

Thực thi
```

Thiết kế này đơn giản nhưng nhanh chóng bộc lộ hạn chế khi Go được sử dụng trên các máy chủ nhiều lõi.

---

# Những hạn chế của mô hình GM

## Lock Contention

Mọi Thread đều phải truy cập cùng một Global Queue.

```text
          Global Queue

        +------------+

        ▲ ▲ ▲ ▲ ▲

        │ │ │ │ │

       M1 M2 M3 M4
```

Khi số lượng Thread tăng:

- Lock tăng.
- Wait tăng.
- Throughput giảm.

---

## Không tận dụng CPU Cache

Một Goroutine vừa chạy xong.

Ngay sau đó Scheduler lại chuyển Goroutine tiếp theo sang CPU khác.

```text
CPU 1

↓

CPU 7
```

Cache CPU gần như phải xây dựng lại từ đầu.

---

## Thread Block

Một Thread gọi:

```go
syscall.Read()
```

OS sẽ block Thread.

Nếu Queue gắn với Thread:

```text
Queue

↓

Thread

↓

Blocked
```

Toàn bộ công việc trên Thread bị ảnh hưởng.

---

## Memory Cache không ổn định

Memory Cache gắn với Thread.

Khi Thread bị hủy:

```text
Thread

↓

Destroy

↓

Cache mất
```

Runtime phải xây dựng lại Cache mới.

Đây là nguyên nhân làm tăng Fragmentation và giảm hiệu quả cấp phát bộ nhớ.

---

# Sự ra đời của GMP

Sau nhiều nghiên cứu và benchmark, nhóm phát triển Go quyết định tách Scheduler thành ba thành phần.

```text
          G

          │

          ▼

          P

          │

          ▼

          M
```

Đây là kiến trúc vẫn được sử dụng cho đến ngày nay.

---

# Vai trò của từng thành phần

```text
G

↓

Đơn vị thực thi
```

```text
P

↓

Quản lý tài nguyên
```

```text
M

↓

Thực thi Goroutine trên OS Thread
```

Ba thành phần này phối hợp với nhau tạo thành Scheduler hiện đại của Go.

---

# Những cải tiến mà GMP mang lại

Sau khi giới thiệu `Processor`, Go Runtime đạt được rất nhiều cải tiến.

- Local Run Queue cho từng Processor.
- Work Stealing giữa các Processor.
- Memory Cache theo Processor.
- Timer theo Processor.
- GC Assist theo Processor.
- Thread có thể Block mà Scheduler vẫn tiếp tục hoạt động.
- Khả năng mở rộng gần tuyến tính trên hệ thống nhiều CPU.

---

# Part II — Kiến trúc của Processor (`struct p`)

Ở phần trước, chúng ta đã hiểu **vì sao Go Runtime cần `Processor (P)`** và cách nó giải quyết những hạn chế của mô hình GM Scheduler.

Trong phần này, chúng ta sẽ đi sâu vào kiến trúc bên trong của `struct p` để trả lời các câu hỏi:

- `Processor` thực chất là gì?
- Bên trong `P` chứa những thành phần nào?
- `P` được tạo và hủy khi nào?
- Một `P` có những trạng thái (State) nào trong suốt vòng đời của nó?

Đây là nền tảng để hiểu các chương tiếp theo như **Scheduler**, **Memory Allocator**, **Garbage Collector**, **Netpoll** và **Timer**.

---

# 13.3 Anatomy of `struct p`

## `struct p` nằm ở đâu?

Trong Go Runtime, định nghĩa của `Processor` nằm trong:

```text
src/runtime/runtime2.go
```

Đây là một trong những cấu trúc dữ liệu quan trọng nhất của toàn bộ Runtime.

Nếu mở file `runtime2.go`, bạn sẽ thấy:

```go
type p struct {
    ...
}
```

Trong thực tế, `struct p` có hơn 100 trường dữ liệu. Tuy nhiên, phần lớn trong số đó phục vụ mục đích tối ưu hóa nội bộ.

Đối với việc hiểu Scheduler, chúng ta có thể nhóm các trường dữ liệu thành các nhóm chức năng thay vì học từng field riêng lẻ.

---

# Kiến trúc tổng thể của Processor

Có thể hình dung `Processor` như một "trung tâm điều phối" (Execution Context) của Runtime.

```text
                 +----------------------------+
                 |         Processor (P)      |
                 +----------------------------+
                 | Local Run Queue            |
                 | runnext                    |
                 | mcache                     |
                 | Timer Heap                 |
                 | GC Assist                  |
                 | Defer Pool                 |
                 | Sudog Cache                |
                 | Trace / Profiler           |
                 | Scheduler State            |
                 +----------------------------+
```

Mỗi `P` sở hữu gần như toàn bộ tài nguyên cần thiết để thực thi Goroutine.

Machine (`M`) chỉ cần "gắn" vào một `P` để bắt đầu làm việc.

---

# Nhóm 1 — Scheduler

Đây là phần quan trọng nhất.

```text
Processor

├── Local Run Queue
├── runnext
└── Scheduler State
```

Nhiệm vụ:

- quản lý Goroutine sẵn sàng chạy
- ưu tiên Goroutine mới tạo
- hỗ trợ Work Stealing

Trong chương sau chúng ta sẽ nghiên cứu chi tiết từng thành phần.

---

## Local Run Queue

Đây là hàng đợi Goroutine riêng của từng Processor.

```text
           P0

+----------------------+

G1

G2

G3

G4

+----------------------+
```

Thay vì tất cả Processor cùng truy cập một Global Queue, mỗi `P` đều có Queue riêng.

Điều này giúp:

- giảm Lock Contention
- tăng Cache Locality
- tăng Throughput

---

## `runnext`

Đây là một tối ưu rất nổi tiếng của Go Runtime.

Thay vì đưa Goroutine mới vào Queue.

Runtime có thể đặt trực tiếp vào:

```text
P

↓

runnext
```

Khi Goroutine hiện tại kết thúc:

```text
Current G

↓

runnext

↓

Execute ngay
```

Không cần:

- Push Queue
- Pop Queue

Giảm đáng kể chi phí Scheduler.

---

# Nhóm 2 — Memory Allocator

Processor còn sở hữu bộ cấp phát bộ nhớ cục bộ.

```text
Processor

↓

mcache
```

Điều này cho phép:

```go
new(T)
```

được cấp phát mà không cần Lock toàn bộ Heap.

Luồng cấp phát:

```text
new()

↓

mcache

↓

mcentral

↓

mheap
```

Đây là một trong những lý do Go có khả năng cấp phát bộ nhớ rất nhanh.

---

# Nhóm 3 — Garbage Collector

Garbage Collector cũng phụ thuộc rất nhiều vào Processor.

Mỗi `P` lưu trữ:

```text
GC Assist

GC Work Buffer

Mark Worker State
```

Nhờ vậy:

- mỗi Processor có thể tham gia Marking
- giảm tranh chấp khi GC chạy song song

---

# Nhóm 4 — Timer

Go Runtime không còn sử dụng một Timer Queue toàn cục.

Thay vào đó:

```text
P0

Timer Heap

-------------------

P1

Timer Heap

-------------------

P2

Timer Heap
```

Mỗi Processor quản lý Timer của riêng mình.

Điều này giúp:

- giảm Lock
- giảm Contention
- Wakeup nhanh hơn

Chúng ta sẽ tìm hiểu chi tiết trong Chapter 18.

---

# Nhóm 5 — Sudog Cache

Khi Goroutine Block trên:

- Channel
- Mutex
- Cond

Runtime cần tạo một đối tượng `sudog`.

Thay vì cấp phát liên tục.

Processor duy trì:

```text
Sudog Cache
```

để tái sử dụng.

Điều này giảm:

- Allocation
- GC Pressure

---

# Nhóm 6 — Defer Pool

Lệnh:

```go
defer f()
```

không phải lúc nào cũng cấp phát mới.

Processor có:

```text
Defer Pool
```

để tái sử dụng các đối tượng defer.

---

# Nhóm 7 — Trace & Profiling

Runtime còn lưu trên `P` nhiều thông tin phục vụ:

- Tracing
- Profiling
- Scheduler Statistics

Ví dụ:

```text
P

↓

Running Time

↓

GC Time

↓

Scheduling Count
```

Các công cụ như:

- `go tool trace`
- `pprof`

đều sử dụng những dữ liệu này.

---

# Tổng quan kiến trúc của Processor

```text
                    Processor (P)

        +----------------------------------+

        Scheduler
        ├── Local Run Queue
        ├── runnext
        └── Scheduler State

        Memory
        ├── mcache
        └── Tiny Allocator

        Garbage Collector
        ├── GC Assist
        ├── Work Buffer
        └── Mark State

        Timer
        └── Timer Heap

        Synchronization
        ├── Sudog Cache
        └── Defer Pool

        Debug
        └── Trace / Profiling

        +----------------------------------+
```

Có thể thấy `Processor` không chỉ phục vụ Scheduler mà còn là trung tâm của nhiều subsystem khác trong Runtime.

---

# 13.4 Life Cycle of `P`

Giống như `M`, `Processor` cũng có vòng đời riêng.

Tuy nhiên, vòng đời của `P` ổn định hơn rất nhiều.

---

# Khởi tạo

Ngay khi Runtime khởi động.

Go sẽ tạo số lượng `P` bằng:

```go
runtime.GOMAXPROCS()
```

Ví dụ:

```go
runtime.GOMAXPROCS(4)
```

Runtime sẽ tạo:

```text
P0

P1

P2

P3
```

Các `P` này tồn tại gần như suốt vòng đời của chương trình.

---

# Gắn với Machine

Một `P` không tự chạy được.

Nó phải gắn với một `M`.

```text
P0

↓

Attach

↓

M1
```

Khi đó:

```text
M1

↓

Execute Goroutine
```

---

# Tách khỏi Machine

Nếu Thread bị block.

Ví dụ:

```go
syscall.Read()
```

Runtime sẽ thực hiện:

```text
M1

↓

Blocked

↓

Detach P0

↓

P0

↓

Attach

↓

M2
```

Processor không bị block theo Thread.

Đây là một trong những ưu điểm lớn nhất của GMP.

---

# Processor tiếp tục hoạt động

Sau khi chuyển sang Machine khác:

```text
P0

↓

Local Queue

↓

Tiếp tục chạy
```

Không có Goroutine nào bị ảnh hưởng.

---

# Dừng hoạt động

Khi chương trình kết thúc.

Runtime sẽ giải phóng toàn bộ:

```text
P

↓

Memory

↓

Destroy
```

Thông thường `P` không được tạo hoặc hủy trong quá trình chương trình đang chạy.

Chúng được tạo khi Runtime khởi động và tồn tại đến khi Runtime kết thúc.

---

# Thay đổi số lượng Processor

Khi gọi:

```go
runtime.GOMAXPROCS(n)
```

Runtime có thể:

- tạo thêm `P`
- vô hiệu hóa bớt `P`
- điều chỉnh Scheduler

Điều này cho phép chương trình thay đổi mức độ song song ngay trong lúc chạy.

---

# Life Cycle Diagram

```text
          Runtime Start

                 │

                 ▼

         Create P0...Pn

                 │

                 ▼

        Attach vào Machine

                 │

                 ▼

        Execute Goroutines

                 │

        ┌────────┴────────┐

        ▼                 ▼

System Call         Continue Running

        │

        ▼

Detach khỏi M

        │

        ▼

Attach sang M khác

        │

        ▼

Continue Scheduling

        │

        ▼

Runtime Exit

        │

        ▼

Destroy P
```

---

# 13.5 P States

Mặc dù `Processor` ổn định hơn `Machine`, nhưng nó vẫn có nhiều trạng thái khác nhau.

Runtime sử dụng các trạng thái này để điều phối Scheduler.

---

# PIdle

Processor chưa gắn với Machine.

```text
P

↓

Idle
```

Trạng thái này xuất hiện khi:

- chương trình ít Goroutine
- vừa tạo thêm Processor
- vừa tách khỏi Machine

---

# PRunning

Đây là trạng thái phổ biến nhất.

```text
P

↓

Attached

↓

Machine

↓

Running Goroutine
```

Processor đang phục vụ việc thực thi.

---

# PSyscall

Machine đang thực hiện System Call.

Runtime sẽ cố gắng nhanh chóng tách `P` khỏi `M`.

```text
M

↓

Syscall

↓

Detach P
```

Nhờ đó các Goroutine khác không bị chặn.

---

# PGC

Trong quá trình Garbage Collection.

Processor tham gia:

- Marking
- Scanning
- Assisting

```text
P

↓

GC Worker
```

Trong pha này, một phần tài nguyên của `P` được sử dụng cho GC thay vì chạy Goroutine thông thường.

---

# PDead

Khi giảm:

```go
runtime.GOMAXPROCS()
```

Một số `P` có thể bị chuyển sang trạng thái không còn hoạt động.

```text
P

↓

Dead
```

Các `P` này không còn được Scheduler sử dụng cho đến khi cần kích hoạt lại.

---

# Sơ đồ chuyển trạng thái

```text
                 +-----------+
                 |   PIdle   |
                 +-----------+
                       |
                       | Attach vào M
                       v
                 +-----------+
                 | PRunning  |
                 +-----------+
                  |    |    |
                  |    |    |
      Syscall ----+    |    +---- GC Phase
                  |    |             
                  v    |             
             +---------+             
             | PSyscall|             
             +---------+             
                  |                  
                  | Detach / Attach  
                  v                  
             +-----------+           
             | PRunning  |           
             +-----------+
                  |
                  | Runtime Exit /
                  | GOMAXPROCS giảm
                  v
             +-----------+
             |   PDead   |
             +-----------+
```

---

## Key Takeaways

- `Processor` là **Execution Context** của Go Runtime.
- `P` chứa hầu hết tài nguyên phục vụ việc thực thi Goroutine, không chỉ Scheduler.
- Mỗi `P` sở hữu Local Run Queue, `mcache`, Timer Heap và nhiều cache nội bộ khác.
- `P` được tạo khi Runtime khởi động và thường tồn tại suốt vòng đời chương trình.
- `P` có thể tách khỏi một `Machine` và gắn sang `Machine` khác mà không làm gián đoạn Scheduler.
- Hiểu rõ kiến trúc và vòng đời của `Processor` là nền tảng để nghiên cứu **GMP Scheduler**, **Memory Allocator**, **Garbage Collector** và **Netpoll** trong các chương tiếp theo.

# Part III — Local Scheduling

Ở các phần trước, chúng ta đã biết rằng mỗi **Processor (`P`)** đều sở hữu một **Local Run Queue** riêng.

Đây là một trong những cải tiến quan trọng nhất của GMP Scheduler.

Nếu không có Local Run Queue, tất cả Goroutine sẽ phải đi qua một Global Queue duy nhất, khiến Scheduler nhanh chóng trở thành điểm nghẽn khi số lượng CPU tăng lên.

Trong phần này, chúng ta sẽ tìm hiểu cơ chế lập lịch cục bộ (Local Scheduling) của Go Runtime:

- Local Run Queue hoạt động như thế nào?
- Tại sao `runnext` lại giúp Scheduler nhanh hơn?
- Cơ chế Work Stealing cân bằng tải giữa các Processor ra sao?
- Khi nào Runtime mới truy cập Global Queue?

---

# 13.6 Local Run Queue

## Từ Global Queue đến Local Queue

Giả sử hệ thống chỉ có một Global Queue.

```text
                 Global Run Queue

+----------------------------------------------+
| G1 G2 G3 G4 G5 G6 G7 G8 G9 G10 ...           |
+----------------------------------------------+

          ▲        ▲         ▲
          │        │         │
         P0       P1        P2
```

Mỗi Processor muốn lấy Goroutine đều phải:

```go
lock(globalQueue)

g := pop(globalQueue)

unlock(globalQueue)
```

Nếu máy có:

- 16 Logical CPU
- 16 Processor

thì sẽ có 16 Processor cùng tranh chấp một mutex.

Scheduler sẽ không thể mở rộng theo số lượng CPU.

---

## Ý tưởng của Local Run Queue

Go Runtime giải quyết bài toán này bằng cách cấp cho **mỗi Processor một hàng đợi riêng**.

```text
          +----------------+
          |      P0        |
          +----------------+
          | G1             |
          | G2             |
          | G3             |
          +----------------+

          +----------------+
          |      P1        |
          +----------------+
          | G4             |
          | G5             |
          +----------------+

          +----------------+
          |      P2        |
          +----------------+
          | G6             |
          | G7             |
          +----------------+
```

Bây giờ:

- P0 chỉ thao tác trên Queue của P0.
- P1 chỉ thao tác trên Queue của P1.
- P2 chỉ thao tác trên Queue của P2.

Không còn phải tranh chấp Global Lock cho hầu hết các lần Scheduling.

---

## Local Queue được lưu ở đâu?

Bên trong `struct p`.

Đơn giản hóa:

```go
type p struct {
    runq     [256]guintptr
    runqhead uint32
    runqtail uint32
}
```

Trong Runtime thực tế:

- `runq` là Circular Buffer.
- `runqhead`
- `runqtail`

được sử dụng để quản lý Queue.

```text
            Local Queue

+--------------------------------+

head                     tail

 ↓                        ↓

 G1 G2 G3 G4 G5 ...

+--------------------------------+
```

---

## Tại sao lại là Circular Queue?

Thay vì:

```text
Push

↓

Dịch toàn bộ mảng
```

Runtime sử dụng Circular Buffer.

```text
        0   1   2   3   4

      +---+---+---+---+---+

Head->| G | G | G |   |   |<-Tail

      +---+---+---+---+---+
```

Sau nhiều lần Push/Pop.

```text
      +---+---+---+---+---+

      | G |   |   | G | G |

      +---+---+---+---+---+
```

Không cần copy dữ liệu.

Độ phức tạp:

- Push: **O(1)**
- Pop: **O(1)**

---

## Local Queue ưu tiên điều gì?

Scheduler luôn ưu tiên:

```text
Local Queue

↓

Global Queue

↓

Work Stealing

↓

Netpoll

↓

Idle
```

Điều này giúp giảm tối đa việc đồng bộ giữa các Processor.

---

# Tại sao Local Queue lại nhanh?

Có ba nguyên nhân chính.

## 1. Không cần Global Lock

Mỗi Processor thao tác trên Queue của chính nó.

```text
P0

↓

Queue0

-----------------

P1

↓

Queue1
```

Không có Processor nào cùng ghi vào một Queue.

---

## 2. Cache Locality

Processor thường tiếp tục chạy các Goroutine mà chính nó vừa tạo.

```go
func worker() {

    go taskA()

    go taskB()

    go taskC()

}
```

Các Goroutine này:

- dùng cùng dữ liệu
- dùng cùng Cache

Nếu tiếp tục chạy trên cùng Processor:

```text
CPU Cache

↓

Reuse
```

Hiệu năng tăng đáng kể.

---

## 3. Ít Synchronization hơn

Hầu hết Scheduling diễn ra hoàn toàn cục bộ.

Không cần:

- Mutex
- Atomic nhiều lần
- Wakeup Processor khác

---

# 13.7 `runnext` Optimization

## Vấn đề

Giả sử Goroutine A tạo Goroutine B.

```go
go B()
```

Nếu Runtime luôn Push vào Queue.

```text
Queue

↓

Push B

↓

Current G kết thúc

↓

Pop B
```

Scheduler phải:

- Push
- Pop

mặc dù gần như chắc chắn B sẽ chạy ngay sau A.

---

## Ý tưởng của `runnext`

Runtime tạo thêm một vị trí đặc biệt.

```text
Processor

+---------------------------+

runnext

Local Queue

+---------------------------+
```

Thay vì Push.

Runtime đặt trực tiếp:

```text
Current G

↓

Create B

↓

runnext = B
```

---

## Khi Goroutine hiện tại kết thúc

Scheduler kiểm tra:

```text
runnext ?

↓

Có

↓

Execute luôn
```

Không cần:

- Push Queue
- Pop Queue

---

## Minh họa

Không có `runnext`

```text
A

↓

Push B

↓

Queue

↓

Pop B

↓

Run B
```

Có `runnext`

```text
A

↓

runnext = B

↓

Run B
```

Ngắn hơn rất nhiều.

---

## Lợi ích

Giảm:

- Queue Operation
- Lock
- Atomic
- Cache Miss

Đây là một trong những tối ưu nhỏ nhưng cực kỳ hiệu quả của Go Runtime.

---

## Khi nào `runnext` được sử dụng?

Thường gặp khi:

- tạo Goroutine mới
- Goroutine vừa được Wakeup
- Runtime muốn ưu tiên Goroutine vừa tương tác

Ví dụ:

```go
go worker()
```

Rất nhiều trường hợp `worker()` sẽ được đặt vào `runnext`.

---

# 13.8 Work Stealing

## Tại sao cần Work Stealing?

Giả sử:

```text
P0

100 Goroutine

-------------------

P1

0 Goroutine
```

Nếu không có Work Stealing:

```text
P0

Busy

--------------------

P1

Idle
```

Một CPU làm việc.

Một CPU đứng chờ.

Lãng phí tài nguyên.

---

# Ý tưởng

Processor rảnh sẽ "đánh cắp" Goroutine từ Processor bận.

```text
P1

Idle

↓

Steal

↓

P0
```

---

## Minh họa

Trước khi Steal

```text
P0

G1
G2
G3
G4
G5
G6

--------------------

P1

(empty)
```

Sau khi Steal

```text
P0

G1
G2
G3

--------------------

P1

G4
G5
G6
```

Hai Processor cùng làm việc.

---

## Tại sao chỉ lấy một nửa?

Runtime thường lấy:

> Khoảng một nửa Local Queue.

Ví dụ:

```text
P0

10 Goroutine
```

P1 sẽ lấy:

```text
5 Goroutine
```

Không lấy toàn bộ.

Nếu lấy hết.

```text
P0

↓

Empty
```

P0 lại phải đi Steal ngược.

Chi phí tăng.

---

## Chọn Processor nào để Steal?

Runtime chọn ngẫu nhiên.

```text
P3

↓

Random

↓

P7
```

Điều này giúp:

- tránh nhiều Processor cùng Steal một nơi
- cân bằng tải tốt hơn

---

## Khi nào mới Steal?

Chỉ khi:

- Local Queue rỗng.
- `runnext` rỗng.
- Global Queue rỗng.

Lúc đó Runtime mới bắt đầu tìm Processor khác.

Điều này giúp Work Stealing không trở thành chi phí thường xuyên.

---

# 13.9 Global Queue Interaction

## Global Queue còn tồn tại không?

Có.

Nhưng Global Queue không còn là trung tâm của Scheduler.

Nó chỉ đóng vai trò:

- nơi nhận Goroutine khi Local Queue đầy
- nơi chia sẻ Goroutine giữa các Processor
- nơi tiếp nhận một số Goroutine đặc biệt

---

## Khi Local Queue đầy

Giả sử Queue của P0 đã đầy.

```text
P0

256 Goroutine
```

Runtime sẽ chuyển một phần Goroutine sang Global Queue.

```text
P0

↓

Global Queue
```

Điều này giúp tránh việc Local Queue tăng vô hạn.

---

## Khi Processor khởi động

Một Processor mới có thể lấy Goroutine từ Global Queue.

```text
Global Queue

↓

P2
```

Điều này giúp Scheduler phân phối công việc ban đầu.

---

## Chu kỳ tìm Goroutine

Thứ tự Scheduler tìm Goroutine gần như luôn là:

```text
           +----------------+
           |  runnext ?     |
           +----------------+
                    |
                   Yes
                    |
                    v
              Execute ngay
                    |
                   No
                    |
                    v
           +----------------+
           | Local Queue ?  |
           +----------------+
                    |
                   Yes
                    |
                    v
              Execute
                    |
                   No
                    |
                    v
           +----------------+
           | Global Queue ? |
           +----------------+
                    |
                   Yes
                    |
                    v
              Execute
                    |
                   No
                    |
                    v
           +----------------+
           | Work Stealing  |
           +----------------+
                    |
                   Yes
                    |
                    v
              Execute
                    |
                   No
                    |
                    v
           +----------------+
           |   Netpoll ?    |
           +----------------+
                    |
                   Yes
                    |
                    v
              Execute
                    |
                   No
                    |
                    v
                 Processor Idle
```

Đây là chiến lược giúp Runtime:

- ưu tiên dữ liệu cục bộ,
- giảm đồng bộ,
- tận dụng CPU Cache,
- chỉ truy cập tài nguyên dùng chung khi thật sự cần thiết.

---

# Tổng kết Part III

Có thể mô tả Local Scheduling của Go Runtime bằng sơ đồ sau:

```text
                    +----------------+
                    |   Processor P  |
                    +----------------+
                            |
                     +------+------+
                     |             |
                     v             |
                  runnext          |
                     |             |
                     v             |
               Local Run Queue     |
                     |             |
                     +-------------+
                           |
                           v
                    Execute Goroutine
                           |
                  Local Queue rỗng?
                           |
                          Yes
                           |
                           v
                     Global Queue
                           |
                    Vẫn không có?
                           |
                           v
                     Work Stealing
                           |
                    Vẫn không có?
                           |
                           v
                         Netpoll
                           |
                    Không có việc?
                           |
                           v
                           Idle
```

## Key Takeaways

- Mỗi `Processor` sở hữu một **Local Run Queue** riêng để giảm tranh chấp khóa và tăng khả năng mở rộng.
- `runnext` là một tối ưu quan trọng giúp Goroutine kế tiếp được thực thi ngay mà không cần đi qua Queue.
- **Work Stealing** cho phép các Processor rảnh lấy bớt công việc từ Processor bận, giúp cân bằng tải giữa các CPU.
- **Global Queue** vẫn tồn tại nhưng chỉ đóng vai trò hỗ trợ, không còn là trung tâm của Scheduler như trong Go 1.0.
- Chiến lược ưu tiên của Scheduler là: **`runnext` → Local Queue → Global Queue → Work Stealing → Netpoll → Idle**.
- Kiến trúc Local Scheduling là nền tảng giúp Go Runtime đạt được hiệu năng cao trên các hệ thống đa lõi với hàng triệu Goroutine.

# Part IV — Memory Ownership

Ở Chapter 12 chúng ta đã biết rằng `Machine (M)` đại diện cho một OS Thread.

Ở Part II của chương này, chúng ta cũng thấy rằng rất nhiều tài nguyên được đặt trên `Processor (P)` thay vì `Machine`.

Một trong những quyết định quan trọng nhất của Go Runtime là:

> **Memory Allocator không thuộc về `M` mà thuộc về `P`.**

Điều này có ảnh hưởng rất lớn đến:

- hiệu năng cấp phát bộ nhớ
- khả năng mở rộng trên nhiều CPU
- Garbage Collector
- Scheduler

Trong phần này chúng ta sẽ tìm hiểu:

- tại sao `mcache` lại nằm trên `P`
- luồng cấp phát bộ nhớ nhanh (Fast Path)
- Tiny Allocator hoạt động như thế nào
- tại sao Goroutine phải "trả nợ" cho Garbage Collector (GC Assist)

---

# 13.10 `mcache`

## Memory Allocator của Go

Go Runtime không cấp phát bộ nhớ trực tiếp từ OS mỗi khi gọi:

```go
new(T)

make([]int, 100)

&User{}
```

Nếu làm như vậy:

```text
Allocate

↓

OS

↓

Allocate

↓

OS

↓

Allocate

↓

OS
```

Mỗi lần cấp phát đều phải gọi vào Kernel.

Điều này cực kỳ chậm.

Thay vào đó Go Runtime xây dựng một Memory Allocator nhiều tầng.

```text
           mcache

              │

              ▼

           mcentral

              │

              ▼

            mheap

              │

              ▼

         Operating System
```

Đây là kiến trúc sẽ được nghiên cứu chi tiết ở Chapter 19.

Trong chương này chúng ta chỉ tập trung vào vai trò của `mcache`.

---

# `mcache` là gì?

`mcache` là **bộ nhớ đệm (Memory Cache)** dành riêng cho từng Processor.

```text
            Processor P0

        +--------------------+

        Local Queue

        mcache

        Timer Heap

        GC State

        +--------------------+
```

Mỗi Processor có đúng **một** `mcache`.

Điều này có nghĩa:

```text
P0

↓

mcache0

--------------------

P1

↓

mcache1

--------------------

P2

↓

mcache2
```

Không Processor nào chia sẻ `mcache` với Processor khác.

---

# Tại sao không đặt `mcache` trên Machine?

Giả sử:

```text
Machine

↓

mcache
```

Khi Thread bị Block.

```text
Machine

↓

System Call

↓

Sleep
```

Memory Cache cũng "đi ngủ" theo.

Hoặc tệ hơn.

Thread bị Destroy.

```text
Machine

↓

Destroy

↓

mcache mất
```

Runtime phải tạo Cache mới.

Điều này làm tăng:

- Fragmentation
- Lock
- Memory Waste

Trong khi `Processor` gần như tồn tại suốt vòng đời chương trình.

Do đó đặt `mcache` trên `P` sẽ ổn định hơn rất nhiều.

---

# `mcache` chứa gì?

Có thể hình dung đơn giản.

```text
mcache

├── Tiny Allocator

├── Span Class 8B

├── Span Class 16B

├── Span Class 32B

├── Span Class 64B

├── ...

└── Span Class lớn hơn
```

Mỗi Size Class đều có một Span riêng.

Ví dụ:

```text
8 bytes

↓

Span A

-------------------

16 bytes

↓

Span B

-------------------

32 bytes

↓

Span C
```

Khi cần cấp phát.

Runtime chỉ cần lấy Object trong Span tương ứng.

Không cần Lock toàn bộ Heap.

---

# Vai trò của `mcache`

`mcache` giúp Runtime đạt được:

- Lock-free Allocation
- Cache Locality
- Allocation cực nhanh
- Giảm áp lực lên Garbage Collector

Đây là tầng được sử dụng nhiều nhất trong toàn bộ Memory Allocator.

---

# 13.11 Allocation Fast Path

## Fast Path là gì?

Fast Path là đường đi nhanh nhất khi cấp phát bộ nhớ.

Mục tiêu của Runtime:

> Cấp phát mà không cần Lock.

Ví dụ:

```go
u := &User{}
```

Nếu `mcache` còn Object.

Luồng cấp phát sẽ là:

```text
new()

↓

mcache

↓

Return Pointer
```

Không cần:

- Mutex
- Scheduler
- Garbage Collector
- Operating System

Đây chính là Fast Path.

---

# Minh họa

```text
          Allocate

              │

              ▼

        Check mcache

              │

      +-------+-------+

      |               |

      ▼               ▼

   Có Object      Hết Object

      │               │

      ▼               ▼

 Return Pointer    Slow Path
```

Hơn 90% Allocation của Go diễn ra trên Fast Path.

Đó là lý do Go có tốc độ cấp phát rất cao.

---

# Tại sao Fast Path nhanh?

Giả sử Object cần:

```text
32 bytes
```

Runtime sẽ:

```text
mcache

↓

Span 32B

↓

Free Object

↓

Return
```

Không cần tìm kiếm.

Không cần Lock.

Không cần gọi OS.

---

# Slow Path

Nếu Span đã hết Object.

```text
mcache

↓

Empty
```

Runtime sẽ chuyển sang:

```text
mcentral
```

```text
mcache

↓

mcentral

↓

Refill Span

↓

mcache
```

Sau đó tiếp tục Allocation.

Nếu `mcentral` cũng hết.

```text
mcentral

↓

mheap

↓

OS
```

Đây là Slow Path.

May mắn là nó xảy ra ít hơn rất nhiều so với Fast Path.

---

# Luồng Allocation hoàn chỉnh

```text
             Allocate

                 │

                 ▼

              mcache

         +------+------+

         |             |

         ▼             ▼

    Fast Path      Slow Path

                        │

                        ▼

                    mcentral

                        │

                        ▼

                     mheap

                        │

                        ▼

                        OS
```

---

# 13.12 Tiny Allocator

## Vấn đề

Giả sử chương trình liên tục tạo các Object rất nhỏ.

```go
type A struct {
    x byte
}
```

Kích thước chỉ:

```text
1 byte
```

Nếu mỗi Object đều chiếm một Allocation riêng.

```text
Heap

A

A

A

A

A
```

Chi phí quản lý còn lớn hơn dữ liệu thực tế.

---

# Tiny Allocator

Go Runtime tạo một vùng nhớ đặc biệt.

```text
Tiny Block

16 Bytes
```

Các Object nhỏ sẽ được đặt chung vào Block này.

Ví dụ:

```text
+----------------+

A

B

C

D

E

+----------------+
```

Thay vì:

```text
Allocate

A

Allocate

B

Allocate

C
```

Runtime chỉ Allocate một lần.

---

# Điều kiện sử dụng Tiny Allocator

Tiny Allocator chỉ dành cho:

- Object rất nhỏ
- Không chứa Pointer

Ví dụ:

```go
struct {
    a byte
}
```

Hoặc:

```go
struct {
    x uint16
}
```

Nhưng:

```go
struct {
    next *Node
}
```

sẽ không sử dụng Tiny Allocator.

Lý do:

GC cần theo dõi Pointer.

---

# Lợi ích

Tiny Allocator giúp:

- Giảm số lần Allocation
- Giảm Fragmentation
- Giảm GC
- Tăng Throughput

Đặc biệt hiệu quả với:

- JSON
- Logging
- String Builder
- Encoding

---

# Minh họa

Không có Tiny Allocator.

```text
Allocate

↓

1 byte

Allocate

↓

2 byte

Allocate

↓

4 byte
```

Có Tiny Allocator.

```text
Allocate

↓

16 bytes

↓

Reuse nhiều lần
```

---

# 13.13 Assist GC

## Tại sao cần GC Assist?

Giả sử chương trình chỉ:

```go
for {

    new(User)

}
```

Allocation sẽ diễn ra liên tục.

Trong khi Garbage Collector chưa kịp thu gom.

```text
Allocate

Allocate

Allocate

Allocate

Allocate
```

Heap sẽ tăng rất nhanh.

Nếu Runtime để mặc điều này xảy ra.

Memory sẽ bùng nổ.

---

# Ý tưởng

Người cấp phát nhiều bộ nhớ hơn phải giúp Garbage Collector nhiều hơn.

Đây gọi là:

> **GC Assist**

---

# GC Assist hoạt động như thế nào?

Giả sử Goroutine vừa Allocate:

```text
100 MB
```

Runtime ghi nhận:

```text
Allocation Debt

100 MB
```

Sau đó.

Trước khi tiếp tục Allocation.

Goroutine phải giúp GC.

```text
Allocate

↓

Debt tăng

↓

Assist GC

↓

Debt giảm

↓

Allocate tiếp
```

---

# Assist làm gì?

Trong thời gian Assist.

Goroutine tham gia:

- Mark Object
- Scan Pointer
- Quét Heap

Thực chất Goroutine trở thành một GC Worker tạm thời.

---

# Tại sao đặt trên Processor?

Thông tin Assist được lưu trên `P`.

```text
Processor

├── GC Credit

├── GC Debt

├── Work Buffer
```

Điều này giúp:

- tránh chia sẻ giữa các Thread
- giảm Lock
- mỗi Processor tự cân bằng lượng Allocation của mình

---

# Ví dụ

Giả sử:

```text
P0

Allocate rất nhiều
```

```text
P1

Ít Allocation
```

P0 sẽ Assist GC nhiều hơn.

P1 hầu như không phải Assist.

Điều này tạo ra cơ chế cân bằng rất tự nhiên.

---

# Mối quan hệ giữa Allocation và GC

Có thể mô tả như sau.

```text
          Allocate

              │

              ▼

          GC Debt

              │

              ▼

      Debt vượt ngưỡng?

         +------+------+

         |             |

        No            Yes

         |             |

         ▼             ▼

 Continue        Assist GC

                      │

                      ▼

              Debt giảm

                      │

                      ▼

                 Continue
```

GC Assist giúp Go Runtime tránh tình trạng một Goroutine liên tục cấp phát bộ nhớ trong khi các Goroutine khác phải "gánh" toàn bộ chi phí Garbage Collection.

---

# Memory Ownership của Processor

Sau khi tìm hiểu bốn thành phần trên, chúng ta có thể thấy `Processor` không chỉ quản lý Scheduling mà còn sở hữu phần lớn tài nguyên liên quan đến bộ nhớ.

```text
                  Processor (P)

        +----------------------------------+

        Scheduler
        ├── Local Run Queue
        ├── runnext

        Memory
        ├── mcache
        ├── Tiny Allocator
        ├── Span Cache

        Garbage Collector
        ├── GC Credit
        ├── GC Debt
        ├── Assist State
        ├── Work Buffer

        +----------------------------------+
```

Đây chính là lý do chương này được đặt tên là **Memory Ownership**.

Processor không chỉ quyết định **Goroutine nào sẽ chạy**, mà còn quyết định **bộ nhớ được cấp phát và Garbage Collector được phối hợp như thế nào**.

---

## Key Takeaways

- `mcache` là bộ nhớ đệm riêng của từng `Processor`, giúp hầu hết các Allocation diễn ra mà không cần Lock.
- Fast Path ưu tiên cấp phát trực tiếp từ `mcache`; chỉ khi hết tài nguyên mới chuyển sang `mcentral` và `mheap`.
- Tiny Allocator gom nhiều đối tượng nhỏ, không chứa Pointer vào cùng một vùng nhớ để giảm Allocation và áp lực lên Garbage Collector.
- GC Assist buộc Goroutine cấp phát nhiều bộ nhớ phải tham gia hỗ trợ Garbage Collector, giúp cân bằng chi phí thu gom.
- Việc đặt `mcache`, Tiny Allocator và trạng thái GC trên `Processor` thay vì `Machine` là một quyết định thiết kế quan trọng, giúp Go Runtime đạt được khả năng mở rộng và hiệu năng cao trên hệ thống đa lõi.

# Part V — Timer & Netpoll

Ở những phần trước, chúng ta đã thấy `Processor (P)` không chỉ quản lý Scheduler mà còn sở hữu:

- Local Run Queue
- Memory Cache (`mcache`)
- Garbage Collector State

Trong phần này, chúng ta sẽ khám phá hai subsystem cực kỳ quan trọng khác cũng được đặt trên `P`:

- **Timer**
- **Netpoll**

Hai thành phần này là nền tảng giúp Go Runtime thực hiện:

- `time.Sleep()`
- `time.After()`
- `time.Timer`
- `time.Ticker`
- Deadline của Network I/O
- Context Timeout
- HTTP Server
- gRPC
- Database Driver
- Async Network Programming

Nếu Scheduler quyết định **"Goroutine nào sẽ chạy"**, thì Timer và Netpoll quyết định:

> **"Khi nào Goroutine được đánh thức để tiếp tục chạy."**

---

# 13.14 Timer Heap

## Tại sao Runtime cần Timer?

Hầu như mọi chương trình Go đều sử dụng Timer.

Ví dụ:

```go
time.Sleep(time.Second)
```

```go
time.After(5 * time.Second)
```

```go
context.WithTimeout(...)
```

```go
ticker := time.NewTicker(...)
```

Mỗi API trên đều cần Runtime biết:

> "Khi nào Goroutine nên được đánh thức?"

---

## Cách đơn giản nhất

Một cách đơn giản là Runtime tạo một danh sách Timer.

```text
Timer List

T1

T2

T3

T4
```

Mỗi lần Scheduler chạy sẽ quét toàn bộ danh sách.

```text
Scheduler

↓

Check T1

↓

Check T2

↓

Check T3

↓

Check T4
```

Giả sử có:

- 1 triệu Timer

Runtime sẽ phải quét:

```text
1,000,000 Timer
```

mỗi lần Scheduling.

Điều này hoàn toàn không khả thi.

---

# Heap là lời giải

Go Runtime sử dụng:

> **Min Heap**

để lưu Timer.

```text
             T1

           /    \

         T2      T3

        /  \

      T4   T5
```

Đặc điểm của Min Heap:

```text
Root

↓

Timer hết hạn sớm nhất
```

Scheduler chỉ cần kiểm tra Root.

Không cần quét toàn bộ Heap.

---

## Tại sao lại là Min Heap?

Giả sử có các Timer:

```text
10 ms

50 ms

100 ms

200 ms
```

Min Heap sẽ luôn giữ:

```text
10 ms
```

ở Root.

```text
             10

          /      \

        50       100

       /

     200
```

Scheduler chỉ cần nhìn vào:

```text
Root
```

là biết Timer tiếp theo sẽ hết hạn khi nào.

---

# Mỗi Processor có Timer Heap riêng

Đây là điểm thay đổi rất quan trọng từ Go 1.14 trở đi.

Trước đây.

```text
Global Timer Heap
```

Hiện nay.

```text
P0

Timer Heap

------------------

P1

Timer Heap

------------------

P2

Timer Heap
```

Mỗi Processor sở hữu Timer Heap riêng.

---

## Tại sao không dùng Global Timer?

Giả sử:

```text
100 CPU
```

Nếu tất cả đều truy cập:

```text
Global Timer Heap
```

Sẽ xảy ra:

```text
Lock

↓

Wait

↓

Lock

↓

Wait
```

Scheduler lại gặp Lock Contention.

Đó là lý do Runtime chuyển Timer sang `P`.

---

# Thêm Timer

Ví dụ:

```go
time.Sleep(time.Second)
```

Runtime tạo Timer.

```text
Timer

↓

Insert

↓

P.Timer Heap
```

Nếu Heap hiện tại:

```text
100 ms

200 ms

300 ms
```

Timer mới:

```text
50 ms
```

Runtime sẽ Heapify.

```text
50

↓

100

↓

200

↓

300
```

---

# Xóa Timer

Sau khi Timer hết hạn.

```text
Pop Root
```

Heap sẽ được cân bằng lại.

```text
100

↓

200

↓

300
```

Độ phức tạp:

```
Insert

O(log n)
```

```
Delete

O(log n)
```

```
Peek

O(1)
```

---

# Luồng hoạt động của Timer

```text
        time.Sleep()

              │

              ▼

         Create Timer

              │

              ▼

      Insert vào Timer Heap

              │

              ▼

      Scheduler kiểm tra Root

              │

     Timer hết hạn?

        │         │

       No        Yes

        │         │

        ▼         ▼

     Continue   Wakeup G
```

---

# Timer không đánh thức Goroutine ngay

Đây là một điểm dễ gây hiểu nhầm.

Timer chỉ làm nhiệm vụ:

```text
Expire
```

Sau đó Runtime:

```text
Move Goroutine

↓

Runnable

↓

Local Queue
```

Scheduler mới quyết định:

```text
Khi nào Goroutine thực sự chạy.
```

Timer không trực tiếp chạy Goroutine.

---

# 13.15 Poll Cache

## Netpoll là gì?

Khi Goroutine thực hiện:

```go
conn.Read()
```

Runtime không muốn Block toàn bộ Thread.

Thay vào đó.

Runtime sử dụng:

```text
Netpoll
```

để theo dõi trạng thái của File Descriptor.

---

## Poll Descriptor

Mỗi kết nối mạng đều có:

```text
PollDesc
```

Ví dụ:

```text
Socket

↓

PollDesc
```

PollDesc lưu:

- File Descriptor
- Read Waiter
- Write Waiter
- Deadline
- State

Có thể hình dung.

```text
PollDesc

├── fd
├── read waiter
├── write waiter
├── deadline
└── status
```

---

# Poll Cache

Việc tạo PollDesc liên tục rất tốn kém.

Go Runtime tạo:

```text
Poll Cache
```

để tái sử dụng.

```text
Processor

↓

Poll Cache

↓

PollDesc
```

Khi Connection đóng.

PollDesc không bị hủy.

```text
Connection Close

↓

Return

↓

Poll Cache
```

Connection tiếp theo sẽ tái sử dụng.

---

# Luồng hoạt động

```text
New Connection

↓

Poll Cache

↓

Có PollDesc ?

     │

 Yes │ No

     │

Reuse   Allocate
```

Điều này giúp:

- giảm Allocation
- giảm Fragmentation
- giảm GC

---

# Poll Cache và Netpoll

```text
Socket

↓

PollDesc

↓

Netpoll

↓

Wakeup
```

Poll Cache giúp Netpoll hoạt động hiệu quả hơn.

---

# 13.16 Wakeup Logic

## Khi nào Goroutine được đánh thức?

Có ba nguồn chính.

```text
Timer

Netpoll

Scheduler
```

Tất cả cuối cùng đều dẫn tới:

```text
Runnable
```

---

# Wakeup từ Timer

Giả sử:

```go
time.Sleep(time.Second)
```

Luồng hoạt động.

```text
Sleep

↓

Create Timer

↓

Timer Heap

↓

Expire

↓

Runnable

↓

Local Queue
```

Sau đó Scheduler mới chọn Goroutine để chạy.

---

# Wakeup từ Network

Ví dụ:

```go
conn.Read()
```

Nếu chưa có dữ liệu.

```text
Goroutine

↓

Waiting

↓

Netpoll
```

Khi Kernel thông báo.

```text
EPOLLIN

↓

Netpoll

↓

Runnable

↓

Local Queue
```

---

# Wakeup từ Channel

Một Goroutine đang:

```go
<-ch
```

Khi có dữ liệu.

```text
Send

↓

Receiver

↓

Runnable

↓

Local Queue
```

---

# Wakeup từ Mutex

```go
mutex.Lock()
```

Nếu Mutex được Unlock.

```text
Unlock

↓

Wake Waiter

↓

Runnable
```

---

# Wakeup Flow

Có thể mô tả toàn bộ Runtime như sau.

```text
                 Block

                   │

        +----------+-----------+

        │                      │

        ▼                      ▼

     Timer                 Netpoll

        │                      │

        +----------+-----------+

                   │

                   ▼

             Runnable

                   │

                   ▼

           Local Run Queue

                   │

                   ▼

             Scheduler

                   │

                   ▼

             Running G
```

---

# Tại sao Wakeup không chạy Goroutine ngay?

Một câu hỏi rất phổ biến.

Tại sao Runtime không làm:

```text
Wakeup

↓

Run luôn
```

Lý do là:

Scheduler cần quyết định:

- CPU nào sẽ chạy.
- Processor nào còn rảnh.
- Có cần Work Stealing không.
- Có Goroutine ưu tiên hơn không.

Do đó.

Wakeup chỉ chuyển Goroutine sang trạng thái:

```text
Runnable
```

Việc chạy hay chưa hoàn toàn do Scheduler quyết định.

---

# Timer + Netpoll + Scheduler

Có thể xem ba subsystem này hoạt động như ba tầng riêng biệt.

```text
                 +----------------+

                 |   Timer Heap   |

                 +----------------+

                          │

                          ▼

                 +----------------+

                 |    Netpoll     |

                 +----------------+

                          │

                          ▼

                 +----------------+

                 | Runnable Queue |

                 +----------------+

                          │

                          ▼

                 +----------------+

                 |   Scheduler    |

                 +----------------+

                          │

                          ▼

                     Running G
```

Điều này giúp Runtime tách biệt rõ ràng:

- Theo dõi sự kiện (Timer/Network).
- Đánh thức Goroutine.
- Lập lịch thực thi.

Mỗi subsystem chỉ tập trung vào đúng một trách nhiệm.

---

# Timer và Netpoll trong `Processor`

Đến thời điểm này, chúng ta có thể nhìn lại toàn bộ những gì `Processor` đang sở hữu.

```text
                 Processor (P)

        +--------------------------------------+

        Scheduler
        ├── Local Run Queue
        ├── runnext

        Memory
        ├── mcache
        ├── Tiny Allocator

        Garbage Collector
        ├── Assist GC
        ├── Work Buffer

        Timer
        ├── Timer Heap

        Network
        ├── Poll Cache

        +--------------------------------------+
```

Có thể thấy `Processor` không chỉ là trung tâm của Scheduler mà còn là nơi tập trung hầu hết các tài nguyên cần thiết để thực thi Goroutine với độ trễ thấp và khả năng mở rộng cao.

---

## Key Takeaways

- Mỗi `Processor` sở hữu **Timer Heap** riêng, giúp loại bỏ phần lớn tranh chấp khóa khi quản lý hàng triệu Timer.
- Timer được lưu trong **Min Heap**, cho phép Runtime luôn tìm được Timer hết hạn sớm nhất với chi phí thấp.
- `Poll Cache` tái sử dụng `PollDesc`, giảm Allocation và áp lực lên Garbage Collector trong các ứng dụng mạng.
- **Netpoll** theo dõi các sự kiện I/O của Kernel và chỉ đánh thức Goroutine khi dữ liệu thực sự sẵn sàng.
- Timer, Netpoll và các primitive đồng bộ khác **không trực tiếp chạy Goroutine**; chúng chỉ chuyển Goroutine sang trạng thái **Runnable**.
- Scheduler là thành phần duy nhất quyết định **khi nào** và **trên Processor nào** Goroutine sẽ được thực thi.

# Part VI — Scheduler Cooperation

Ở các phần trước, chúng ta đã nghiên cứu từng thành phần riêng lẻ:

- `G` đại diện cho **công việc cần thực hiện**.
- `M` đại diện cho **OS Thread**.
- `P` đại diện cho **Execution Context** của Runtime.

Tuy nhiên, Go Scheduler không hoạt động dựa trên từng thành phần độc lập.

Điểm mạnh nhất của GMP Scheduler nằm ở sự **phối hợp (Cooperation)** giữa ba thành phần này.

Trong phần này chúng ta sẽ tìm hiểu:

- `P` phối hợp với `M` như thế nào.
- `P` quản lý `G` ra sao.
- Cơ chế Handoff khi Thread bị Block.
- Khái niệm Spinning Processor giúp Runtime giảm độ trễ Scheduling.

Đây là những kiến thức quan trọng để hiểu toàn bộ Scheduler trong Chapter 14.

---

# 13.17 P và M

## Vai trò của từng thành phần

Trước tiên hãy nhắc lại trách nhiệm của từng đối tượng.

### Machine (`M`)

Machine chỉ có một nhiệm vụ:

> Thực thi Goroutine trên OS Thread.

```text
Machine

↓

Execute Instructions
```

Machine **không sở hữu**:

- Local Queue
- Memory Cache
- Timer
- GC State

Machine chỉ là "động cơ" thực thi.

---

### Processor (`P`)

Processor mới là nơi chứa toàn bộ tài nguyên.

```text
Processor

├── Local Queue
├── mcache
├── Timer Heap
├── GC State
├── Scheduler State
└── runnext
```

Có thể hiểu đơn giản:

> M làm việc.

> P mang theo dụng cụ.

---

# Một Machine phải có Processor

Một Machine không thể chạy Goroutine nếu không có Processor.

```text
        +---------+
        |   M1    |
        +---------+

             ✗

Không có Processor

↓

Không thể Execute
```

Muốn chạy.

Machine phải Attach vào Processor.

```text
        +---------+

        |   M1    |

        +---------+

             │

             ▼

        +---------+

        |   P0    |

        +---------+
```

Khi đó.

```text
M1

↓

Lấy G từ P0

↓

Execute
```

---

# Quan hệ giữa M và P

Một thời điểm chỉ có:

```text
Một Processor

↓

Một Machine
```

```text
P0 ─────► M0

P1 ─────► M1

P2 ─────► M2
```

Tuy nhiên.

Một Machine có thể đổi Processor.

```text
Ban đầu

P0

↓

M1

Sau đó

P2

↓

M1
```

Và một Processor cũng có thể đổi Machine.

```text
Ban đầu

P0

↓

M1

Sau đó

P0

↓

M3
```

Điều này giúp Runtime cực kỳ linh hoạt.

---

# Tại sao phải tách M và P?

Giả sử Thread gọi:

```go
syscall.Read()
```

Machine sẽ bị Block.

```text
M1

↓

Blocked
```

Nếu Scheduler gắn với Machine.

```text
Queue

↓

Blocked
```

Toàn bộ Goroutine phía sau sẽ phải chờ.

Đó là điều Runtime muốn tránh.

---

# Sau khi tách

Runtime chỉ cần:

```text
M1

↓

Blocked

↓

Detach P0

↓

Attach sang M2
```

Processor tiếp tục hoạt động.

```text
P0

↓

M2

↓

Execute
```

Không cần chờ System Call kết thúc.

---

# Ví dụ

```text
Trước khi Block

G1

↓

P0

↓

M1

↓

CPU
```

Sau khi Block.

```text
G1

↓

System Call

↓

M1 (Blocked)
```

Runtime thực hiện.

```text
P0

↓

Detach

↓

M2

↓

G2

↓

CPU
```

CPU vẫn tiếp tục có việc để làm.

---

# 13.18 P và G

Nếu Machine là "người thực thi".

Thì Processor chính là "người quản lý Goroutine".

---

# Processor sở hữu Goroutine

Mỗi Processor có:

```text
Local Queue
```

```text
P0

↓

G1

G2

G3

G4
```

Machine chỉ lấy Goroutine từ Queue.

```text
P0

↓

Pop G

↓

M Execute
```

---

# Goroutine mới

Giả sử:

```go
go worker()
```

Runtime sẽ tạo:

```text
New G
```

Sau đó.

Đưa vào:

```text
P.Local Queue
```

Hoặc:

```text
P.runnext
```

Không đưa ngay vào Global Queue.

---

# Goroutine Block

Ví dụ.

```go
<-ch
```

Runtime chuyển Goroutine sang:

```text
Waiting
```

Processor tiếp tục lấy Goroutine khác.

```text
P

↓

Next G
```

Không có CPU nào bị lãng phí.

---

# Goroutine Wakeup

Sau khi Channel có dữ liệu.

```text
Wakeup

↓

Runnable

↓

Local Queue
```

Processor sẽ xử lý như một Goroutine bình thường.

---

# Processor quyết định thứ tự chạy

Giả sử Queue.

```text
G1

G2

G3
```

Processor sẽ quyết định:

```text
Pop

↓

Execute
```

Machine chỉ làm đúng việc:

```text
Run()
```

---

# Processor và Ownership

Một Goroutine luôn thuộc về:

```text
Processor
```

Không thuộc về:

```text
Machine
```

Đây là khác biệt rất quan trọng.

---

# 13.19 Handoff

## Handoff là gì?

Handoff là quá trình:

> Chuyển Processor từ Machine này sang Machine khác.

Đây là cơ chế quan trọng nhất giúp Scheduler không bị Block.

---

# Ví dụ

Ban đầu.

```text
P0

↓

M1

↓

Running
```

M1 gọi.

```go
read(fd)
```

OS Block Thread.

```text
M1

↓

Blocked
```

Nếu không có Handoff.

```text
P0

↓

Blocked
```

Scheduler dừng hoạt động.

---

# Handoff diễn ra

Runtime thực hiện.

```text
Detach P0

↓

Idle List
```

Sau đó.

```text
Find Idle M

↓

Attach P0
```

```text
P0

↓

M2

↓

Continue Scheduling
```

---

# Minh họa

```text
        Trước

      P0

       │

       ▼

      M1

       │

       ▼

      CPU

----------------------------

System Call

----------------------------

      P0

       │

       ▼

   Detached

       │

       ▼

      M2

       │

       ▼

      CPU
```

---

# Lợi ích

Không cần:

```text
Wait syscall()
```

CPU vẫn chạy các Goroutine khác.

Đây chính là nền tảng giúp Go xử lý hàng trăm nghìn kết nối mạng.

---

# Handoff không Copy dữ liệu

Nhiều người nghĩ.

Runtime sẽ:

```text
Copy Queue

↓

Machine khác
```

Thực tế.

Không hề.

Runtime chỉ đổi:

```text
P.owner
```

Machine mới tiếp tục sử dụng:

- Local Queue
- Timer Heap
- mcache

đã tồn tại trên Processor.

Chi phí rất nhỏ.

---

# 13.20 Spinning Processor

## Tại sao cần Spinning?

Giả sử.

Một Processor vừa xử lý xong tất cả Goroutine.

```text
Local Queue

↓

Empty
```

Runtime có hai lựa chọn.

### Lựa chọn 1

Sleep ngay.

```text
Idle
```

Nhưng nếu ngay sau đó.

```go
go worker()
```

Runtime phải:

- Wake Thread
- Attach Processor
- Context Switch

Độ trễ tăng.

---

### Lựa chọn 2

Processor chờ thêm một khoảng thời gian ngắn.

```text
Spin

↓

Có việc?

↓

Run luôn
```

Đây chính là Spinning.

---

# Spinning là gì?

Processor vẫn đang hoạt động.

Nhưng:

```text
Không Execute Goroutine
```

Thay vào đó.

Nó tìm kiếm công việc.

```text
Local Queue

↓

Global Queue

↓

Work Stealing

↓

Netpoll
```

Nếu tìm thấy.

```text
Run ngay
```

---

# Minh họa

```text
Queue Empty

↓

Spin

↓

Steal

↓

Found G

↓

Execute
```

Không cần Sleep.

---

# Khi nào mới Sleep?

Nếu sau một khoảng thời gian.

```text
Không có việc
```

Processor mới thực sự Idle.

```text
Spin

↓

Nothing

↓

Idle
```

---

# Tại sao không Spin mãi?

Nếu Runtime luôn Spin.

```text
CPU

100%
```

Ngay cả khi chương trình không làm gì.

Điều này gây lãng phí điện năng.

Do đó Runtime giới hạn số lượng Processor được phép Spin.

---

# Idle Processor

```text
P

↓

Sleep
```

Không tiêu tốn CPU.

---

# Spinning Processor

```text
P

↓

Searching

↓

Queue

↓

Steal

↓

Netpoll
```

Tiêu tốn rất ít CPU.

Nhưng giúp giảm đáng kể độ trễ Scheduling.

---

# Trạng thái chuyển đổi

```text
Running

↓

Queue Empty

↓

Spin

↓

Có việc?

  │

Yes│No

  │

  ▼

Run

      ↓

Continue

         

No

↓

Idle
```

---

# Vai trò của Spinning trong Scheduler

Spinning giúp Runtime đạt được sự cân bằng giữa:

- Throughput
- Latency
- CPU Usage

Nếu Sleep quá sớm.

```text
Latency tăng
```

Nếu Spin quá lâu.

```text
CPU Waste
```

Go Runtime sử dụng nhiều heuristic để quyết định:

- có nên Spin không.
- có nên Wake thêm Machine không.
- có nên Sleep không.

Đây là một trong những phần tinh vi nhất của Scheduler.

---

# Sự phối hợp giữa P, M và G

Sau toàn bộ chương này, chúng ta có thể hình dung mối quan hệ giữa ba thành phần như sau.

```text
                    +------------------+
                    |    Goroutine     |
                    |       (G)        |
                    +------------------+
                              │
                              ▼
                    +------------------+
                    |   Processor (P)  |
                    |------------------|
                    | Local Run Queue  |
                    | runnext          |
                    | mcache           |
                    | Timer Heap       |
                    | GC State         |
                    +------------------+
                              │
                              ▼
                    +------------------+
                    |    Machine (M)   |
                    |  OS Thread       |
                    +------------------+
                              │
                              ▼
                             CPU
```

Trong mô hình này:

- **G** chỉ đại diện cho công việc.
- **P** quản lý toàn bộ tài nguyên và trạng thái thực thi.
- **M** chỉ chịu trách nhiệm thực thi Goroutine trên OS Thread.

Ba thành phần phối hợp với nhau tạo nên một Scheduler có khả năng:

- xử lý hàng triệu Goroutine,
- tận dụng tối đa CPU đa lõi,
- giảm Lock Contention,
- giảm Context Switch,
- giảm Allocation,
- tối ưu độ trễ và Throughput.

---

## Key Takeaways

- `Machine (M)` không thể thực thi Goroutine nếu không gắn với một `Processor (P)`.
- `Processor` mới là chủ sở hữu của Local Run Queue, `mcache`, Timer Heap và nhiều tài nguyên khác; `Machine` chỉ sử dụng các tài nguyên đó để chạy Goroutine.
- **Handoff** cho phép tách `P` khỏi một `M` đang bị Block và gắn sang một `M` khác, giúp Scheduler không bị gián đoạn bởi System Call.
- `Processor` quản lý toàn bộ vòng đời Scheduling của Goroutine: tạo mới, đưa vào Queue, đánh thức và lựa chọn Goroutine để thực thi.
- **Spinning Processor** là cơ chế giảm độ trễ bằng cách chủ động tìm kiếm công việc trước khi chuyển sang trạng thái Idle.
- Sự tách biệt rõ ràng giữa **G (Work)**, **P (Execution Context)** và **M (Execution Engine)** là nền tảng tạo nên hiệu năng và khả năng mở rộng của GMP Scheduler.

Part VII — Engineering
13.25 Best Practices
13.26 Common Mistakes
13.27 Interview Questions
13.28 Chapter Summary