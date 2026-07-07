# Chapter 12 - Machine (`struct m`)

---

# Giới thiệu

Trong Chapter 11, chúng ta đã nghiên cứu toàn bộ về **Goroutine (`struct g`)**.

Chúng ta đã hiểu:

* Goroutine được biểu diễn như thế nào.
* Execution Context được lưu ở đâu.
* Dynamic Stack hoạt động ra sao.
* Context Switching được thực hiện như thế nào.
* Scheduler nhìn Goroutine dưới góc độ nào.

Tuy nhiên, vẫn còn một câu hỏi rất quan trọng.

> **Ai là người thực sự thực thi Goroutine?**

Rất nhiều lập trình viên trả lời.

```
CPU
```

Điều này đúng nhưng chưa đầy đủ.

CPU không thể trực tiếp chạy một Goroutine.

CPU chỉ có thể chạy một:

```
OS Thread
```

Hay nói cách khác.

```
CPU

↓

OS Thread

↓

Machine

↓

Goroutine
```

Đây chính là lý do Go Runtime cần một thành phần trung gian mang tên:

```
Machine

(struct m)
```

Nếu Chapter 11 giúp chúng ta hiểu **đối tượng được thực thi**.

Thì Chapter 12 sẽ giúp chúng ta hiểu **đối tượng thực thi**.

Đây là chương giải thích cách Go Runtime quản lý các OS Thread và cách Runtime xây dựng cầu nối giữa Goroutine với hệ điều hành.

---

# Mục tiêu của chương

Sau khi hoàn thành Chapter này, bạn sẽ trả lời được các câu hỏi sau.

* Machine (`struct m`) là gì?
* Machine khác OS Thread ở điểm nào?
* Runtime tạo Thread khi nào?
* Runtime hủy Thread khi nào?
* Machine lưu những thông tin gì?
* Vì sao Runtime cần Thread Cache?
* Thread Spinning là gì?
* Thread Parking là gì?
* Runtime xử lý Blocking System Call như thế nào?
* Một Machine có thể chạy bao nhiêu Goroutine?
* Điều gì xảy ra khi Machine bị Block?
* Scheduler quản lý Machine như thế nào?

Đây là những kiến thức nền tảng trước khi chúng ta bước sang **Processor (`struct p`)** và cuối cùng là **G-M-P Scheduler**.

---

# Cấu trúc chương

Chapter 12 được chia thành tám phần.

## Phần I - Tổng quan

* 12.1 Machine là gì?
* 12.2 OS Thread vs Machine
* 12.3 Vì sao Runtime cần `struct m`
* 12.4 Mối quan hệ giữa G và M

---

## Phần II - Runtime Internals

* 12.5 `struct m`
* 12.6 Các trường quan trọng trong `struct m`
* 12.7 Curg
* 12.8 G0
* 12.9 MCache
* 12.10 Locks và trạng thái của Machine

---

## Phần III - Thread Lifecycle

* 12.11 Tạo Machine
* 12.12 Khởi tạo OS Thread
* 12.13 Machine Start
* 12.14 Machine Stop
* 12.15 Thread Destruction

---

## Phần IV - Scheduler Interaction

* 12.16 Gắn G vào M
* 12.17 Tách G khỏi M
* 12.18 Machine Blocking
* 12.19 Thread Parking
* 12.20 Thread Spinning

---

## Phần V - System Call

* 12.21 Blocking Syscall
* 12.22 Non-blocking Syscall
* 12.23 Runtime entersyscall
* 12.24 Runtime exitsyscall
* 12.25 Mối quan hệ giữa Syscall và Scheduler

---

## Phần VI - Performance

* 12.26 Thread Creation Cost
* 12.27 Thread Cache
* 12.28 Thread Reuse
* 12.29 Performance Optimization

---

## Phần VII - Debugging

* 12.30 Machine Dump
* 12.31 Thread Leak
* 12.32 Runtime Trace dưới góc nhìn Machine

---

## Phần VIII - Engineering

* 12.33 Best Practices
* 12.34 Common Mistakes
* 12.35 Interview Questions
* 12.36 Tổng kết chương

---

# Phần I - Tổng quan

Trong phần đầu tiên, chúng ta sẽ trả lời bốn câu hỏi quan trọng nhất.

1. Machine là gì?
2. Machine có phải là OS Thread không?
3. Vì sao Go Runtime cần `struct m`?
4. Machine đóng vai trò gì trong kiến trúc G-M-P?

Sau khi hoàn thành phần này, chúng ta sẽ có một bức tranh tổng thể về vai trò của Machine trước khi đi sâu vào source code của Go Runtime.

---

# 12.1 Machine là gì?

Trong Chapter trước, chúng ta đã biết rằng:

```
Goroutine

↓

Không được CPU thực thi trực tiếp.
```

CPU chỉ biết một khái niệm duy nhất.

```
OS Thread
```

Điều này có nghĩa.

Nếu Runtime muốn chạy một Goroutine.

Nó bắt buộc phải có một Thread của hệ điều hành.

Tuy nhiên.

Go Runtime **không làm việc trực tiếp với OS Thread**.

Thay vào đó.

Runtime xây dựng một lớp trừu tượng mới.

```
Machine

(struct m)
```

Có thể hình dung.

```text
Application

↓

Goroutine (G)

↓

Machine (M)

↓

OS Thread

↓

CPU
```

Machine chính là đại diện của một OS Thread bên trong Go Runtime.

Nói cách khác.

> **Mỗi `struct m` tương ứng với đúng một OS Thread.**

Nhưng.

```
Machine

≠

OS Thread
```

Đây là điểm rất quan trọng.

OS Thread là tài nguyên do hệ điều hành quản lý.

Trong khi.

Machine là một cấu trúc dữ liệu do Go Runtime quản lý.

`struct m` không chỉ lưu Thread ID.

Nó còn lưu rất nhiều thông tin khác mà Scheduler cần để vận hành.

Ví dụ.

* Goroutine hiện tại đang chạy.
* Processor đang gắn với Thread.
* G0.
* Thread Local Cache.
* Trạng thái của Thread.
* Bộ đếm Lock.
* Thông tin phục vụ Garbage Collector.

Có thể xem Machine như "hồ sơ" đầy đủ của một OS Thread bên trong Runtime.

---

# Vì sao Runtime không dùng trực tiếp OS Thread?

Nếu chỉ làm việc với Thread của hệ điều hành.

Runtime sẽ phải liên tục gọi Kernel để lấy thông tin.

Điều này:

* Tăng chi phí.
* Khó quản lý.
* Khó tối ưu Scheduler.

Thay vào đó.

Runtime chỉ tương tác với Kernel khi thật sự cần.

Mọi thông tin còn lại được lưu trong `struct m`.

Nhờ đó.

Scheduler có thể thao tác với Machine rất nhanh mà không cần gọi System Call liên tục.

---

# Vai trò của Machine

Có thể tóm tắt vai trò của Machine như sau.

```text
CPU

↓

OS Thread

↓

Machine (struct m)

↓

Processor (P)

↓

Goroutine (G)
```

Machine là cầu nối giữa:

* Hệ điều hành.
* Go Runtime.
* Scheduler.

Nếu không có Machine.

Go Runtime sẽ không thể quản lý Thread một cách hiệu quả và cũng không thể triển khai mô hình G-M-P.

---

# Những điều cần ghi nhớ

* CPU không chạy Goroutine trực tiếp.
* CPU chỉ chạy OS Thread.
* Go Runtime sử dụng `struct m` để biểu diễn một OS Thread.
* Mỗi Machine tương ứng với một OS Thread.
* Machine là thành phần trung gian giữa Kernel và Scheduler.
* `struct m` chứa nhiều thông tin hơn Thread ID và là nền tảng để Scheduler quản lý việc thực thi Goroutine.

---

# Tóm tắt

Machine (`struct m`) là đại diện của một OS Thread bên trong Go Runtime.

Thay vì thao tác trực tiếp với Thread của hệ điều hành, Runtime sử dụng `struct m` để lưu toàn bộ trạng thái cần thiết cho Scheduler, Garbage Collector và Memory Allocator.

Đây là mảnh ghép thứ hai của kiến trúc G-M-P và là nền tảng để hiểu cách Go Runtime quản lý Thread với chi phí rất thấp.

Trong phần tiếp theo, chúng ta sẽ đi sâu vào câu hỏi:

> **12.2 - OS Thread vs Machine**

Ở đó, chúng ta sẽ phân tích chi tiết sự khác nhau giữa Thread của hệ điều hành và `struct m`, cũng như lý do Go Runtime phải xây dựng thêm một lớp trừu tượng thay vì sử dụng Thread trực tiếp.

---

# 12.2 OS Thread vs Machine

Trong phần trước, chúng ta đã biết:

```text
Machine (struct m)

↓

Đại diện cho

↓

OS Thread
```

Điều này khiến nhiều người kết luận.

> **Machine chính là OS Thread.**

Đây là một kết luận chưa chính xác.

Thực tế.

```text
Machine

≠

OS Thread
```

Machine chỉ là **đối tượng Runtime dùng để quản lý một OS Thread**.

Để hiểu điều này.

Trước tiên cần hiểu CPU chỉ biết duy nhất một khái niệm.

```
OS Thread
```

Kernel hoàn toàn không biết.

- Goroutine
- Scheduler
- Processor
- struct g
- struct p
- struct m

Tất cả các khái niệm trên đều được Go Runtime xây dựng.

---

# OS Thread là gì?

OS Thread là đơn vị thực thi do hệ điều hành quản lý.

Kernel chịu trách nhiệm.

- Tạo Thread
- Hủy Thread
- Context Switch
- Scheduling
- Thread Priority
- Thread State

Có thể hình dung.

```text
Linux Kernel

↓

Thread A

↓

Thread B

↓

Thread C
```

Kernel biết mọi Thread.

Nhưng.

Kernel không hề biết Goroutine.

---

# Machine là gì?

Machine là một cấu trúc dữ liệu bên trong Go Runtime.

```go
type m struct {

    ...

}
```

Runtime dùng nó để lưu toàn bộ thông tin liên quan tới một Thread.

Ví dụ.

```text
Machine

↓

Current Goroutine

↓

Current Processor

↓

Thread Local Cache

↓

Thread State

↓

GC Information

↓

Lock Count
```

OS Thread không chứa các thông tin này.

---

# So sánh OS Thread và Machine

| OS Thread | Machine |
|------------|---------|
| Thuộc Kernel | Thuộc Go Runtime |
| Kernel tạo | Runtime tạo và quản lý |
| Kernel Scheduling | Runtime Scheduling |
| Không biết Goroutine | Quản lý Goroutine |
| Không biết Processor | Biết Processor hiện tại |
| Không có G0 | Có G0 |

---

# Góc nhìn nhiều lớp

Có thể hình dung.

```text
Application

↓

Go Runtime

↓

Machine

↓

OS Thread

↓

Kernel

↓

CPU
```

Ở đây.

Machine là cầu nối.

Giữa.

```
Go Runtime

↓

Kernel
```

---

# Vì sao không lưu mọi thứ trong Thread?

Có người sẽ hỏi.

> Runtime có thể lưu dữ liệu bằng Thread Local Storage được không?

Có thể.

Nhưng.

Không đủ.

Runtime cần quản lý.

- Scheduler
- Memory Allocator
- Garbage Collector
- Netpoll
- Timer

Những thành phần này đều cần truy cập rất nhanh.

Nếu phụ thuộc hoàn toàn vào Kernel.

Hiệu năng sẽ giảm.

Do đó.

Runtime tạo.

```
struct m
```

để lưu toàn bộ metadata.

---

# Machine luôn gắn với Thread?

Có.

Một Machine luôn đại diện cho đúng một OS Thread.

```text
M1

↓

Thread 1
```

```text
M2

↓

Thread 2
```

Điều ngược lại cũng đúng.

Một Thread cũng chỉ có một Machine.

Đây là quan hệ.

```
1 : 1
```

---

# Machine có thể đổi Thread không?

Không.

Một `struct m` không bao giờ chuyển sang Thread khác.

Nếu Runtime tạo Thread mới.

Runtime tạo luôn.

```
Machine mới
```

---

# Thread có thể đổi Machine không?

Không.

Một Thread sinh ra.

↓

Luôn thuộc về.

```
Machine đó
```

cho tới khi Thread kết thúc.

---

# Vì sao cần Machine?

Machine cho phép Runtime.

- Quản lý Scheduler.
- Quản lý GC.
- Quản lý Memory Cache.
- Quản lý Lock.
- Quản lý Thread State.

Mà không cần gọi Kernel.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Machine chính là Thread.

Sai.

Machine là Runtime Object.

Thread là Kernel Object.

---

### Hiểu lầm 2

> Machine được Kernel quản lý.

Sai.

Machine hoàn toàn do Runtime quản lý.

---

### Hiểu lầm 3

> Machine có thể chạy mà không có Thread.

Sai.

Machine luôn gắn với một OS Thread.

---

# Những điều cần ghi nhớ

- Thread là khái niệm của Kernel.
- Machine là khái niệm của Go Runtime.
- Quan hệ giữa Machine và Thread là 1:1.
- Machine chứa rất nhiều metadata mà Thread không có.
- Scheduler luôn làm việc với Machine thay vì Thread.

---

# 12.3 Vì sao Runtime cần `struct m`?

Sau khi biết.

```
Machine

↓

OS Thread
```

Một câu hỏi tự nhiên xuất hiện.

> **Tại sao Runtime không làm việc trực tiếp với OS Thread?**

Tại sao phải tạo thêm.

```
struct m
```

Đây chính là một trong những quyết định thiết kế quan trọng nhất của Go Runtime.

---

# Nếu không có Machine

Giả sử.

Runtime chỉ dùng.

```
OS Thread
```

Scheduler sẽ phải.

```text
Query Thread

↓

Query Kernel

↓

Update Thread

↓

Context Switch
```

Mỗi lần cần lấy thông tin.

Runtime lại phải phụ thuộc Kernel.

Điều này.

- Chậm.
- Khó mở rộng.
- Khó tối ưu.

---

# Runtime cần Metadata

Scheduler cần rất nhiều thông tin.

Ví dụ.

```text
Current G

Current P

Spinning?

Parking?

Lock Count?

GC State?

Signal Mask?

Thread Cache?
```

Kernel không cung cấp các thông tin này.

Runtime phải tự lưu.

---

# Machine là nơi lưu Runtime Metadata

Có thể hình dung.

```text
Machine

+----------------------+

Current G

Current P

G0

Thread ID

Locks

MCache

Signal

Profiler

GC

+----------------------+
```

Đây chính là lý do.

```
struct m

↓

rất lớn
```

so với Thread ID.

---

# Scheduler cần Machine

Scheduler không bao giờ hỏi.

```
Thread số mấy?
```

Scheduler chỉ quan tâm.

```text
Machine đang Idle?

Machine đang Spinning?

Machine đang chạy G nào?

Machine đang giữ P nào?
```

Đây đều là dữ liệu của.

```
struct m
```

---

# Garbage Collector cần Machine

GC cần biết.

- Thread nào đang chạy.
- Safe Point.
- Stack hiện tại.
- Current G.

Machine giúp GC truy cập rất nhanh.

---

# Memory Allocator cần Machine

Mỗi Machine đều có.

```
mcache
```

để giảm Lock khi cấp phát Memory.

Đây là một tối ưu rất lớn.

Chúng ta sẽ học kỹ ở chương Memory Allocator.

---

# Signal Handler cần Machine

Runtime xử lý.

- SIGPROF
- SIGURG
- SIGSEGV

đều thông qua.

```
struct m
```

Không phải Thread trực tiếp.

---

# Những điều cần ghi nhớ

- `struct m` là nơi Runtime lưu toàn bộ Thread Metadata.
- Scheduler, GC và Allocator đều phụ thuộc vào Machine.
- Machine giúp Runtime giảm phụ thuộc vào Kernel.
- Machine là nền tảng để xây dựng mô hình G-M-P.

---

# 12.4 Mối quan hệ giữa G và M

Đến thời điểm này.

Chúng ta đã biết.

```
G

=

Goroutine
```

và.

```
M

=

Machine
```

Một câu hỏi quan trọng.

> **Một Machine có thể chạy bao nhiêu Goroutine?**

---

# Quan hệ cơ bản

Một Machine.

Tại một thời điểm.

Chỉ chạy được.

```
Một Goroutine
```

Có thể hình dung.

```text
Machine

↓

Running

↓

G1
```

CPU không thể thực thi hai Goroutine trên cùng một Thread tại cùng một thời điểm.

---

# Nhưng Machine có thể chạy nhiều Goroutine

Theo thời gian.

```text
Time 1

M1

↓

G1
```

```text
Time 2

M1

↓

G8
```

```text
Time 3

M1

↓

G24
```

Machine chỉ chạy.

```
Một G tại một thời điểm.

Nhưng.

Nhiều G theo thời gian.
```

---

# Goroutine có thể đổi Machine

Điều này rất quan trọng.

Ví dụ.

```text
Time 1

G5

↓

M1
```

Sau Context Switching.

```text
Time 2

G5

↓

M3
```

Điều này hoàn toàn bình thường.

Execution Context của Goroutine nằm trong.

```
gobuf
```

không nằm trong Machine.

Do đó.

Scheduler có thể Restore Goroutine trên bất kỳ Machine nào.

---

# Machine không sở hữu Goroutine

Đây là hiểu lầm phổ biến.

Machine không "sở hữu" Goroutine.

Machine chỉ.

```
Thực thi

↓

Current G
```

Sau khi Goroutine Waiting.

Machine lập tức.

```
Run G khác
```

---

# Machine lấy Goroutine ở đâu?

Machine không tự tìm Goroutine.

Nó lấy từ.

```
Processor
```

Điều này dẫn tới mối quan hệ.

```text
G

↓

P

↓

M
```

Processor mới là nơi quản lý Run Queue.

Machine chỉ là nơi thực thi.

Chúng ta sẽ nghiên cứu chi tiết ở Chapter 13.

---

# Quan hệ G và M

Có thể mô tả.

```text
Một thời điểm

1 M

↓

1 G
```

Nhưng.

```text
Suốt vòng đời

1 M

↓

N G
```

và.

```text
1 G

↓

Có thể chạy trên

↓

N M
```

---

# Ví dụ thực tế

Giả sử.

```text
8 CPUs

↓

8 Threads

↓

100.000 Goroutines
```

Scheduler liên tục.

```text
M1

↓

G100
```

```text
M2

↓

G5
```

```text
M3

↓

G2000
```

...

Sau vài mili giây.

Tất cả lại thay đổi.

Đó chính là Concurrency.

---

# Vì sao Goroutine có thể đổi Machine?

Bởi vì.

Runtime lưu toàn bộ Execution Context trong.

```
gobuf
```

Machine chỉ cần.

```text
Restore Context

↓

Continue Execution
```

Không quan trọng.

Machine nào đang thực hiện.

---

# Chuẩn bị cho Processor

Hiện tại.

Có vẻ như.

```
Machine

↓

Tự chọn Goroutine.
```

Thực tế.

Không phải vậy.

Giữa G và M.

Còn thiếu một thành phần.

```
Processor

(struct p)
```

Processor mới là nơi.

- Quản lý Run Queue.
- Work Stealing.
- Local Cache.
- Scheduling Context.

Machine chỉ lấy Goroutine từ Processor để chạy.

Đây sẽ là nội dung của Chapter 13.

---

# Những điều cần ghi nhớ

- Một Machine chỉ chạy một Goroutine tại một thời điểm.
- Một Machine có thể chạy rất nhiều Goroutine theo thời gian.
- Một Goroutine có thể chạy trên nhiều Machine khác nhau.
- Execution Context nằm trong `gobuf`, không nằm trong Machine.
- Machine không sở hữu Goroutine.
- Processor là cầu nối giữa G và M.

---

# Tổng kết

Sau ba phần đầu tiên của Chapter 12, chúng ta đã hiểu rằng **Machine (`struct m`) không phải là OS Thread**, mà là **đại diện của OS Thread bên trong Go Runtime**. Machine tồn tại để Runtime quản lý toàn bộ metadata liên quan đến Thread, giảm sự phụ thuộc vào Kernel và hỗ trợ Scheduler, Garbage Collector cũng như Memory Allocator hoạt động hiệu quả.

Quan trọng hơn, chúng ta đã thấy rằng mối quan hệ giữa **Goroutine (`G`)** và **Machine (`M`)** không phải là 1:1. Một Machine chỉ thực thi một Goroutine tại một thời điểm, nhưng có thể lần lượt chạy hàng nghìn Goroutine. Ngược lại, một Goroutine có thể được Scheduler khôi phục và tiếp tục thực thi trên bất kỳ Machine nào nhờ Execution Context được lưu trong `gobuf`.

Tuy nhiên, vẫn còn một mắt xích quan trọng chưa xuất hiện:

> **Nếu Machine không tự quản lý Goroutine, vậy ai quyết định Goroutine nào sẽ được Machine thực thi?**

Câu trả lời nằm ở thành phần thứ ba của mô hình G-M-P:

> **Processor (`struct p`)**.

Trước khi chuyển sang Chapter 13, trong các phần tiếp theo của Chapter 12, chúng ta sẽ đi sâu vào **`struct m`**, phân tích từng trường dữ liệu bên trong và hiểu cách Runtime quản lý Thread ở mức thấp nhất.

---

---

# Phần II - Runtime Internals

Trong phần trước, chúng ta đã biết:

- Machine (`struct m`) là gì.
- Sự khác nhau giữa Machine và OS Thread.
- Vì sao Go Runtime cần `struct m`.
- Mối quan hệ giữa Goroutine (`G`) và Machine (`M`).

Tuy nhiên, tất cả mới chỉ là góc nhìn kiến trúc.

Trong phần này, chúng ta sẽ đi sâu vào source code của Go Runtime để trả lời câu hỏi:

> **Bên trong `struct m` thực sự chứa những gì?**

Đây là một trong ba cấu trúc dữ liệu quan trọng nhất của Go Runtime.

```text
struct g

↓

struct m

↓

struct p
```

Nếu `struct g` mô tả Goroutine.

Thì `struct m` mô tả toàn bộ trạng thái của một OS Thread.

Sau phần này, bạn sẽ có thể đọc được phần lớn source code liên quan đến Thread trong:

```text
runtime/runtime2.go
runtime/proc.go
runtime/os_linux.go
runtime/os_windows.go
```

---

# 12.5 `struct m`

Trong source code Go Runtime.

Machine được định nghĩa gần giống như sau.

```go
type m struct {

    g0      *g

    curg    *g

    p       puintptr

    oldp    puintptr

    mcache  *mcache

    locks   int32

    spinning bool

    blocked bool

    syscalltick uint32

    ...

}
```

Trên thực tế.

`struct m` còn nhiều trường hơn.

Tùy từng phiên bản Go.

Có thể hơn 70 trường dữ liệu.

Tuy nhiên.

Chỉ khoảng 15–20 trường là cực kỳ quan trọng đối với Scheduler.

---

# Vai trò của `struct m`

Có thể hình dung.

```text
Machine

+---------------------------+

Current Goroutine

Scheduler Stack (G0)

Current Processor

Memory Cache

Thread State

Lock Counter

Signal Info

Profiler

GC Metadata

Thread ID

+---------------------------+
```

Machine giống như "hồ sơ" đầy đủ của một Thread.

---

# Một Machine đại diện cho điều gì?

Một Machine luôn tương ứng với.

```text
1 Machine

↓

1 OS Thread
```

Machine tồn tại từ khi Thread được Runtime tạo ra.

Cho đến khi Thread kết thúc.

Trong suốt thời gian đó.

Runtime luôn sử dụng `struct m` thay vì thao tác trực tiếp với Thread.

---

# Vì sao `struct m` lớn?

Nhiều người nghĩ.

Machine chỉ cần lưu.

```text
Thread ID
```

Điều này hoàn toàn sai.

Runtime cần lưu rất nhiều thông tin.

Ví dụ.

- Scheduler
- GC
- Allocator
- Signal
- Profiler
- Stack
- Locks

Do đó.

`struct m` khá lớn.

---

# Các nhóm thông tin trong `struct m`

Có thể chia thành các nhóm.

```text
Execution

↓

Current G

G0

P
```

```text
Scheduling

↓

Spinning

Blocked

Old P
```

```text
Memory

↓

MCache
```

```text
Synchronization

↓

Locks
```

```text
Debug

↓

Trace

Profiler

Signal
```

Đây là lý do Runtime rất ít khi truy cập trực tiếp Thread của Kernel.

---

# 12.6 Các trường quan trọng trong `struct m`

Chúng ta sẽ không phân tích toàn bộ hơn 70 trường.

Thay vào đó.

Tập trung vào những trường có ảnh hưởng trực tiếp đến Scheduler.

---

# `g0`

```go
m.g0
```

Đây là Scheduler Goroutine.

Không phải Goroutine do lập trình viên tạo.

Mọi Machine đều có đúng một `g0`.

Chúng ta sẽ nghiên cứu chi tiết ở phần sau.

---

# `curg`

```go
m.curg
```

Trỏ tới.

```text
Current Running Goroutine
```

Đây là Goroutine mà Machine đang thực thi.

---

# `p`

```go
m.p
```

Trỏ tới.

```
Current Processor
```

Nếu Machine không giữ Processor.

Machine không thể chạy Goroutine.

---

# `oldp`

Trong một số trường hợp.

Machine phải tạm thời trả Processor.

Runtime lưu Processor cũ vào.

```go
oldp
```

để có thể khôi phục sau.

Ví dụ.

Blocking System Call.

---

# `mcache`

Đây là Thread Local Memory Cache.

Giúp giảm Lock khi cấp phát Memory.

Chúng ta sẽ nghiên cứu kỹ ở chương Memory Allocator.

---

# `locks`

Runtime sử dụng.

```go
locks int32
```

để đếm số Lock mà Machine đang giữ.

Thông tin này rất quan trọng với Garbage Collector.

---

# `spinning`

Một Machine có thể ở trạng thái.

```text
Spinning
```

Nghĩa là.

Machine đang tìm Goroutine để chạy.

Thay vì Sleep.

Điều này giúp giảm Latency.

---

# `blocked`

Cho biết Machine có đang bị Block hay không.

Ví dụ.

System Call.

---

# `syscalltick`

Được Runtime sử dụng để theo dõi.

```
System Call
```

và hỗ trợ Scheduler.

---

# 12.7 `curg`

Đây là trường quan trọng nhất của Machine.

```go
m.curg
```

Nó luôn trỏ tới.

```text
Current Running Goroutine
```

Ví dụ.

```text
Machine M1

↓

curg

↓

G15
```

Điều này có nghĩa.

```
Thread hiện tại

↓

đang chạy

↓

G15
```

---

# `curg` thay đổi khi nào?

Ví dụ.

```text
M1

↓

G1
```

Scheduler Context Switch.

```text
M1

↓

G8
```

Runtime cập nhật.

```go
m.curg = G8
```

---

# Machine luôn có `curg`?

Không.

Nếu Machine đang.

- Idle.
- Park.
- Chưa có Processor.

`curg` có thể bằng.

```go
nil
```

---

# `curg` và Scheduler

Scheduler luôn cập nhật.

```text
Save G1

↓

m.curg = G8

↓

Restore G8
```

Do đó.

`curg` luôn phản ánh Goroutine hiện tại.

---

# 12.8 `g0`

Đây là một trong những khái niệm quan trọng nhất của Go Runtime.

Và cũng là nơi rất nhiều lập trình viên nhầm lẫn.

---

# G0 là gì?

Mỗi Machine đều có một Goroutine đặc biệt.

```text
G0
```

Đây không phải Goroutine của người dùng.

Mà là.

```
Scheduler Stack
```

---

# Vì sao cần G0?

Giả sử.

Runtime đang Context Switching.

Nếu Runtime tiếp tục sử dụng Stack của.

```
G1
```

trong lúc.

```
Switch G1

↓

G2
```

Stack có thể bị thay đổi.

Điều này rất nguy hiểm.

Do đó.

Runtime chuyển sang.

```
G0
```

để thực hiện.

- Scheduler.
- Context Switch.
- Runtime Internal.

---

# Minh họa

```text
Machine

↓

G0

↓

Scheduler

↓

Switch

↓

G1
```

Sau khi hoàn thành.

```text
G0

↓

Restore

↓

G8
```

---

# G0 có chạy User Code không?

Không.

Không bao giờ.

G0 chỉ chạy.

- Runtime.
- Scheduler.
- GC.
- Signal.

---

# Mỗi Machine có bao nhiêu G0?

```text
1 Machine

↓

1 G0
```

Đây là quan hệ cố định.

---

# G0 và `curg`

Một Machine luôn có.

```text
g0

↓

Scheduler Stack
```

và.

```text
curg

↓

User Goroutine
```

Hai Goroutine này có nhiệm vụ hoàn toàn khác nhau.

---

# 12.9 `MCache`

Một Machine còn sở hữu.

```go
mcache
```

Đây là.

```
Thread Local Memory Cache
```

---

# Vì sao cần MCache?

Giả sử.

100 Thread.

Cùng cấp phát Memory.

Nếu tất cả cùng dùng.

```
Global Allocator
```

Mọi Thread phải Lock.

Hiệu năng giảm mạnh.

---

# Giải pháp

Mỗi Machine.

Có một Cache riêng.

```text
Machine

↓

MCache
```

Khi cấp phát Object nhỏ.

Machine lấy trực tiếp từ Cache.

Không cần Lock.

---

# Minh họa

```text
Machine 1

↓

MCache
```

```text
Machine 2

↓

MCache
```

Hai Machine.

Không tranh chấp.

---

# Khi nào dùng Global Allocator?

Nếu.

```
MCache

↓

Empty
```

Machine mới lấy thêm Memory từ.

```
Central Cache
```

Sau đó.

Tiếp tục cấp phát cục bộ.

---

# Lợi ích

- Giảm Lock.
- Giảm Cache Miss.
- Tăng Throughput.

Đây là một trong những tối ưu quan trọng nhất của Go Runtime.

---

# 12.10 Locks và trạng thái của Machine

Runtime sử dụng.

```go
m.locks
```

để theo dõi.

```
Lock Count
```

---

# Vì sao cần Lock Count?

Giả sử.

Runtime đang.

- Scheduler.
- Allocator.
- GC.

Một số đoạn mã.

Không được phép bị.

```
Preempt
```

Runtime tăng.

```text
locks++

↓

Critical Section
```

Sau đó.

```text
locks--
```

---

# Garbage Collector

GC kiểm tra.

```text
locks > 0
```

Nếu đúng.

GC biết.

Machine đang ở Critical Section.

Không nên dừng Thread.

---

# Trạng thái của Machine

Machine có nhiều trạng thái.

Ví dụ.

```text
Running
```

Đang chạy Goroutine.

---

```text
Spinning
```

Đang tìm Goroutine.

---

```text
Blocked
```

Đang System Call.

---

```text
Idle
```

Không có Processor.

---

```text
Parked
```

Thread đang Sleep.

---

# Spinning Machine

Đây là một tối ưu rất thú vị.

Thay vì.

```text
Sleep

↓

Wakeup
```

Runtime cho phép một số Machine.

```
Spinning
```

để tìm Goroutine mới.

Giúp giảm đáng kể Latency.

---

# Idle Machine

Nếu không còn Goroutine.

Machine có thể.

```
Park
```

Kernel sẽ Suspend Thread.

Tiết kiệm CPU.

---

# Blocking Machine

Nếu Machine thực hiện.

```
Blocking Syscall
```

Runtime sẽ.

- Tách Processor khỏi Machine.
- Đưa Processor cho Machine khác.

Đây là cơ chế rất quan trọng.

Chúng ta sẽ học ở phần sau.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> `curg` luôn khác `nil`.

Sai.

Machine Idle có thể không chạy Goroutine nào.

---

### Hiểu lầm 2

> `g0` là Goroutine đầu tiên của chương trình.

Sai.

`g0` là Goroutine nội bộ của Runtime.

Người dùng không thể truy cập trực tiếp.

---

### Hiểu lầm 3

> `MCache` là Global Cache.

Sai.

Mỗi Machine có một `MCache` riêng.

---

### Hiểu lầm 4

> `locks` dùng để đồng bộ giữa các Goroutine.

Sai.

`locks` phản ánh số lượng Runtime Lock mà Machine đang giữ và hỗ trợ Scheduler cùng Garbage Collector.

---

# Những điều cần ghi nhớ

- `struct m` là trung tâm quản lý Thread trong Go Runtime.
- `curg` luôn trỏ tới Goroutine đang chạy.
- Mỗi Machine có đúng một `g0`.
- `g0` chỉ chạy mã của Runtime và Scheduler.
- `MCache` giúp giảm Lock khi cấp phát Memory.
- `locks` giúp Runtime bảo vệ các Critical Section.
- Machine có nhiều trạng thái như Running, Spinning, Idle và Blocked.

---

# Tổng kết

Trong phần này, chúng ta đã đi sâu vào **`struct m`** và phân tích những trường dữ liệu quan trọng nhất của Machine.

Chúng ta đã hiểu vai trò của:

- `curg` trong việc theo dõi Goroutine đang chạy.
- `g0` như Scheduler Stack của mỗi Thread.
- `mcache` trong việc tối ưu Memory Allocation.
- `locks` và trạng thái của Machine để phối hợp với Scheduler và Garbage Collector.

Đây mới chỉ là cấu trúc dữ liệu.

Trong phần tiếp theo, chúng ta sẽ nghiên cứu **vòng đời của một Machine**:

> **Machine được Runtime tạo ra khi nào?**

> **Thread được khởi động và kết thúc như thế nào?**

Đó sẽ là nội dung của **Phần III - Thread Lifecycle**.

---

---

# Phần III - Thread Lifecycle

Trong hai phần trước, chúng ta đã tìm hiểu:

- Machine (`struct m`) là gì.
- Vai trò của Machine trong Go Runtime.
- Cấu trúc của `struct m`.
- `curg`, `g0`, `mcache`.
- Trạng thái của Machine.

Tuy nhiên.

Một câu hỏi rất quan trọng vẫn chưa được trả lời.

> **Machine được tạo ra khi nào?**

Hay nói cách khác.

> **Go Runtime tạo OS Thread như thế nào?**

Đây là điểm khác biệt rất lớn giữa Go Runtime và nhiều Runtime khác.

Go không tạo trước hàng trăm Thread.

Runtime chỉ tạo Thread khi thật sự cần.

Đây là một trong những nguyên nhân giúp Go có khả năng mở rộng rất tốt.

Trong phần này.

Chúng ta sẽ nghiên cứu toàn bộ vòng đời của một Machine.

```text
Create M

↓

Create OS Thread

↓

Start

↓

Run Goroutines

↓

Idle

↓

Stop

↓

Destroy
```

Đây chính là Thread Lifecycle trong Go Runtime.

---

# 12.11 Tạo Machine

Machine không được tạo khi chúng ta gọi.

```go
go worker()
```

Đây là một hiểu lầm rất phổ biến.

Ví dụ.

```go
for i := 0; i < 100000; i++ {

    go worker()

}
```

Không có nghĩa Runtime sẽ tạo.

```
100000 Threads
```

Thực tế.

Runtime chỉ tạo thêm Machine khi.

```
Scheduler thật sự cần thêm Thread.
```

---

# Khi nào Runtime tạo Machine?

Một số trường hợp phổ biến.

- Khởi động chương trình.
- Không còn Machine khả dụng.
- Blocking System Call.
- CGO.
- Runtime cần thêm Thread để duy trì Scheduler.

Không phải mỗi Goroutine đều tạo Machine.

---

# Machine đầu tiên

Khi chương trình bắt đầu.

Go Runtime khởi tạo.

```text
M0
```

Đây là Machine đầu tiên của toàn bộ tiến trình.

Nó thường được gọi là.

```
Bootstrap Machine
```

M0 được tạo rất sớm.

Trước cả khi.

```
main.main()
```

được gọi.

---

# M0 có gì đặc biệt?

M0 khác các Machine khác ở nhiều điểm.

Ví dụ.

- Không được Runtime hủy.
- Dùng để khởi động Runtime.
- Khởi tạo Scheduler.
- Khởi tạo Garbage Collector.
- Khởi tạo Main Goroutine.

Có thể xem.

```
M0
```

là "Thread gốc" của Go Runtime.

---

# Runtime quyết định tạo Machine

Giả sử.

```text
P0

↓

Run Queue đầy
```

Trong khi.

Không còn Machine nào rảnh.

Scheduler có thể quyết định.

```text
Create M
```

Sau đó.

Machine mới sẽ nhận Processor.

Và tiếp tục chạy Goroutine.

---

# Tạo `struct m`

Bước đầu tiên.

Runtime cấp phát.

```text
struct m
```

Có thể hình dung.

```text
Allocate

↓

struct m

↓

Initialize Fields
```

Lúc này.

Machine vẫn chưa có Thread.

---

# Tạo G0

Ngay sau đó.

Runtime tạo.

```
G0
```

cho Machine.

```text
Machine

↓

G0
```

Đây sẽ là Scheduler Stack.

---

# Chưa có Thread

Sau bước này.

Machine chỉ tồn tại trong Runtime.

Kernel vẫn chưa biết gì.

Tiếp theo.

Runtime mới yêu cầu hệ điều hành tạo Thread.

---

# 12.12 Khởi tạo OS Thread

Sau khi có.

```
struct m
```

Runtime thực hiện.

```
newosproc()
```

Đây là hàm chịu trách nhiệm tạo Thread của hệ điều hành.

---

# `newosproc()`

Tên hàm có thể khác nhau theo từng nền tảng.

Ví dụ.

Linux.

```text
clone()
```

Windows.

```text
CreateThread()
```

macOS.

```text
pthread_create()
```

Go Runtime che giấu sự khác biệt này.

Scheduler chỉ cần gọi.

```
newosproc()
```

---

# Sau khi Kernel tạo Thread

Kernel trả về.

```
Thread Handle
```

Runtime cập nhật.

```text
Machine

↓

OS Thread
```

Lúc này.

Quan hệ.

```
1 Machine

↓

1 Thread
```

được thiết lập.

---

# Thread bắt đầu từ đâu?

Thread mới không chạy.

```
main()
```

Nó bắt đầu từ một Entry Point của Runtime.

Ví dụ.

```text
mstart()
```

Đây là hàm khởi động mọi Machine.

---

# Tại sao không chạy Goroutine ngay?

Bởi vì.

Thread mới chưa có.

- Processor.
- Current Goroutine.

Runtime phải hoàn tất quá trình khởi tạo trước.

---

# 12.13 Machine Start

Sau khi Thread được tạo.

Runtime bắt đầu.

```
Machine Start
```

Có thể mô tả.

```text
Create Thread

↓

Initialize G0

↓

Bind Thread

↓

mstart()

↓

Scheduler Loop
```

---

# `mstart()`

Đây là Entry Point của mọi Machine.

Nhiệm vụ.

- Chuẩn bị Stack.
- Thiết lập TLS.
- Gắn G0.
- Khởi động Scheduler.

Sau đó.

Machine bắt đầu chạy.

---

# Scheduler Loop

Sau khi.

```
mstart()
```

Machine đi vào vòng lặp.

```text
Find Runnable G

↓

Run

↓

Find Runnable G

↓

Run
```

Có thể kéo dài.

Cho đến khi Thread kết thúc.

---

# Gắn Processor

Machine chưa thể chạy Goroutine nếu chưa có.

```
Processor
```

Runtime thực hiện.

```text
Acquire P

↓

Scheduler Loop
```

Chúng ta sẽ học chi tiết ở Chapter 13.

---

# Machine đầu tiên chạy gì?

Machine không chạy.

```
main()
```

ngay lập tức.

Nó chạy.

```
G0

↓

Scheduler

↓

Main Goroutine
```

Điều này rất quan trọng.

Ngay cả.

```
main()
```

cũng là một Goroutine.

---

# 12.14 Machine Stop

Machine không phải lúc nào cũng chạy.

Nếu không còn việc.

Runtime có thể dừng Machine.

---

# Khi nào Machine Stop?

Ví dụ.

- Không còn Runnable Goroutine.
- Không có Processor.
- Runtime giảm số Thread.
- Chờ sự kiện mới.

---

# Machine Park

Thông thường.

Runtime không hủy Machine ngay.

Thay vào đó.

```text
Running

↓

Idle

↓

Park
```

Kernel Suspend Thread.

---

# Vì sao không hủy ngay?

Tạo Thread rất tốn chi phí.

Nếu.

```text
Destroy

↓

Create Again
```

liên tục.

Hiệu năng giảm.

Do đó.

Runtime ưu tiên.

```
Reuse Thread
```

---

# Thread Parking

Có thể hình dung.

```text
Machine

↓

Sleep

↓

Wakeup

↓

Continue
```

Thời gian Wakeup.

Nhanh hơn rất nhiều so với.

```
Create Thread mới
```

---

# Ai đánh thức Machine?

Scheduler.

Hoặc.

Runtime.

Khi xuất hiện.

```
Runnable Goroutine
```

Machine được Wakeup.

---

# 12.15 Thread Destruction

Cuối cùng.

Một Machine có thể bị hủy.

Đây là trường hợp ít gặp hơn.

---

# Khi nào Thread bị hủy?

Ví dụ.

- Runtime Shutdown.
- Thread không còn cần thiết.
- Runtime quyết định thu hẹp Thread Pool.
- Một số trường hợp đặc biệt của CGO.

---

# Quá trình hủy

Có thể mô tả.

```text
Stop Scheduler

↓

Release Processor

↓

Release MCache

↓

Cleanup Runtime State

↓

Exit Thread
```

Sau đó.

Kernel thu hồi.

```
OS Thread
```

---

# Điều gì xảy ra với `struct m`?

Sau khi Thread kết thúc.

Runtime thu hồi.

```text
Machine

↓

GC

hoặc

Reuse
```

Tùy theo từng trường hợp.

---

# Machine có bị GC không?

Không trực tiếp.

`struct m` là đối tượng đặc biệt của Runtime.

Không phải Object thông thường trên Heap.

Runtime tự quản lý vòng đời của Machine.

---

# Vì sao Runtime ít hủy Thread?

Thread Creation rất đắt.

Bao gồm.

- Kernel Call.
- TLS.
- Stack.
- Scheduling.

Do đó.

Runtime cố gắng.

```
Reuse

>

Destroy
```

---

# Vòng đời đầy đủ

Có thể mô tả.

```text
Allocate struct m

↓

Create G0

↓

Create OS Thread

↓

mstart()

↓

Acquire Processor

↓

Run Goroutines

↓

Idle

↓

Park

↓

Wakeup

↓

Continue

↓

Stop

↓

Destroy
```

Đây là toàn bộ vòng đời của một Machine.

---

# Mối quan hệ với Goroutine

Điều thú vị.

Machine.

Có thể được tạo.

↓

Trước.

↓

Khi có Goroutine.

Hoặc.

Sau khi.

Có rất nhiều Goroutine.

Runtime luôn điều chỉnh.

```
Số Machine

↓

Theo nhu cầu thực tế.
```

---

# Tại sao Runtime không tạo sẵn nhiều Thread?

Ví dụ.

Máy có.

```
8 CPUs
```

Không có nghĩa.

Runtime tạo.

```
100 Threads
```

ngay từ đầu.

Điều này sẽ.

- Lãng phí RAM.
- Tăng Scheduler Cost.
- Tăng Context Switching.

Runtime áp dụng chiến lược.

```
Lazy Creation
```

Chỉ tạo thêm Machine khi thật sự cần.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Mỗi Goroutine tạo một Machine.

Sai.

Một Machine có thể chạy hàng nghìn Goroutine trong suốt vòng đời của nó.

---

### Hiểu lầm 2

> Thread được tạo ngay khi gọi từ khóa `go`.

Sai.

Việc tạo Goroutine và tạo Thread là hai quá trình hoàn toàn độc lập.

---

### Hiểu lầm 3

> Runtime hủy Thread ngay khi Machine rảnh.

Sai.

Thông thường Runtime sẽ Park Thread để tái sử dụng.

---

### Hiểu lầm 4

> `main.main()` là hàm đầu tiên chạy trên Thread.

Sai.

Thread luôn bắt đầu từ mã khởi động của Runtime (`mstart()` và các hàm liên quan), sau đó Scheduler mới lập lịch cho Main Goroutine.

---

### Hiểu lầm 5

> Thread Creation rất rẻ.

Sai.

Tạo Thread yêu cầu Kernel tham gia và có chi phí cao hơn rất nhiều so với tạo Goroutine.

---

# Những điều cần ghi nhớ

- Machine được tạo độc lập với Goroutine.
- `M0` là Bootstrap Machine của Go Runtime.
- Mỗi Machine đều có một `G0`.
- Runtime tạo OS Thread thông qua các API của hệ điều hành.
- Mọi Thread đều bắt đầu từ `mstart()`.
- Runtime ưu tiên Park và tái sử dụng Thread thay vì hủy.
- Thread chỉ bị hủy trong một số trường hợp đặc biệt.

---

# Tổng kết

Trong phần này, chúng ta đã nghiên cứu toàn bộ **vòng đời của Machine**, từ khi Runtime cấp phát `struct m`, tạo `G0`, khởi tạo OS Thread, khởi động Scheduler cho đến khi Machine dừng và Thread được thu hồi.

Một điểm quan trọng cần nhớ là **Machine không tự tìm việc để làm**. Sau khi khởi động, Machine phải **gắn với một `Processor (P)`** thì mới có thể lấy Goroutine để thực thi.

Điều này dẫn đến câu hỏi quan trọng tiếp theo:

> **Machine lấy Processor từ đâu?**

> **Điều gì xảy ra nếu Machine bị Block bởi một System Call?**

Đó sẽ là nội dung của **Phần IV - Scheduler Interaction**, nơi chúng ta sẽ phân tích cách Machine phối hợp với Processor và Goroutine để tạo nên mô hình thực thi G-M-P của Go Runtime.

---

---

# Phần IV - Scheduler Interaction

Trong các phần trước, chúng ta đã tìm hiểu:

- Machine (`struct m`)
- Cấu trúc của Machine
- Vòng đời của Machine
- Thread Lifecycle

Đến thời điểm này, chúng ta đã biết.

- Machine đại diện cho một OS Thread.
- Mỗi Machine có một `g0`.
- Machine được Runtime tạo khi cần.
- Machine được Kernel thực thi.

Nhưng vẫn còn một câu hỏi rất lớn.

> **Machine thực sự chạy Goroutine như thế nào?**

Hay nói cách khác.

> **Scheduler làm cách nào để đưa một Goroutine lên một Machine?**

Đây chính là lúc ba thành phần của Go Runtime bắt đầu phối hợp.

```text
G

↓

P

↓

M
```

Trong phần này, chúng ta sẽ tập trung vào mối quan hệ giữa:

- Goroutine (`G`)
- Machine (`M`)

Processor (`P`) sẽ được giải thích chi tiết ở Chapter 13.

---

# 12.16 Gắn G vào M

Sau khi Scheduler chọn được một Goroutine.

Một câu hỏi xuất hiện.

> **Làm thế nào Goroutine bắt đầu chạy trên một Machine?**

Điều đầu tiên cần hiểu.

Machine không tự tìm Goroutine.

Runtime cũng không "copy" Goroutine vào Thread.

Thay vào đó.

Runtime chỉ thay đổi một Pointer.

---

# `curg`

Trong `struct m`.

Có trường.

```go
curg *g
```

Đây là.

```
Current Running Goroutine
```

Ví dụ.

Ban đầu.

```text
Machine

↓

curg = nil
```

Sau khi Scheduler chọn.

```
G15
```

Runtime thực hiện.

```go
m.curg = G15
```

Lúc này.

Machine biết.

```
Current Goroutine

=

G15
```

---

# Chỉ thay đổi Pointer?

Đúng.

Việc "gắn Goroutine vào Machine" về bản chất chỉ là.

```text
Machine

↓

curg

↓

G15
```

Không có:

- Copy Stack.
- Copy Heap.
- Copy Object.

Execution Context vẫn nằm trong.

```
struct g
```

---

# Restore Context

Sau khi.

```go
m.curg = G15
```

Runtime thực hiện.

```text
Read gobuf

↓

Restore Registers

↓

Restore SP

↓

Restore PC
```

CPU lập tức tiếp tục.

```
G15
```

từ đúng vị trí trước đó.

---

# Gắn Goroutine đầu tiên

Ví dụ.

```go
func main() {

}
```

Ngay cả.

```
main.main()
```

cũng phải trải qua.

```text
Scheduler

↓

Machine

↓

curg

↓

Main Goroutine
```

---

# Scheduler không gọi Function

Một hiểu lầm phổ biến.

Nhiều người nghĩ.

```text
Scheduler

↓

Call Function
```

Điều này sai.

Scheduler chỉ.

- Gắn `curg`.
- Restore Context.

Sau đó.

CPU tiếp tục thực thi.

---

# Minh họa

```text
Run Queue

↓

G8

↓

Scheduler

↓

M1.curg = G8

↓

Restore Context

↓

Running
```

Đây là toàn bộ quá trình.

---

# 12.17 Tách G khỏi M

Nếu có bước.

```
Attach
```

thì chắc chắn phải có.

```
Detach
```

Một Goroutine không thể gắn với Machine mãi mãi.

---

# Khi nào tách?

Ví dụ.

- Channel Receive.
- Mutex Lock.
- Sleep.
- System Call.
- Preemption.
- Goroutine Exit.

Lúc này.

Machine cần chạy Goroutine khác.

---

# Save Context

Trước khi tách.

Runtime thực hiện.

```text
Save Registers

↓

Save SP

↓

Save PC

↓

Update gobuf
```

Execution Context đã an toàn.

---

# Xóa `curg`

Sau đó.

Runtime.

```go
m.curg = nil
```

Hoặc.

```go
m.curg = G khác
```

Nếu Scheduler lập tức chạy Goroutine mới.

---

# Goroutine không mất gì

Điều thú vị.

Sau khi bị tách.

```
G15
```

vẫn giữ nguyên.

- Stack.
- Heap.
- Registers (trong gobuf).
- State.

Do đó.

Runtime có thể.

```
Resume

↓

bất kỳ lúc nào.
```

---

# Tách không có nghĩa là kết thúc

Một Goroutine.

```text
Detach

↓

Waiting
```

hoặc.

```text
Detach

↓

Runnable
```

Sau đó.

Có thể tiếp tục chạy.

---

# 12.18 Machine Blocking

Đây là phần rất quan trọng.

Đồng thời cũng là lý do Go Runtime tạo ra mô hình.

```
G-M-P
```

---

# Blocking là gì?

Giả sử.

```go
syscall.Read(...)
```

Kernel cần.

```
100 ms
```

để trả dữ liệu.

Nếu Thread bị Block.

Điều gì xảy ra?

---

# Nếu không có Go Runtime

Trong mô hình truyền thống.

```text
Thread

↓

Read()

↓

Block

↓

100 ms
```

CPU không thể chạy công việc khác trên Thread này.

---

# Go Runtime làm gì?

Runtime phát hiện.

```
Blocking Syscall
```

Thực hiện.

```text
Detach Processor

↓

Machine Block

↓

Processor

↓

Machine khác
```

Điều này cực kỳ quan trọng.

---

# Machine bị Block

Ví dụ.

```text
M1

↓

Read()

↓

Blocked
```

Runtime.

Không chờ.

```
M1
```

Thay vào đó.

```text
P

↓

M2
```

Machine mới tiếp tục chạy Goroutine khác.

---

# Blocking không làm Scheduler dừng

Đây là thiết kế quan trọng nhất.

Nếu.

```
M1

↓

Blocked
```

Scheduler vẫn tiếp tục.

```text
M2

↓

Run G
```

Đó là lý do.

Một lời gọi File I/O.

Không làm toàn bộ chương trình dừng.

---

# Sau khi System Call kết thúc

Kernel đánh thức.

```
M1
```

Runtime.

```text
M1

↓

Runnable
```

Nếu còn Processor.

M1 tiếp tục.

Nếu không.

M1 chờ.

---

# 12.19 Thread Parking

Không phải lúc nào Machine cũng có việc.

Nếu.

- Không có Runnable Goroutine.
- Không có Processor.

Runtime có thể.

```
Park Machine
```

---

# Parking là gì?

Có thể hình dung.

```text
Running

↓

Idle

↓

Sleep
```

Kernel Suspend Thread.

---

# Vì sao không Destroy?

Thread Creation rất đắt.

Runtime ưu tiên.

```
Park

>

Destroy
```

Machine được giữ lại.

Để tái sử dụng.

---

# Wakeup

Khi xuất hiện.

```
Runnable G
```

Runtime.

```text
Wake Thread

↓

Continue Scheduler
```

Điều này nhanh hơn.

```
Create Thread mới
```

---

# Parking và G0

Thread Park.

Không có nghĩa.

```
G0

↓

Destroy
```

Mọi dữ liệu.

- G0.
- MCache.
- Metadata.

vẫn tồn tại.

---

# Thread Parking trong Production

Một Server.

Ban đêm.

Có thể.

```text
100 Machines

↓

10 Active

↓

90 Park
```

Sáng hôm sau.

Runtime Wakeup.

Các Machine.

Không cần tạo mới.

---

# 12.20 Thread Spinning

Parking giúp tiết kiệm CPU.

Nhưng.

Wakeup cũng tốn thời gian.

Runtime tạo thêm một tối ưu.

```
Spinning Thread
```

---

# Spinning là gì?

Thay vì.

```text
Sleep
```

Machine.

```text
Running

↓

Searching Runnable G
```

Trong thời gian ngắn.

---

# Vì sao?

Giả sử.

```text
Sleep

↓

Wake

↓

Sleep

↓

Wake
```

liên tục.

Latency tăng.

Runtime để một vài Machine.

```
Spinning
```

Nếu Goroutine mới xuất hiện.

Machine lập tức chạy.

Không cần Wakeup.

---

# Spinning Machine làm gì?

Có thể mô tả.

```text
Find Local Queue

↓

Find Global Queue

↓

Work Stealing

↓

Still Empty?

↓

Park
```

Đây là vòng đời của.

```
Spinning Thread
```

---

# Runtime giới hạn số Spinning Machine

Nếu.

```
100 Threads

↓

Spinning
```

CPU sẽ bị lãng phí.

Do đó.

Runtime chỉ cho phép.

```
Một số ít

Machine

↓

Spinning
```

Các Machine còn lại.

```
Park
```

---

# Lợi ích

Spinning giúp.

- Giảm Wakeup Cost.
- Giảm Latency.
- Tăng Throughput.

Đây là một tối ưu quan trọng của Scheduler hiện đại.

---

# Vòng đời Machine

Kết hợp tất cả.

Có thể mô tả.

```text
Start

↓

Acquire P

↓

Attach G

↓

Run

↓

Detach G

↓

Run Next

↓

Blocking?

↓

Park?

↓

Spinning?

↓

Continue
```

Đây là vòng đời thực tế của một Machine.

---

# Mối quan hệ G và M

Có thể hình dung.

```text
G1

↓

M1

↓

Detach
```

```text
G2

↓

M1
```

```text
G1

↓

M3
```

Điều này hoàn toàn bình thường.

Execution Context luôn nằm trong.

```
gobuf
```

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Gắn Goroutine vào Machine là sao chép Goroutine.

Sai.

Runtime chỉ cập nhật `m.curg` và khôi phục Execution Context từ `gobuf`.

---

### Hiểu lầm 2

> Goroutine luôn chạy trên cùng một Machine.

Sai.

Sau mỗi Context Switch, Scheduler có thể chạy Goroutine trên bất kỳ Machine nào có Processor.

---

### Hiểu lầm 3

> Blocking System Call làm toàn bộ Scheduler dừng lại.

Sai.

Runtime tách Processor khỏi Machine bị Block và giao Processor đó cho một Machine khác.

---

### Hiểu lầm 4

> Thread Parking và Thread Destruction giống nhau.

Sai.

Parking chỉ tạm dừng Thread để tái sử dụng.

Destruction là giải phóng hoàn toàn Thread.

---

### Hiểu lầm 5

> Spinning luôn tốt hơn Parking.

Sai.

Spinning giảm Latency nhưng tiêu tốn CPU.

Runtime phải cân bằng giữa Spinning và Parking để đạt hiệu năng tối ưu.

---

# Những điều cần ghi nhớ

- Machine chạy Goroutine thông qua trường `curg`.
- Attach và Detach Goroutine chỉ thay đổi trạng thái và Pointer, không sao chép dữ liệu.
- Blocking System Call chỉ làm Block Machine, không làm dừng toàn bộ Scheduler.
- Runtime ưu tiên Park và Wakeup Thread thay vì tạo mới.
- Spinning là kỹ thuật giảm Latency bằng cách giữ một số Machine ở trạng thái sẵn sàng tìm Goroutine.
- Processor là thành phần quyết định Goroutine nào sẽ được Machine thực thi.

---

# Tổng kết

Trong phần này, chúng ta đã hiểu cách Machine tương tác với Goroutine trong quá trình lập lịch.

Machine không sở hữu Goroutine mà chỉ **tạm thời thực thi** Goroutine thông qua `curg`. Khi Goroutine bị Block, Waiting hoặc bị Preempt, Runtime sẽ lưu Execution Context, tách Goroutine khỏi Machine và tiếp tục sử dụng Machine đó để chạy công việc khác.

Chúng ta cũng đã tìm hiểu hai cơ chế tối ưu quan trọng của Go Runtime:

- **Thread Parking** để giảm tiêu thụ tài nguyên khi hệ thống nhàn rỗi.
- **Thread Spinning** để giảm độ trễ khi có Goroutine mới xuất hiện.

Tuy nhiên, một câu hỏi quan trọng vẫn còn bỏ ngỏ:

> **Điều gì thực sự xảy ra khi Machine thực hiện một Blocking System Call?**

> **Runtime làm thế nào để vừa chờ Kernel, vừa tiếp tục sử dụng CPU hiệu quả?**

Đó sẽ là nội dung của **Phần V - System Call**, nơi chúng ta sẽ phân tích chi tiết các hàm `entersyscall()` và `exitsyscall()` trong Go Runtime.

---

---

# Phần V - System Call

Trong các phần trước, chúng ta đã tìm hiểu:

- Machine (`struct m`)
- Thread Lifecycle
- Scheduler Interaction
- Thread Parking
- Thread Spinning

Đến đây, chúng ta đã biết rằng:

> **Một Machine chính là một OS Thread.**

Điều này dẫn đến một câu hỏi rất quan trọng.

> **Điều gì xảy ra nếu OS Thread bị Block trong một System Call?**

Đây là một trong những thách thức lớn nhất mà mọi Runtime đều phải giải quyết.

Nếu xử lý không tốt.

Chỉ một lời gọi:

```go
Read()
```

có thể làm dừng toàn bộ Scheduler.

Go Runtime đã giải quyết vấn đề này bằng một thiết kế rất thông minh.

Trong phần này.

Chúng ta sẽ nghiên cứu.

- Blocking System Call
- Non-blocking System Call
- entersyscall()
- exitsyscall()
- Cách Scheduler phối hợp với Kernel

Đây là một trong những phần quan trọng nhất để hiểu vì sao Go có thể xử lý hàng triệu kết nối đồng thời.

---

# 12.21 Blocking Syscall

Trước tiên.

Chúng ta cần hiểu.

```
System Call
```

là gì.

---

# System Call là gì?

Go Runtime chạy trong.

```
User Space
```

Trong khi.

Một số thao tác chỉ Kernel mới được phép thực hiện.

Ví dụ.

- Đọc File.
- Ghi File.
- Network.
- Socket.
- Process.
- Memory Mapping.

Khi cần thực hiện các thao tác này.

Runtime phải gọi.

```
System Call
```

---

# Ví dụ

```go
file.Read(buf)
```

Hoặc.

```go
syscall.Read(...)
```

Lúc này.

CPU chuyển từ.

```text
User Mode

↓

Kernel Mode
```

Kernel bắt đầu xử lý.

---

# Blocking Syscall là gì?

Một số System Call.

Không trả kết quả ngay.

Ví dụ.

```go
Read(socket)
```

Nếu chưa có dữ liệu.

Kernel sẽ.

```
Wait
```

Trong thời gian này.

Thread bị Block.

---

# Góc nhìn của Kernel

Có thể mô tả.

```text
Thread

↓

Read()

↓

Kernel

↓

Waiting Data
```

Kernel không thể tiếp tục Thread.

Cho đến khi dữ liệu xuất hiện.

---

# Nếu không có Go Runtime

Giả sử.

```text
Thread A

↓

Read()

↓

Block
```

Thread không thể.

- Chạy Function khác.
- Chạy Task khác.
- Chạy Goroutine khác.

CPU bị lãng phí.

---

# Điều gì xảy ra trong Go?

Go Runtime phát hiện.

```
Blocking Syscall
```

Trước khi Thread đi vào Kernel.

Runtime thực hiện.

```text
Detach Processor

↓

Save State

↓

Enter Kernel
```

Điều này cực kỳ quan trọng.

Processor không bị Block.

---

# Minh họa

Ban đầu.

```text
G1

↓

P1

↓

M1
```

Sau khi.

```
Read()
```

```text
G1

↓

M1 (Blocked)

P1

↓

M2
```

Scheduler tiếp tục chạy.

```
G2
```

---

# Tại sao cần tách Processor?

Nếu Processor vẫn nằm trên.

```
M1
```

Toàn bộ Goroutine khác sẽ không có nơi để chạy.

Go Runtime tránh điều này bằng cách.

```
Release P
```

---

# 12.22 Non-blocking Syscall

Không phải mọi System Call đều Block.

Ví dụ.

```go
GetPID()
```

Hoặc.

```go
GetTime()
```

Những lời gọi này.

Thường hoàn thành rất nhanh.

---

# Runtime xử lý như thế nào?

Nếu Runtime biết.

```
Syscall

↓

Very Short
```

Thread chỉ chuyển sang Kernel.

Rồi quay trở lại.

Không cần thay đổi Scheduler.

---

# Network có phải Blocking?

Đây là câu hỏi rất hay.

Trong Go.

Network Socket.

Thông thường được đặt ở chế độ.

```
Non-blocking
```

Kết hợp với.

```
Netpoll
```

Điều này có nghĩa.

Thread không bị Block lâu.

Thay vào đó.

Runtime sử dụng.

- epoll
- kqueue
- IOCP

để chờ sự kiện.

Chúng ta sẽ nghiên cứu ở Chapter 17.

---

# Blocking và Non-blocking

Có thể so sánh.

```text
Blocking

↓

Thread Wait
```

```text
Non-blocking

↓

Return Immediately
```

Hoặc.

```text
Netpoll

↓

Wakeup Later
```

---

# 12.23 Runtime `entersyscall`

Đây là một trong những hàm quan trọng nhất của Scheduler.

Khi Runtime biết.

Thread sắp thực hiện.

```
Blocking Syscall
```

Runtime gọi.

```
entersyscall()
```

---

# Mục tiêu

Mục tiêu của.

```
entersyscall()
```

không phải.

```
Call Kernel
```

Mà là.

```
Chuẩn bị cho Scheduler.
```

---

# Các bước

Có thể mô tả.

```text
Save State

↓

Detach P

↓

Update M

↓

Update G

↓

Enter Kernel
```

---

# Save Execution Context

Runtime trước tiên.

```text
Save Registers

↓

Save PC

↓

Save SP
```

Đảm bảo Goroutine có thể Resume sau này.

---

# Đổi trạng thái Goroutine

Runtime.

```text
Running

↓

Syscall
```

Lúc này.

Scheduler biết.

```
G đang trong Kernel.
```

---

# Tách Processor

Đây là bước quan trọng nhất.

```text
M1

↓

Release P1
```

Processor.

```
P1
```

trở lại Scheduler.

---

# Machine tiếp tục vào Kernel

Sau khi mất Processor.

Machine.

```
M1
```

vẫn tiếp tục.

```text
Kernel

↓

Read()
```

---

# Scheduler

Lúc này.

Scheduler nhìn thấy.

```text
P1

↓

Available
```

Runtime lập tức.

```text
P1

↓

M2
```

Machine mới tiếp tục chạy Goroutine khác.

---

# 12.24 Runtime `exitsyscall`

Sau khi Kernel hoàn thành.

Thread quay lại.

```
User Space
```

Runtime gọi.

```
exitsyscall()
```

---

# Mục tiêu

Đưa Machine trở lại Scheduler.

---

# Các bước

```text
Return Kernel

↓

Acquire P

↓

Restore State

↓

Runnable
```

---

# Acquire Processor

Nếu còn Processor.

Runtime.

```text
P

↓

M1
```

Machine tiếp tục.

---

# Nếu không còn Processor?

Ví dụ.

Tất cả Processor đang bận.

Runtime.

```text
M1

↓

Wait
```

Cho đến khi.

Có Processor rảnh.

---

# Restore Context

Sau khi có Processor.

Runtime.

```text
Restore Registers

↓

Restore PC

↓

Continue
```

Goroutine tiếp tục chạy.

---

# Không phải luôn quay lại Goroutine cũ

Đây là điểm rất thú vị.

Sau khi.

```
Read()
```

kết thúc.

Machine không bắt buộc phải chạy lại.

```
G1
```

Scheduler có thể quyết định.

```
G5
```

```
G8
```

hoặc.

```
G100
```

Tùy tình hình.

---

# 12.25 Mối quan hệ giữa Syscall và Scheduler

Đây là phần kết nối toàn bộ chương.

---

# Không có Scheduler

Nếu Runtime không có Scheduler.

Một Blocking Syscall.

```text
Read()

↓

Block Thread

↓

CPU Idle
```

---

# Có Scheduler

Go Runtime.

```text
Read()

↓

entersyscall()

↓

Release P

↓

Run G khác
```

Trong khi.

Kernel vẫn xử lý.

```
Read()
```

---

# Khi Kernel hoàn tất

```text
Read Done

↓

exitsyscall()

↓

Acquire P

↓

Continue
```

Đây là lý do.

Một Goroutine Block.

Không làm toàn bộ chương trình Block.

---

# Góc nhìn tổng thể

Có thể mô tả.

```text
Running

↓

entersyscall()

↓

Release P

↓

Kernel

↓

Waiting

↓

Kernel Return

↓

exitsyscall()

↓

Acquire P

↓

Running
```

---

# Scheduler luôn ưu tiên Processor

Điều thú vị.

Scheduler không quan tâm.

```
Thread
```

Nó quan tâm.

```
Processor
```

Miễn là.

```
P

↓

Available
```

Scheduler luôn tìm Machine phù hợp.

---

# Vì sao đây là thiết kế quan trọng?

Nếu không tách Processor.

Một Database Query.

Có thể làm.

```
8 CPUs

↓

Idle
```

Go Runtime tránh điều này.

Bằng cách.

```
Detach P
```

---

# Liên hệ với Netpoll

Trong thực tế.

Phần lớn Network I/O.

Không đi theo.

```
Blocking Syscall
```

Mà đi qua.

```
Netpoll
```

Điều này giúp.

Số Thread.

Giữ ở mức rất thấp.

Ngay cả khi.

```
1.000.000 Connections
```

---

# Minh họa toàn bộ

```text
G1

↓

M1

↓

P1

↓

Read()

↓

entersyscall()

↓

Release P1

↓

M1 Block

↓

P1

↓

M2

↓

Run G2

↓

Read Done

↓

exitsyscall()

↓

Acquire P

↓

Continue
```

Đây là toàn bộ quá trình.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Một Blocking Syscall sẽ làm dừng toàn bộ Go Runtime.

Sai.

Runtime chỉ để Machine bị Block và giải phóng Processor để Machine khác tiếp tục thực thi Goroutine.

---

### Hiểu lầm 2

> `entersyscall()` dùng để gọi System Call.

Sai.

System Call do mã chương trình hoặc thư viện thực hiện.

`entersyscall()` chỉ chuẩn bị trạng thái của Runtime trước khi Thread đi vào Kernel.

---

### Hiểu lầm 3

> Sau `exitsyscall()`, Goroutine luôn tiếp tục trên cùng Processor.

Sai.

Runtime chỉ cần một Processor khả dụng.

Không nhất thiết là Processor cũ.

---

### Hiểu lầm 4

> Mọi lời gọi Network đều là Blocking Syscall.

Sai.

Phần lớn Network I/O trong Go sử dụng Netpoll với các cơ chế như `epoll`, `kqueue` hoặc `IOCP`, giúp tránh Block Thread trong thời gian dài.

---

### Hiểu lầm 5

> Machine bị Block thì Scheduler cũng bị Block.

Sai.

Scheduler vẫn tiếp tục hoạt động trên các Machine khác miễn là còn Processor khả dụng.

---

# Những điều cần ghi nhớ

- System Call là cầu nối giữa User Space và Kernel Space.
- Blocking Syscall chỉ Block Machine, không Block toàn bộ Runtime.
- `entersyscall()` lưu trạng thái và giải phóng Processor trước khi Thread vào Kernel.
- `exitsyscall()` đưa Machine trở lại Scheduler sau khi Kernel hoàn thành.
- Processor là tài nguyên quan trọng nhất mà Scheduler cần bảo toàn.
- Cơ chế này giúp Go duy trì Throughput cao ngay cả khi nhiều Goroutine thực hiện I/O.

---

# Tổng kết

Trong phần này, chúng ta đã hiểu cách Go Runtime xử lý một trong những vấn đề khó nhất của lập lịch: **Blocking System Call**.

Thông qua hai hàm `entersyscall()` và `exitsyscall()`, Runtime tách Processor khỏi Machine trước khi Thread đi vào Kernel và gắn lại Processor khi Thread quay trở lại. Nhờ vậy, các Goroutine khác vẫn có thể tiếp tục chạy trong khi một Thread đang chờ hệ điều hành.

Đây là một trong những cơ chế quan trọng giúp Go Runtime đạt được khả năng xử lý đồng thời rất cao mà không cần tạo quá nhiều Thread.

Trong phần tiếp theo, chúng ta sẽ chuyển sang **Phần VI - Performance**, nơi sẽ phân tích chi phí tạo Thread, cơ chế tái sử dụng Machine và các tối ưu hiệu năng mà Go Runtime áp dụng để giảm chi phí quản lý OS Thread.

---

---

# Phần VI - Performance

Trong các phần trước, chúng ta đã nghiên cứu:

- Machine (`struct m`)
- Thread Lifecycle
- Scheduler Interaction
- System Call

Đến thời điểm này.

Chúng ta đã hiểu cách Go Runtime quản lý OS Thread.

Tuy nhiên.

Một câu hỏi quan trọng vẫn còn.

> **Việc tạo một Thread có đắt không?**

Nếu đắt.

Go Runtime đã làm gì để giảm chi phí này?

Đây chính là mục tiêu của phần Performance.

Trong phần này.

Chúng ta sẽ nghiên cứu.

- Chi phí tạo Thread.
- Cơ chế Thread Cache.
- Thread Reuse.
- Các tối ưu của Go Runtime.

Đây cũng là lý do Go có thể duy trì số lượng Thread rất nhỏ ngay cả khi chạy hàng triệu Goroutine.

---

# 12.26 Thread Creation Cost

Trước tiên.

Chúng ta cần hiểu.

```
Tạo Thread

≠

Tạo Goroutine
```

Đây là hai khái niệm hoàn toàn khác nhau.

---

# Tạo Goroutine

Ví dụ.

```go
go worker()
```

Runtime thực hiện.

```text
Allocate struct g

↓

Allocate Initial Stack

↓

Put vào Run Queue
```

Toàn bộ quá trình.

Không cần Kernel.

---

# Tạo Thread

Ngược lại.

Khi Runtime tạo Machine mới.

Runtime phải.

```text
Call Kernel

↓

Create Thread

↓

Allocate Kernel Stack

↓

Initialize TLS

↓

Initialize Registers

↓

Schedule Thread
```

Đây là một quá trình tốn kém hơn rất nhiều.

---

# Vì sao Thread Creation đắt?

Có nhiều nguyên nhân.

---

## Kernel Transition

Runtime phải chuyển từ.

```text
User Mode

↓

Kernel Mode
```

Mỗi lần chuyển.

CPU cần.

- Lưu Register.
- Đổi Privilege Level.
- Đồng bộ Pipeline.

Đây là chi phí cố định.

---

## Kernel Stack

Mỗi Thread đều có.

```
Kernel Stack
```

Được Kernel cấp phát.

Thông thường.

Có kích thước lớn hơn rất nhiều so với Initial Stack của Goroutine.

---

## Thread Local Storage

Thread mới cần.

```
TLS
```

để lưu.

- errno
- Signal
- Runtime Metadata

Việc khởi tạo TLS cũng tiêu tốn thời gian.

---

## Scheduler của Kernel

Ngay sau khi Thread được tạo.

Kernel phải.

```text
Add Thread

↓

OS Scheduler
```

Điều này cũng tạo thêm chi phí.

---

# So sánh

Có thể hình dung.

```text
Create Goroutine

↓

Microseconds hoặc thấp hơn
```

Trong khi.

```text
Create Thread

↓

Đắt hơn nhiều
```

Chi phí chính xác phụ thuộc vào:

- Hệ điều hành.
- CPU.
- Phiên bản Kernel.

Điều quan trọng là.

Runtime luôn cố gắng.

```
Reuse

>

Create
```

---

# 12.27 Thread Cache

Nếu Thread Creation đắt.

Giải pháp tự nhiên là.

```
Cache Thread
```

Đây cũng là chiến lược mà Go Runtime sử dụng.

---

# Ý tưởng

Thay vì.

```text
Destroy Thread

↓

Create Thread
```

Runtime làm.

```text
Park Thread

↓

Wakeup
```

Machine được giữ lại.

Để tái sử dụng.

---

# Thread Pool ngầm

Go Runtime không có.

```
Thread Pool
```

theo nghĩa truyền thống.

Nhưng.

Các Machine đang Park.

Có thể xem như.

```
Thread Cache
```

---

# Minh họa

```text
Running Machines

↓

M1

↓

M2
```

```text
Idle Machines

↓

M3

↓

M4
```

Khi cần.

Runtime đánh thức.

```
M3
```

Thay vì tạo.

```
M5
```

---

# Lợi ích

- Không cần Kernel Call.
- Không cần cấp phát Stack mới.
- Không cần khởi tạo TLS.

Latency thấp hơn đáng kể.

---

# Machine Metadata vẫn còn

Khi Park.

Machine vẫn giữ.

- G0
- MCache
- Thread Metadata

Do đó.

Wakeup rất nhanh.

---

# 12.28 Thread Reuse

Thread Reuse là một trong những tối ưu quan trọng nhất của Go Runtime.

---

# Không tạo rồi hủy liên tục

Giả sử.

Server.

```text
Request

↓

Create Thread

↓

Destroy
```

Nếu có.

```
100000 Requests
```

Chi phí sẽ rất lớn.

---

# Runtime làm gì?

Runtime.

```text
Request

↓

Wakeup Thread

↓

Run

↓

Park Again
```

Một Thread.

Có thể phục vụ.

Hàng triệu Goroutine.

---

# Ví dụ

```text
M1

↓

G1
```

Sau khi.

```
G1 Exit
```

```text
M1

↓

G2
```

Sau đó.

```text
M1

↓

G300
```

...

Machine gần như không bị hủy.

---

# Thread Reuse và Scheduler

Scheduler luôn ưu tiên.

```text
Wake Idle Machine
```

Thay vì.

```
Create New Machine
```

Đây là lý do.

Số Thread trong Go.

Thường ổn định.

---

# Khi nào Runtime tạo Thread mới?

Chỉ khi.

- Không còn Machine rảnh.
- Blocking Syscall.
- Runtime cần thêm Thread.

Đây là.

```
Lazy Creation
```

---

# Thread Reuse và Memory

Việc tái sử dụng Thread.

Giúp giữ lại.

- G0.
- MCache.
- TLS.

Giảm đáng kể Allocation.

---

# 12.29 Performance Optimization

Đây là phần tổng hợp.

Go Runtime sử dụng rất nhiều kỹ thuật để tối ưu Thread.

---

# Tối ưu 1 - Lazy Thread Creation

Runtime không tạo Thread trước.

Chỉ tạo khi.

```
Need More M
```

Điều này giúp.

- Giảm Memory.
- Giảm Kernel Call.

---

# Tối ưu 2 - Thread Reuse

Thay vì.

```
Destroy

↓

Create
```

Runtime.

```
Park

↓

Wakeup
```

---

# Tối ưu 3 - Thread Local Cache

Mỗi Machine có.

```
MCache
```

Giúp.

- Giảm Lock.
- Giảm Cache Miss.
- Giảm Allocation Cost.

---

# Tối ưu 4 - Processor Detach

Khi Blocking Syscall.

Runtime.

```text
Release P

↓

Run Other G
```

CPU không bị lãng phí.

---

# Tối ưu 5 - Spinning Machine

Một số Machine.

```
Searching Runnable G
```

Thay vì Sleep.

Giảm Wakeup Latency.

---

# Tối ưu 6 - Parking Machine

Nếu quá nhiều Machine.

Runtime.

```text
Park
```

Giảm CPU Usage.

---

# Tối ưu 7 - Dynamic Thread Count

Runtime không cố định.

```
10 Threads

hay

100 Threads
```

Số Machine.

Thay đổi theo.

- Workload.
- Blocking.
- Scheduler.

---

# Tối ưu 8 - Giảm Kernel Transition

Runtime cố gắng.

- Context Switch trong User Space.
- Scheduler trong User Space.
- Goroutine Switching không cần Kernel.

Kernel chỉ tham gia khi thật sự cần.

---

# Tối ưu 9 - Kết hợp với Netpoll

Network I/O.

Thông thường.

Không tạo thêm Thread.

Runtime.

```text
Netpoll

↓

Wakeup G
```

Giảm số Machine cần thiết.

---

# Tối ưu 10 - G-M-P

Đây là tối ưu lớn nhất.

Nếu không có.

```
Processor
```

Runtime sẽ rất khó.

- Tách Thread.
- Reuse Thread.
- Work Stealing.

Mô hình G-M-P là nền tảng của toàn bộ Performance.

---

# So sánh

| Không tối ưu | Go Runtime |
|--------------|------------|
| Tạo Thread liên tục | Reuse Machine |
| Block Thread | Release Processor |
| Scheduler trong Kernel | Scheduler trong User Space |
| Thread cố định | Dynamic Thread Count |
| Không có Cache | MCache |

---

# Production

Một Server Go.

Có thể.

```text
500.000 Goroutines

↓

16 Threads
```

Điều này xảy ra nhờ.

- Dynamic Stack.
- Scheduler.
- Thread Reuse.
- Netpoll.
- MCache.

Không phải vì Thread nhẹ.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Tạo Goroutine và tạo Thread có cùng chi phí.

Sai.

Tạo Goroutine chỉ cần Runtime cấp phát `struct g` và Stack ban đầu.

Tạo Thread yêu cầu Kernel tham gia.

---

### Hiểu lầm 2

> Runtime tạo trước rất nhiều Thread để tăng hiệu năng.

Sai.

Go Runtime sử dụng chiến lược Lazy Creation và chỉ tạo thêm Machine khi thật sự cần.

---

### Hiểu lầm 3

> Thread Park giống Thread Destroy.

Sai.

Park giữ nguyên Machine để tái sử dụng.

Destroy giải phóng hoàn toàn Thread.

---

### Hiểu lầm 4

> Thread Reuse chỉ giúp tiết kiệm RAM.

Sai.

Thread Reuse còn giảm:

- Kernel Call.
- TLS Initialization.
- Thread Startup Time.
- Scheduler Overhead.

---

### Hiểu lầm 5

> Performance của Go chủ yếu đến từ Goroutine.

Sai.

Hiệu năng của Go là kết quả của nhiều thành phần phối hợp:

- Goroutine.
- Machine.
- Processor.
- Scheduler.
- MCache.
- Netpoll.
- Garbage Collector.

---

# Những điều cần ghi nhớ

- Tạo Thread đắt hơn rất nhiều so với tạo Goroutine.
- Runtime ưu tiên Park và Reuse Thread.
- `MCache` giúp Machine cấp phát bộ nhớ nhanh hơn.
- Lazy Thread Creation giúp giảm chi phí Kernel.
- G-M-P là nền tảng của toàn bộ tối ưu hiệu năng trong Go Runtime.
- Mục tiêu của Runtime là giảm số Thread nhưng vẫn tận dụng tối đa CPU.

---

# Tổng kết

Trong phần Performance, chúng ta đã hiểu vì sao Go Runtime có thể đạt hiệu năng rất cao dù chỉ sử dụng một số lượng nhỏ OS Thread.

Thông qua các kỹ thuật như **Lazy Thread Creation**, **Thread Reuse**, **MCache**, **Processor Detach**, **Thread Parking** và **Thread Spinning**, Runtime giảm đáng kể chi phí tạo Thread, hạn chế việc gọi Kernel và tận dụng tối đa tài nguyên hệ thống.

Đây là nền tảng để các ứng dụng Go có thể xử lý khối lượng lớn Goroutine với chi phí thấp và độ trễ nhỏ.

Trong phần tiếp theo, chúng ta sẽ chuyển sang **Phần VII - Debugging**, nơi sẽ tìm hiểu cách quan sát và chẩn đoán các vấn đề liên quan đến Machine và OS Thread trong môi trường Production.

---

---

# Phần VII - Debugging

Cho đến thời điểm này.

Chúng ta đã nghiên cứu gần như toàn bộ vòng đời của Machine.

Bao gồm.

- `struct m`
- Thread Lifecycle
- Scheduler Interaction
- System Call
- Performance

Đây đều là kiến thức bên trong Go Runtime.

Tuy nhiên.

Trong môi trường Production.

Điều quan trọng hơn là.

> **Làm thế nào để quan sát Machine đang hoạt động?**

Khi hệ thống gặp vấn đề.

- CPU tăng cao.
- Throughput giảm.
- Latency tăng.
- Hàng nghìn Goroutine bị Block.

Chúng ta cần biết.

- Có bao nhiêu Thread?
- Thread đang làm gì?
- Thread nào đang Block?
- Scheduler có tạo quá nhiều Thread không?

Đây chính là mục tiêu của phần Debugging.

---

# 12.30 Machine Dump

Khác với Goroutine.

Go Runtime hiện nay.

Không cung cấp.

```
Machine Dump
```

dưới dạng Public API.

Ví dụ.

Không tồn tại.

```go
runtime.DumpMachine()
```

hay.

```go
runtime.NumMachine()
```

Điều này khiến nhiều người nghĩ.

```
Machine

↓

Không thể Debug.
```

Điều đó không đúng.

---

# Vì sao Runtime không Public Machine?

Lý do đơn giản.

Machine là.

```
Runtime Internal
```

Lập trình viên.

Không nên phụ thuộc.

Vào số lượng Thread.

Scheduler có quyền.

- Tạo.
- Hủy.
- Park.
- Wakeup.

Machine bất kỳ lúc nào.

---

# Có thể quan sát Machine bằng cách nào?

Có nhiều cách.

Ví dụ.

- Runtime Trace.
- pprof.
- Delve.
- GDB.
- Source Instrumentation.

---

# Góc nhìn từ Goroutine Dump

Giả sử.

```
50000 Goroutines
```

Nhìn vào.

```
Goroutine Dump
```

Chúng ta có thể suy luận.

- Có bao nhiêu Thread đang chạy.
- Bao nhiêu Goroutine đang Running.
- Bao nhiêu Goroutine Waiting.

Mặc dù.

Không thấy trực tiếp.

```
struct m
```

---

# Delve

Debugger.

```
dlv
```

cho phép.

Đọc trực tiếp.

```
runtime.allm
```

Đây là danh sách.

Tất cả Machine.

Ví dụ.

```text
runtime.allm

↓

M0

↓

M1

↓

M2

↓

M3
```

Từ đây.

Có thể đọc.

- curg
- p
- g0
- locks

---

# GDB

Nếu Debug.

Binary Go.

Có thể đọc.

```
runtime.allm
```

hoặc.

```
runtime.allgs
```

Đây là cách.

Go Team.

Debug Runtime.

---

# Runtime Source Instrumentation

Một số công ty.

Fork Go Runtime.

Thêm Log.

Ví dụ.

```text
Create M

Destroy M

Park M

Wakeup M
```

Giúp theo dõi.

Scheduler.

Trong Production.

---

# Dấu hiệu bất thường

Machine.

Thường ổn định.

Nếu.

```
Thread Count

↑↑↑
```

Có thể.

- Blocking Syscall.
- CGO.
- Thread Leak.
- Scheduler Stress.

---

# Thread Count

Một Metric rất quan trọng.

Ví dụ.

Linux.

```text
ps -T

top -H

htop
```

Hoặc.

```text
/proc/PID/task
```

Nếu.

```
16 CPUs

↓

800 Threads
```

Đây thường là dấu hiệu.

Có vấn đề.

---

# 12.31 Thread Leak

Chúng ta đã học.

```
Goroutine Leak
```

ở Chapter trước.

Bây giờ.

Chúng ta nghiên cứu.

```
Thread Leak
```

---

# Thread Leak là gì?

Thread Leak.

Là hiện tượng.

```
Thread

↓

Tăng liên tục

↓

Không giảm.
```

---

# Thread Leak khác Goroutine Leak

Nhiều người nhầm.

Hai khái niệm này.

---

## Goroutine Leak

```text
G

↓

Không Exit
```

---

## Thread Leak

```text
Machine

↓

Không được Reuse

↓

Thread tăng.
```

---

# Thread Leak hiếm hơn

Go Runtime.

Quản lý Thread.

Rất tốt.

Do đó.

Thread Leak.

Ít gặp hơn.

Goroutine Leak.

---

# Nguyên nhân

Có nhiều nguyên nhân.

---

## Blocking Syscall quá nhiều

Ví dụ.

```go
syscall.Read(...)
```

Nếu.

```
1000 Threads

↓

Read()

↓

Waiting
```

Scheduler buộc.

Tạo thêm Machine.

---

## CGO

Đây là nguyên nhân phổ biến.

CGO.

Có thể tạo.

OS Thread.

Ngoài sự kiểm soát.

Của Scheduler.

---

## LockOSThread

Ví dụ.

```go
runtime.LockOSThread()
```

Nếu.

Không.

```go
runtime.UnlockOSThread()
```

Thread.

Có thể.

Không được Reuse.

---

## External Library

Một số Library.

Tạo Thread.

Qua.

CGO.

Go Runtime.

Không thể quản lý.

---

# Dấu hiệu

Ví dụ.

```
Yesterday

↓

20 Threads
```

```text
Today

↓

300 Threads
```

Trong khi.

CPU thấp.

Đây là dấu hiệu.

Nên điều tra.

---

# Cách phát hiện

Có thể.

Theo dõi.

- Number of Threads.
- Runtime Trace.
- Goroutine Dump.
- OS Metrics.

---

# Thread Leak và Goroutine Leak

Hai hiện tượng.

Có thể xảy ra cùng lúc.

Ví dụ.

```text
G Leak

↓

Waiting Read

↓

Thread Block

↓

Scheduler

↓

Create More Thread
```

Cuối cùng.

Cả.

```
G

và

Thread
```

đều tăng.

---

# Production Checklist

Nếu Thread tăng.

Hãy kiểm tra.

- Có CGO không?
- Có Blocking Syscall không?
- Có LockOSThread không?
- Có Thread ngoài Runtime không?

---

# 12.32 Runtime Trace dưới góc nhìn Machine

Trong Chapter 11.

Chúng ta đã học.

```
runtime/trace
```

Dưới góc nhìn.

```
Goroutine
```

Bây giờ.

Chúng ta nhìn.

Từ phía.

```
Machine
```

---

# Runtime Trace ghi lại gì?

Trace.

Không chỉ ghi.

```
Go Routine
```

Mà còn.

- Machine.
- Processor.
- Scheduler.
- Syscall.
- GC.

---

# Machine Events

Ví dụ.

Runtime Trace.

Có thể hiển thị.

```text
Thread Created
```

```text
Thread Park
```

```text
Thread Unpark
```

```text
Syscall Begin
```

```text
Syscall End
```

---

# Scheduler Timeline

Có thể hình dung.

```text
M1

↓

Running

↓

Syscall

↓

Park

↓

Wakeup
```

Trace.

Cho phép.

Quan sát.

Theo thời gian.

---

# Thread Creation Spike

Giả sử.

Bạn thấy.

```text
Thread

20

↓

50

↓

100

↓

200
```

Trong Trace.

Đây thường là.

Dấu hiệu.

- Blocking.
- CGO.
- Scheduler Pressure.

---

# Syscall Analysis

Trace.

Hiển thị.

```text
Enter Syscall

↓

Wait

↓

Exit Syscall
```

Có thể biết.

Thread.

Đã Block.

Bao lâu.

---

# Processor Handoff

Một điểm rất thú vị.

Trace.

Cho thấy.

```text
P1

↓

M1
```

Sau đó.

```text
P1

↓

M3
```

Đây chính là.

```
Processor Detach
```

mà chúng ta học.

Ở phần trước.

---

# Thread Parking

Có thể thấy.

```text
Running

↓

Idle

↓

Park

↓

Wake
```

Nếu Thread.

Park.

Quá thường xuyên.

Có thể.

Workload.

Không ổn định.

---

# Thread Spinning

Trace.

Cũng cho thấy.

Machine.

```text
Searching Work
```

Nếu.

Quá nhiều.

Spinning Thread.

CPU.

Có thể.

Tăng.

---

# Kết hợp với pprof

Runtime Trace.

Cho biết.

```
Thread

↓

Làm gì.
```

pprof.

Cho biết.

```
CPU

↓

Ở đâu.
```

Hai công cụ.

Bổ sung.

Cho nhau.

---

# Khi nào nên dùng Runtime Trace?

- Latency bất thường.
- Scheduler hoạt động không như mong muốn.
- Thread tăng nhanh.
- Blocking Syscall kéo dài.
- CPU thấp nhưng Throughput giảm.
- Nghi ngờ Thread Leak hoặc quá nhiều Spinning Machine.

---

# Quy trình Debug Production

Một quy trình phổ biến.

```text
CPU tăng

↓

pprof

↓

Thread Count

↓

Runtime Trace

↓

Goroutine Dump

↓

Scheduler Analysis
```

Không nên.

Đi thẳng.

Vào Source Code.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Go Runtime có API công khai để Dump toàn bộ Machine.

Sai.

Machine là cấu trúc nội bộ.

Muốn quan sát cần dùng Runtime Trace, Debugger hoặc đọc dữ liệu Runtime.

---

### Hiểu lầm 2

> Thread Leak và Goroutine Leak là một.

Sai.

Goroutine Leak liên quan đến `struct g`.

Thread Leak liên quan đến `struct m` và OS Thread.

Hai hiện tượng có thể độc lập hoặc xảy ra cùng nhau.

---

### Hiểu lầm 3

> Thread Count càng nhiều càng tốt.

Sai.

Quá nhiều Thread làm tăng:

- Context Switching.
- Scheduler Overhead.
- Memory Usage.

---

### Hiểu lầm 4

> Runtime Trace chỉ dùng để Debug Goroutine.

Sai.

Runtime Trace ghi lại cả:

- Machine.
- Processor.
- Syscall.
- Scheduler.
- Garbage Collector.

---

### Hiểu lầm 5

> Chỉ cần nhìn Goroutine Dump là đủ.

Sai.

Goroutine Dump cho biết Goroutine đang ở đâu.

Runtime Trace cho biết Scheduler và Machine đã tương tác như thế nào theo thời gian.

---

# Những điều cần ghi nhớ

- Go Runtime không cung cấp API công khai để Dump Machine.
- Có thể quan sát Machine thông qua Runtime Trace, Delve hoặc GDB.
- Thread Leak hiếm hơn Goroutine Leak nhưng ảnh hưởng nghiêm trọng đến hiệu năng.
- Runtime Trace là công cụ quan trọng để phân tích Thread Lifecycle và Scheduler.
- Luôn kết hợp Runtime Trace, pprof và Goroutine Dump khi Debug Production.

---

# Tổng kết

Trong phần Debugging, chúng ta đã chuyển từ góc nhìn **Runtime Internals** sang góc nhìn **vận hành Production**.

Chúng ta đã thấy rằng mặc dù `struct m` là cấu trúc nội bộ của Go Runtime, vẫn có nhiều cách để quan sát hoạt động của Machine thông qua Runtime Trace, Debugger và các chỉ số của hệ điều hành.

Bên cạnh đó, chúng ta cũng phân biệt rõ **Thread Leak** và **Goroutine Leak**, hiểu được nguyên nhân khiến số lượng Thread tăng bất thường và biết cách sử dụng Runtime Trace để phân tích Scheduler dưới góc nhìn của Machine.

Đây là nền tảng quan trọng trước khi bước sang phần cuối cùng của chương, nơi chúng ta sẽ tổng hợp các **Best Practices**, **Common Mistakes** và **Interview Questions** dành cho Machine (`struct m`) và OS Thread trong Go Runtime.

---

---

# 12.33 Best Practices

Trong suốt Chapter 12, chúng ta đã nghiên cứu Machine (`struct m`) từ rất nhiều góc độ.

Bao gồm.

- Runtime Internals.
- Thread Lifecycle.
- Scheduler Interaction.
- System Call.
- Performance.
- Debugging.

Đây đều là kiến thức ở mức Runtime.

Tuy nhiên.

Trong công việc hằng ngày.

Chúng ta hiếm khi thao tác trực tiếp với.

```
struct m
```

Thay vào đó.

Chúng ta viết.

```go
go worker()
```

hay.

```go
http.ListenAndServe(...)
```

Mặc dù không làm việc trực tiếp với Machine.

Nhưng.

Nếu hiểu Machine.

Chúng ta sẽ viết được những chương trình:

- Hiệu năng hơn.
- Ít Thread hơn.
- Ít Context Switching hơn.
- Ít Blocking hơn.

Đây chính là mục tiêu của phần Engineering.

---

# Best Practice 1 - Đừng tối ưu số lượng Thread, hãy tối ưu Goroutine

Một sai lầm phổ biến.

Là cố gắng.

```
Giảm Thread.
```

Thực tế.

Thread được Runtime quản lý.

Lập trình viên nên tập trung.

- Goroutine.
- Blocking.
- Scheduler Friendly Code.

Runtime sẽ tự quyết định số Machine.

---

# Best Practice 2 - Tránh Blocking System Call kéo dài

Ví dụ.

```go
syscall.Read(...)
```

Nếu Thread Block.

Runtime phải.

- Tách Processor.
- Có thể tạo Machine mới.

Nếu quá nhiều Blocking Syscall.

Thread Count sẽ tăng.

Ưu tiên.

- Context.
- Timeout.
- Netpoll.
- Async I/O.

---

# Best Practice 3 - Sử dụng thư viện chuẩn thay vì gọi `syscall` trực tiếp

Ví dụ.

Thay vì.

```go
syscall.Read(...)
```

Hãy dùng.

```go
os.File

net.Conn

http.Client
```

Các thư viện chuẩn.

Đã tích hợp.

Scheduler.

Netpoll.

Context.

Timeout.

---

# Best Practice 4 - Hạn chế sử dụng `runtime.LockOSThread()`

Đây là API rất đặc biệt.

Ví dụ.

```go
runtime.LockOSThread()
```

Sau khi Lock.

Goroutine.

Không thể chuyển sang Machine khác.

Scheduler mất đi tính linh hoạt.

Chỉ nên sử dụng khi thật sự cần.

Ví dụ.

- OpenGL.
- GUI Toolkit.
- Một số thư viện CGO.

---

# Best Practice 5 - Luôn gọi `runtime.UnlockOSThread()`

Nếu đã.

```go
LockOSThread()
```

Hãy luôn.

```go
UnlockOSThread()
```

sau khi hoàn thành.

Nếu không.

Thread.

Có thể bị giữ.

Không được Reuse.

---

# Best Practice 6 - Ưu tiên Netpoll cho Network I/O

Không nên.

Tự triển khai.

Blocking Socket.

Go Runtime.

Đã tối ưu.

```
Netpoll
```

Sử dụng.

- epoll
- kqueue
- IOCP

Điều này giúp.

Giảm số Machine.

---

# Best Practice 7 - Giới hạn Blocking Operation

Ví dụ.

Database.

Redis.

RPC.

HTTP.

Nên luôn.

- Timeout.
- Retry.
- Circuit Breaker.

Nếu không.

Machine.

Có thể Block.

Trong thời gian dài.

---

# Best Practice 8 - Theo dõi Thread Count trong Production

Ngoài.

```go
runtime.NumGoroutine()
```

Nên theo dõi.

```
OS Thread Count
```

Ví dụ.

Linux.

```text
/proc/PID/task
```

Hoặc.

```
top -H
```

Thread tăng bất thường.

Là dấu hiệu rất quan trọng.

---

# Best Practice 9 - Hiểu rằng Goroutine rẻ, Thread thì không

Đừng nhầm.

```
100000 Goroutines
```

với.

```
100000 Threads
```

Runtime.

Được thiết kế.

Để tạo rất nhiều Goroutine.

Nhưng.

Không phải.

Rất nhiều Thread.

---

# Best Practice 10 - Hạn chế CGO nếu không cần thiết

CGO.

Có thể.

- Tạo thêm Thread.
- Giảm hiệu quả Scheduler.
- Tăng Context Switching.

Nếu có thể.

Ưu tiên.

Pure Go.

---

# Best Practice 11 - Thiết kế ứng dụng theo hướng Event-driven

Thay vì.

Blocking.

Hãy.

- Channel.
- Context.
- Callback.
- Netpoll.

Điều này giúp.

Machine.

Ít Block hơn.

---

# Best Practice 12 - Luôn sử dụng Context với I/O

Ví dụ.

```go
ctx, cancel := context.WithTimeout(...)
```

Nếu không.

Một Thread.

Có thể.

Block.

Rất lâu.

---

# Best Practice 13 - Hiểu chi phí của Thread Creation

Không nên.

Giả định.

```
Runtime tạo Thread

↓

Miễn phí.
```

Thread Creation.

Luôn.

Đắt hơn.

Tạo Goroutine.

---

# Best Practice 14 - Đừng tự xây Thread Pool

Một số lập trình viên.

Đến từ Java.

Có xu hướng.

```
Thread Pool.
```

Trong Go.

Thông thường.

Scheduler.

Đã làm việc này.

Không nên.

Quản lý Thread.

Bằng tay.

---

# Best Practice 15 - Worker Pool nên quản lý Goroutine, không phải Thread

Worker Pool.

Nên.

```text
Worker

↓

Goroutine
```

Không phải.

```
Worker

↓

OS Thread
```

Đây là điểm khác biệt rất lớn.

Giữa Go.

Và Java.

---

# Best Practice 16 - Theo dõi Runtime Trace khi Latency tăng

Nếu.

Latency.

Tăng.

Đừng chỉ nhìn.

CPU.

Hãy.

Xem.

```
Runtime Trace
```

Có thể.

Machine.

Đang.

- Park.
- Spinning.
- Blocking.

---

# Best Practice 17 - Giảm Blocking trong Critical Path

Ví dụ.

Một HTTP Request.

Không nên.

Thực hiện.

Nhiều.

Blocking I/O.

Liên tiếp.

Điều này.

Làm tăng.

Thread Occupancy.

---

# Best Practice 18 - Hiểu rằng Scheduler tối ưu hơn chúng ta nghĩ

Không nên.

Can thiệp.

Vào.

Scheduler.

Nếu.

Không có Benchmark.

Runtime.

Đã tối ưu.

Qua.

Rất nhiều năm.

---

# Best Practice 19 - Đọc Runtime Source khi tối ưu hiệu năng

Nếu.

Làm việc.

Ở mức.

High Performance.

Nên đọc.

```
runtime/proc.go
```

và.

```
runtime/runtime2.go
```

Điều này.

Giúp hiểu.

Scheduler.

Tốt hơn.

---

# Best Practice 20 - Thiết kế theo G-M-P, không theo Thread

Đây là nguyên tắc quan trọng nhất.

Khi thiết kế.

Một hệ thống Go.

Đừng nghĩ.

```
Task

↓

Thread
```

Hãy nghĩ.

```text
Task

↓

Goroutine

↓

Scheduler

↓

Machine

↓

CPU
```

Đây chính là cách.

Go Runtime.

Hoạt động.

---

# Bonus Best Practices

Ngoài 20 nguyên tắc trên.

Một số kinh nghiệm thực tế.

---

## Không thay đổi GOMAXPROCS nếu chưa Benchmark

Runtime.

Đã chọn.

Giá trị.

Phù hợp.

Trong đa số trường hợp.

---

## Đừng dùng LockOSThread để tối ưu

LockOSThread.

Không làm.

Chương trình.

Nhanh hơn.

Ngược lại.

Có thể.

Giảm.

Khả năng lập lịch.

---

## Luôn đo Thread Count cùng Goroutine Count

Hai Metric.

Nên.

Đi cùng nhau.

Ví dụ.

```text
50.000 Goroutines

16 Threads
```

Hoàn toàn.

Bình thường.

Nhưng.

```text
50.000 Goroutines

800 Threads
```

Là tín hiệu.

Cần.

Điều tra.

---

## Theo dõi Blocking Syscall

Nếu.

Thread.

Tăng nhanh.

Hãy.

Kiểm tra.

- Database.
- File I/O.
- DNS.
- CGO.

---

## Luôn ưu tiên Runtime Trace trước khi tối ưu Scheduler

Không nên.

Đoán.

Hãy.

Đo.

---

# Checklist Production

Trước khi triển khai một dịch vụ Go.

Hãy tự kiểm tra.

- [ ] Không lạm dụng `runtime.LockOSThread()`.
- [ ] Có Timeout cho mọi I/O.
- [ ] Hạn chế Blocking Syscall kéo dài.
- [ ] Theo dõi Thread Count và Goroutine Count.
- [ ] Sử dụng Runtime Trace khi phân tích Scheduler.
- [ ] Ưu tiên thư viện chuẩn thay vì `syscall` trực tiếp.
- [ ] Không tự xây dựng Thread Pool.
- [ ] Worker Pool sử dụng Goroutine.
- [ ] Hạn chế phụ thuộc vào CGO nếu không thật sự cần.
- [ ] Thiết kế hệ thống theo mô hình G-M-P thay vì tư duy "một tác vụ - một Thread".

---

# Tổng kết

Machine (`struct m`) là một thành phần nội bộ của Go Runtime và phần lớn lập trình viên sẽ không bao giờ thao tác trực tiếp với nó.

Tuy nhiên, việc hiểu Machine giúp chúng ta đưa ra nhiều quyết định kỹ thuật đúng đắn hơn:

- Viết mã thân thiện với Scheduler.
- Hạn chế Blocking System Call.
- Giảm số lượng OS Thread.
- Tận dụng tối đa cơ chế Thread Reuse.
- Thiết kế hệ thống có khả năng mở rộng với chi phí tài nguyên thấp.

Một Senior Go Engineer không cần quản lý Thread bằng tay.

Điều quan trọng hơn là **hiểu Runtime quản lý Thread như thế nào**, từ đó viết những chương trình giúp Scheduler hoạt động hiệu quả nhất.

---

---

# 12.34 Common Mistakes

Trong Chapter 12, chúng ta đã tìm hiểu rất sâu về **Machine (`struct m`)** và cách Go Runtime quản lý **OS Thread**.

Đây là một chủ đề mà rất nhiều lập trình viên, kể cả những người đã làm Go nhiều năm, vẫn hiểu chưa chính xác.

Nguyên nhân là vì.

Machine.

Là một cấu trúc hoàn toàn nội bộ.

Go Runtime không Public.

Do đó.

Đa số lập trình viên chỉ nhìn thấy.

```go
go worker()
```

mà không nhìn thấy.

```
Machine

↓

Thread

↓

Scheduler
```

Trong phần này.

Chúng ta sẽ tổng hợp những hiểu lầm phổ biến nhất về Machine.

Đây cũng là những lỗi rất thường gặp trong phỏng vấn Senior Go.

---

# Sai lầm 1 - Machine chính là OS Thread

Đây là hiểu lầm phổ biến nhất.

Nhiều người trả lời.

```
Machine

=

OS Thread
```

Không chính xác.

Thực tế.

```
Machine

↓

Runtime Object
```

```text
OS Thread

↓

Kernel Object
```

Machine chỉ đại diện cho Thread bên trong Runtime.

---

# Sai lầm 2 - Runtime tạo một Machine cho mỗi Goroutine

Ví dụ.

```go
for i := 0; i < 100000; i++ {

    go worker()

}
```

Không có nghĩa.

Runtime tạo.

```
100000 Machines
```

Một Machine.

Có thể.

Chạy.

Hàng triệu Goroutine.

Trong suốt vòng đời.

---

# Sai lầm 3 - Mỗi Goroutine luôn chạy trên cùng một Machine

Sai.

Ví dụ.

```text
Time 1

G5

↓

M1
```

Sau Context Switch.

```text
Time 2

G5

↓

M4
```

Điều này.

Hoàn toàn bình thường.

---

# Sai lầm 4 - Machine tự chọn Goroutine để chạy

Không đúng.

Machine.

Không có Run Queue.

Machine.

Không Scheduling.

Machine.

Chỉ thực thi.

Goroutine.

Mà Scheduler giao.

---

# Sai lầm 5 - Machine quản lý Goroutine

Sai.

Machine.

Không quản lý.

Goroutine.

Processor.

Mới quản lý.

Run Queue.

Đây là lý do.

Có.

```
P
```

trong.

G-M-P.

---

# Sai lầm 6 - Thread Creation rất rẻ

Một số người nghĩ.

```
Thread

↓

Giống Goroutine
```

Sai.

Thread Creation.

Phải.

- Kernel Call.
- Allocate Kernel Stack.
- TLS.
- Scheduler.

Chi phí.

Cao hơn.

Rất nhiều.

---

# Sai lầm 7 - Runtime tạo sẵn rất nhiều Thread

Sai.

Runtime.

Áp dụng.

```
Lazy Creation
```

Chỉ tạo.

Machine.

Khi cần.

---

# Sai lầm 8 - Runtime Destroy Thread ngay khi Idle

Không đúng.

Thông thường.

Runtime.

```text
Park

↓

Reuse
```

Không.

Destroy.

---

# Sai lầm 9 - Parking và Destroy là một

Sai.

Parking.

```
Thread còn tồn tại.
```

Destroy.

```
Thread bị Kernel thu hồi.
```

Đây là hai trạng thái hoàn toàn khác nhau.

---

# Sai lầm 10 - Blocking Syscall làm dừng toàn bộ chương trình

Đây là một hiểu lầm rất phổ biến.

Runtime.

Thực hiện.

```text
entersyscall()

↓

Release P

↓

Run G khác
```

Chỉ.

Machine.

Bị Block.

Không phải.

Scheduler.

---

# Sai lầm 11 - entersyscall() gọi System Call

Sai.

System Call.

Do chương trình.

Thực hiện.

```
entersyscall()
```

Chỉ.

Thông báo.

Cho Runtime.

Chuẩn bị.

Scheduler.

---

# Sai lầm 12 - exitsyscall() luôn chạy lại Goroutine cũ

Không đúng.

Sau.

```
Kernel Return
```

Scheduler.

Có thể.

Chạy.

Bất kỳ Goroutine.

Nào.

---

# Sai lầm 13 - Mọi Network I/O đều Blocking

Sai.

Go Runtime.

Ưu tiên.

```
Netpoll
```

Sử dụng.

- epoll.
- kqueue.
- IOCP.

Đa số.

Socket.

Không Block Thread.

---

# Sai lầm 14 - LockOSThread giúp tăng hiệu năng

Sai.

`runtime.LockOSThread()`.

Làm giảm.

Khả năng.

Scheduling.

Chỉ nên.

Dùng.

Khi.

Bắt buộc.

---

# Sai lầm 15 - Quên UnlockOSThread không ảnh hưởng

Sai.

Nếu.

Không.

```
UnlockOSThread()
```

Thread.

Có thể.

Không được.

Reuse.

---

# Sai lầm 16 - Machine luôn có Processor

Không đúng.

Machine.

Có thể.

```text
Running

↓

Release P

↓

Waiting
```

Machine.

Không có Processor.

Không chạy được.

Goroutine.

---

# Sai lầm 17 - Machine luôn có Current Goroutine

Sai.

```go
m.curg
```

Có thể.

```go
nil
```

Ví dụ.

Machine.

Đang Idle.

---

# Sai lầm 18 - G0 là Goroutine đầu tiên của chương trình

Không đúng.

G0.

Không phải.

Main Goroutine.

G0.

Là.

```
Scheduler Stack.
```

---

# Sai lầm 19 - G0 có thể chạy User Code

Sai.

Không bao giờ.

Runtime.

Chỉ sử dụng.

G0.

Cho.

- Scheduler.
- GC.
- Runtime Internal.

---

# Sai lầm 20 - Mỗi Machine có nhiều G0

Sai.

Quan hệ.

Luôn là.

```text
1 Machine

↓

1 G0
```

---

# Sai lầm 21 - MCache là Global Cache

Không đúng.

Mỗi Machine.

Có.

```
MCache

riêng.
```

Điều này.

Giúp.

Giảm Lock.

---

# Sai lầm 22 - Thread Leak giống Goroutine Leak

Sai.

Thread Leak.

Liên quan.

```
Machine
```

Goroutine Leak.

Liên quan.

```
struct g
```

Hai vấn đề.

Khác nhau.

---

# Sai lầm 23 - Thread Count càng nhiều càng tốt

Không đúng.

Thread.

Quá nhiều.

↓

Context Switch.

↓

Memory.

↓

Scheduler Overhead.

Tăng.

---

# Sai lầm 24 - Runtime Trace chỉ dùng để xem Goroutine

Sai.

Runtime Trace.

Hiển thị.

- Thread.
- Machine.
- Processor.
- Syscall.
- Scheduler.
- GC.

---

# Sai lầm 25 - Hiểu Machine là đã hiểu Scheduler

Không đúng.

Machine.

Chỉ là.

Một thành phần.

Để hiểu.

Scheduler.

Cần.

Hiểu.

- Processor.
- Run Queue.
- Work Stealing.
- GMP.

---

# Bonus Mistakes

Một số hiểu lầm.

Ít gặp hơn.

Nhưng.

Rất hay xuất hiện.

Trong phỏng vấn.

---

## Machine được Garbage Collector quản lý

Sai.

Machine.

Được.

Runtime.

Quản lý.

Không phải.

GC.

---

## Runtime luôn tạo đúng số Thread bằng số CPU

Sai.

Số Machine.

Không phụ thuộc.

Trực tiếp.

Vào.

CPU.

---

## Một Thread chỉ chạy một Goroutine trong suốt vòng đời

Sai.

Một Machine.

Có thể.

Chạy.

Hàng triệu Goroutine.

---

## Blocking Database Query luôn tạo Thread mới

Không đúng.

Runtime.

Chỉ tạo.

Machine mới.

Nếu.

Thật sự cần.

---

## Thread luôn bị Destroy sau Program Exit

Không chính xác.

Trong quá trình Shutdown.

Runtime.

Sẽ.

Cleanup.

Theo.

Thứ tự.

Không phải.

Thread nào.

Cũng bị.

Destroy.

Ngay lập tức.

---

# Checklist tự đánh giá

Sau khi học xong Chapter 12.

Bạn hãy thử tự trả lời.

- Machine có phải là OS Thread không?
- Vì sao Runtime cần `struct m`?
- Machine khác Processor ở điểm nào?
- `curg` dùng để làm gì?
- `g0` là gì?
- Vì sao mỗi Machine cần `MCache`?
- Thread được tạo khi nào?
- Thread được Park khi nào?
- Runtime xử lý Blocking Syscall như thế nào?
- Vì sao `entersyscall()` phải Release Processor?
- Sau `exitsyscall()`, Goroutine có luôn quay về Machine cũ không?
- Khi nào Runtime tạo thêm Machine?
- Thread Leak khác Goroutine Leak ở điểm nào?
- Runtime Trace giúp quan sát Machine như thế nào?
- Vì sao Go Runtime ưu tiên Reuse Thread?

Nếu bạn có thể trả lời phần lớn các câu hỏi này mà không cần xem tài liệu, bạn đã hiểu khá vững về **Machine (`struct m`)** và cách Go Runtime quản lý OS Thread.

---

---

# 12.35 Interview Questions

Sau khi hoàn thành Chapter 12, chúng ta đã hiểu gần như toàn bộ cơ chế quản lý **Machine (`struct m`)** trong Go Runtime.

Trong phần này, chúng ta sẽ tổng hợp khoảng **50 câu hỏi phỏng vấn** từ mức Middle đến Senior/Staff.

Các câu hỏi được chia thành nhiều nhóm:

- Foundation
- Runtime Internals
- Thread Lifecycle
- Scheduler
- Performance
- Debugging
- Source Code
- System Design

Mục tiêu không chỉ là ghi nhớ câu trả lời mà còn giúp bạn xây dựng tư duy phân tích Go Runtime dưới góc nhìn của Scheduler.

---

# Phần I - Foundation

## Câu 1

Machine (`struct m`) là gì?

---

## Câu 2

Machine khác OS Thread ở điểm nào?

---

## Câu 3

Tại sao Go Runtime cần `struct m` thay vì làm việc trực tiếp với OS Thread?

---

## Câu 4

Quan hệ giữa Machine và OS Thread là gì?

---

## Câu 5

Một Machine có thể chạy bao nhiêu Goroutine?

---

## Câu 6

Một Goroutine có thể chạy trên bao nhiêu Machine?

---

## Câu 7

Machine có sở hữu Goroutine không?

---

## Câu 8

CPU chạy Goroutine hay chạy Machine?

---

## Câu 9

Machine có được Kernel quản lý không?

---

## Câu 10

Machine đóng vai trò gì trong kiến trúc G-M-P?

---

# Phần II - Runtime Internals

## Câu 11

`struct m` chứa những thông tin gì?

---

## Câu 12

`curg` dùng để làm gì?

---

## Câu 13

Khi nào `m.curg` thay đổi?

---

## Câu 14

`g0` là gì?

---

## Câu 15

Vì sao mỗi Machine cần một `g0`?

---

## Câu 16

`g0` có chạy User Code không?

---

## Câu 17

`mcache` là gì?

---

## Câu 18

Vì sao `mcache` giúp tăng hiệu năng?

---

## Câu 19

`locks` trong `struct m` dùng để làm gì?

---

## Câu 20

Machine có thể tồn tại mà không có Processor không?

---

# Phần III - Thread Lifecycle

## Câu 21

Machine được tạo khi nào?

---

## Câu 22

Machine đầu tiên của chương trình là gì?

---

## Câu 23

Vai trò của `M0` là gì?

---

## Câu 24

Go Runtime tạo OS Thread bằng cách nào?

---

## Câu 25

`newosproc()` làm gì?

---

## Câu 26

`mstart()` là gì?

---

## Câu 27

Machine bắt đầu chạy từ đâu?

---

## Câu 28

Khi nào Runtime Park Thread?

---

## Câu 29

Khi nào Runtime Destroy Thread?

---

## Câu 30

Vì sao Runtime ưu tiên Park thay vì Destroy?

---

# Phần IV - Scheduler Interaction

## Câu 31

Làm thế nào Scheduler gắn một Goroutine vào Machine?

---

## Câu 32

Điều gì xảy ra khi Runtime Detach Goroutine khỏi Machine?

---

## Câu 33

Execution Context được lưu ở đâu khi Context Switch?

---

## Câu 34

Machine có tự chọn Goroutine để chạy không?

---

## Câu 35

Machine lấy Goroutine từ đâu?

---

## Câu 36

Điều gì xảy ra nếu Machine bị Block bởi System Call?

---

## Câu 37

Thread Parking là gì?

---

## Câu 38

Thread Spinning là gì?

---

## Câu 39

Vì sao Runtime cần Spinning Thread?

---

## Câu 40

Spinning Thread khác Parked Thread như thế nào?

---

# Phần V - System Call

## Câu 41

Blocking Syscall là gì?

---

## Câu 42

`entersyscall()` có nhiệm vụ gì?

---

## Câu 43

Vì sao `entersyscall()` phải Release Processor?

---

## Câu 44

`exitsyscall()` làm gì?

---

## Câu 45

Sau `exitsyscall()`, Goroutine có luôn chạy trên Machine cũ không?

---

# Phần VI - Performance & Debugging

## Câu 46

Vì sao Thread Creation đắt hơn Goroutine Creation?

---

## Câu 47

Thread Reuse mang lại lợi ích gì?

---

## Câu 48

Thread Leak là gì?

---

## Câu 49

Runtime Trace giúp Debug Machine như thế nào?

---

## Câu 50

Nếu Production Server có:

- 32 CPUs
- 40.000 Goroutines
- 1.500 Threads

Bạn sẽ bắt đầu điều tra từ đâu?

---

# Bonus Questions (Senior)

Các câu hỏi dưới đây thường xuất hiện trong các buổi phỏng vấn Senior, Staff hoặc khi trao đổi về Go Runtime.

---

## Bonus 1

Giải thích toàn bộ vòng đời của một Machine từ khi được tạo đến khi bị hủy.

---

## Bonus 2

Tại sao Go Runtime cần cả `struct g`, `struct m` và `struct p`?

---

## Bonus 3

Nếu bỏ `struct m` và làm việc trực tiếp với OS Thread thì Runtime sẽ gặp những vấn đề gì?

---

## Bonus 4

Vì sao `g0` không chạy User Code?

---

## Bonus 5

Nếu Machine không có `g0`, Scheduler có thể hoạt động không?

---

## Bonus 6

Giải thích mối quan hệ giữa `curg`, `g0` và Context Switching.

---

## Bonus 7

Điều gì xảy ra nếu Runtime không Release Processor khi `entersyscall()`?

---

## Bonus 8

Tại sao Thread Parking giúp tăng hiệu năng?

---

## Bonus 9

Runtime quyết định khi nào nên tạo thêm Machine?

---

## Bonus 10

Tại sao Runtime không giữ tất cả Machine ở trạng thái Spinning?

---

## Bonus 11

Machine có ảnh hưởng gì đến Garbage Collector?

---

## Bonus 12

Giải thích mối quan hệ giữa `mcache` và Memory Allocator.

---

## Bonus 13

Vì sao Thread Count của Go thường nhỏ hơn rất nhiều so với Goroutine Count?

---

## Bonus 14

Trong trường hợp nào `runtime.LockOSThread()` là cần thiết?

---

## Bonus 15

Nếu quên gọi `runtime.UnlockOSThread()`, hệ thống có thể gặp vấn đề gì?

---

## Bonus 16

Machine có thể chạy mà không có Processor không?

---

## Bonus 17

Processor có thể tồn tại mà không có Machine không?

---

## Bonus 18

Machine và Scheduler phối hợp như thế nào để giảm Context Switching?

---

## Bonus 19

Nếu bạn thiết kế lại `struct m`, bạn sẽ thêm hoặc bỏ trường dữ liệu nào? Vì sao?

---

## Bonus 20

Theo bạn, đâu là tối ưu quan trọng nhất liên quan đến Machine trong Go Runtime?

---

# Coding Interview

Ngoài lý thuyết, nhiều công ty còn yêu cầu ứng viên phân tích hành vi của Runtime thông qua các đoạn mã.

---

## Bài 1

```go
for i := 0; i < 100000; i++ {
    go worker()
}
```

**Câu hỏi**

Runtime có tạo 100.000 Machine không?

Giải thích chi tiết.

---

## Bài 2

```go
runtime.LockOSThread()

go worker()
```

**Câu hỏi**

Điều gì xảy ra với Goroutine hiện tại?

Scheduler bị ảnh hưởng như thế nào?

---

## Bài 3

```go
syscall.Read(fd, buf)
```

**Câu hỏi**

Machine sẽ đi qua những trạng thái nào?

Vai trò của `entersyscall()` và `exitsyscall()` là gì?

---

## Bài 4

```go
time.Sleep(time.Minute)
```

**Câu hỏi**

Machine có bị Block không?

Thread có bị Destroy không?

---

## Bài 5

Một chương trình Go có:

- 8 CPU
- 20 Thread
- 500.000 Goroutine

**Câu hỏi**

Điều này có bình thường không?

Giải thích dựa trên mô hình G-M-P.

---

## Bài 6

Một dịch vụ Go có số lượng Thread tăng liên tục nhưng Goroutine gần như không đổi.

**Câu hỏi**

Bạn sẽ nghi ngờ những nguyên nhân nào?

---

## Bài 7

Một ứng dụng sử dụng rất nhiều CGO.

**Câu hỏi**

CGO có thể ảnh hưởng đến Machine và Scheduler như thế nào?

---

## Bài 8

Một dịch vụ sử dụng `runtime.LockOSThread()` ở nhiều nơi.

**Câu hỏi**

Điều này có thể gây ra những vấn đề gì về hiệu năng và khả năng lập lịch?

---

## Bài 9

Bạn quan sát thấy Runtime Trace xuất hiện rất nhiều sự kiện:

```text
Thread Create

↓

Thread Park

↓

Thread Wake

↓

Thread Create
```

**Câu hỏi**

Bạn sẽ phân tích hiện tượng này như thế nào?

---

## Bài 10

Một Production Server có:

- CPU thấp.
- Latency tăng.
- Thread Count tăng nhanh.
- Blocking Syscall kéo dài.

**Câu hỏi**

Hãy mô tả quy trình Debug của bạn từ đầu đến cuối.

---

# Lời khuyên khi phỏng vấn

Đối với các chủ đề liên quan đến `struct m`, nhà tuyển dụng thường không yêu cầu bạn nhớ chính xác từng trường dữ liệu trong source code.

Điều họ quan tâm hơn là:

- Bạn có hiểu vì sao Go Runtime cần `struct m` hay không.
- Bạn có phân biệt được Machine với OS Thread không.
- Bạn có giải thích được mối quan hệ giữa Goroutine, Machine và Processor không.
- Bạn có hiểu cách Runtime xử lý Blocking System Call và Thread Reuse không.

Khi trả lời, hãy cố gắng:

- Bắt đầu từ bức tranh tổng thể.
- Giải thích bằng sơ đồ G-M-P nếu cần.
- Liên hệ với các cơ chế như `g0`, `mcache`, `entersyscall()` và `exitsyscall()`.
- Đưa ví dụ thực tế trong Production thay vì chỉ lặp lại định nghĩa.

---

---

# 12.36 Tổng kết chương

Chúng ta vừa hoàn thành toàn bộ **Chapter 12 - Machine (`struct m`)**.

Nếu Chapter 11 giúp chúng ta hiểu.

```
G

↓

Goroutine
```

thì Chapter 12 giúp chúng ta hiểu.

```
M

↓

Machine

↓

OS Thread
```

Đây là thành phần thứ hai trong kiến trúc.

```
G-M-P
```

Có thể nói.

Machine chính là cầu nối giữa.

```text
Go Runtime

↓

Operating System
```

Nếu không có Machine.

Go Runtime sẽ không thể.

- Thực thi Goroutine.
- Gọi System Call.
- Context Switching.
- Scheduler.
- Garbage Collector.

---

# Chúng ta đã học những gì?

Trong chương này.

Chúng ta không chỉ học.

```
Machine là gì.
```

Mà còn hiểu.

Toàn bộ vòng đời của.

```
OS Thread
```

bên trong Go Runtime.

---

# Phần I - Tổng quan

Chúng ta bắt đầu.

Bằng những câu hỏi.

- Machine là gì?
- Machine khác OS Thread như thế nào?
- Vì sao Runtime cần `struct m`?
- Mối quan hệ giữa G và M?

Sau phần này.

Chúng ta hiểu.

Machine.

Không phải.

OS Thread.

Mà là.

Runtime Object.

Đại diện.

Cho.

OS Thread.

---

# Phần II - Runtime Internals

Tiếp theo.

Chúng ta đi sâu.

Vào.

```
struct m
```

Bao gồm.

- `curg`
- `g0`
- `mcache`
- `locks`

Đây là những thành phần.

Quan trọng nhất.

Giúp Scheduler.

Quản lý.

Machine.

---

# Chúng ta hiểu

Machine.

Không chỉ lưu.

```
Thread ID
```

Mà còn lưu.

- Current Goroutine.
- Current Processor.
- Scheduler Stack.
- Thread Local Cache.
- Lock Counter.
- GC Metadata.
- Runtime State.

Có thể xem.

Machine.

Là.

"Hồ sơ"

của.

OS Thread.

---

# Phần III - Thread Lifecycle

Đây là phần.

Giúp chúng ta hiểu.

Toàn bộ.

Vòng đời.

Của Machine.

```text
Allocate struct m

↓

Create G0

↓

Create Thread

↓

mstart()

↓

Run

↓

Park

↓

Wakeup

↓

Destroy
```

Sau phần này.

Chúng ta hiểu.

Runtime.

Không tạo.

Thread.

Cho mỗi.

Goroutine.

---

# Chúng ta cũng hiểu

M0.

Là.

Bootstrap Machine.

Được tạo.

Ngay khi.

Runtime.

Khởi động.

Và.

Không giống.

Những Machine khác.

---

# Phần IV - Scheduler Interaction

Đây là phần.

Kết nối.

Machine.

Với.

Goroutine.

Chúng ta đã học.

- Attach.
- Detach.
- Machine Blocking.
- Thread Parking.
- Thread Spinning.

Sau phần này.

Chúng ta hiểu.

Machine.

Không sở hữu.

Goroutine.

Machine.

Chỉ.

Tạm thời.

Thực thi.

Current Goroutine.

---

# Chúng ta cũng hiểu

Một Machine.

Có thể.

Chạy.

```text
G1

↓

G8

↓

G500

↓

G10000
```

Trong suốt.

Vòng đời.

Mà không cần.

Tạo.

Thread mới.

---

# Phần V - System Call

Đây là phần.

Quan trọng nhất.

Của chương.

Chúng ta đã học.

- Blocking Syscall.
- Non-blocking Syscall.
- `entersyscall()`
- `exitsyscall()`

Đồng thời.

Hiểu được.

Một thiết kế.

Rất đẹp.

Của Go Runtime.

```text
Blocking Thread

↓

Không Blocking Scheduler
```

Đây chính là.

Một trong những.

Lý do.

Go.

Có thể.

Xử lý.

Hàng triệu.

Kết nối.

Đồng thời.

---

# Phần VI - Performance

Tiếp theo.

Chúng ta.

Nghiên cứu.

Hiệu năng.

Liên quan.

Đến Machine.

Bao gồm.

- Thread Creation Cost.
- Thread Cache.
- Thread Reuse.
- Performance Optimization.

Sau phần này.

Chúng ta hiểu.

Vì sao.

Go Runtime.

Rất ít.

Tạo Thread.

Mới.

---

# Điều quan trọng

Go Runtime.

Không tối ưu.

Bằng cách.

Làm.

Thread.

Nhẹ hơn.

Go Runtime.

Tối ưu.

Bằng cách.

```
Reuse

↓

Machine
```

Đây là.

Một triết lý.

Rất khác.

So với.

Nhiều Runtime.

Khác.

---

# Phần VII - Debugging

Sau đó.

Chúng ta học.

Cách.

Quan sát.

Machine.

Trong.

Production.

Bao gồm.

- Machine Dump.
- Thread Leak.
- Runtime Trace.

Sau phần này.

Chúng ta hiểu.

Không chỉ.

Runtime.

Hoạt động.

Ra sao.

Mà còn.

Biết.

Làm thế nào.

Debug.

Thread.

Trong.

Production.

---

# Phần VIII - Engineering

Cuối cùng.

Chúng ta.

Tổng hợp.

- Best Practices.
- Common Mistakes.
- Interview Questions.

Đây là phần.

Kết nối.

Runtime.

Với.

Thực tế.

Lập trình.

---

# Bức tranh tổng thể

Sau khi học xong.

Chapter này.

Bạn có thể.

Hình dung.

Toàn bộ.

Vòng đời.

Của Machine.

```text
Runtime Startup

↓

Create M0

↓

Create G0

↓

Create OS Thread

↓

mstart()

↓

Acquire Processor

↓

Run Goroutine

↓

Blocking?

↓

entersyscall()

↓

Release Processor

↓

Kernel

↓

exitsyscall()

↓

Acquire Processor

↓

Continue

↓

Idle

↓

Park

↓

Wakeup

↓

Run

↓

Shutdown

↓

Destroy
```

Đây chính là.

Thread Lifecycle.

Trong Go Runtime.

---

# Những điều quan trọng nhất cần nhớ

Nếu chỉ được giữ lại.

Một vài kiến thức.

Của chương này.

Đó nên là.

---

## 1. Machine không phải OS Thread

Machine.

Là.

Runtime Object.

OS Thread.

Là.

Kernel Object.

Quan hệ.

```text
1 Machine

↓

1 OS Thread
```

---

## 2. Machine không sở hữu Goroutine

Machine.

Chỉ.

Thực thi.

Current Goroutine.

Goroutine.

Có thể.

Chạy.

Trên.

Nhiều Machine.

Khác nhau.

---

## 3. Mỗi Machine đều có G0

G0.

Không chạy.

User Code.

G0.

Là.

Scheduler Stack.

Của Machine.

---

## 4. Thread Creation rất đắt

Đây là lý do.

Runtime.

Ưu tiên.

```text
Park

↓

Reuse
```

Thay vì.

```text
Destroy

↓

Create
```

---

## 5. Blocking Syscall không Block Scheduler

Runtime.

Sử dụng.

```text
entersyscall()

↓

Release Processor

↓

Run Goroutine khác
```

Đây là.

Thiết kế.

Quan trọng nhất.

Liên quan.

Đến Machine.

---

## 6. MCache là một tối ưu lớn

Mỗi Machine.

Có.

```
MCache
```

Riêng.

Giúp.

Giảm.

Lock.

Trong.

Memory Allocation.

---

## 7. Machine chỉ là một phần của G-M-P

Đây là.

Điểm.

Quan trọng nhất.

Machine.

Không Scheduling.

Machine.

Không có.

Run Queue.

Machine.

Không Work Stealing.

Machine.

Không chọn.

Goroutine.

Tất cả.

Những việc đó.

Thuộc về.

```
Processor

(struct p)
```

---

# Kiến thức còn thiếu

Sau Chapter này.

Chúng ta.

Đã hiểu.

- Goroutine.
- Machine.

Nhưng.

Vẫn còn.

Một mảnh ghép.

Quan trọng nhất.

```
Processor

(struct p)
```

Đây chính là.

"Nhạc trưởng"

của.

Scheduler.

Processor.

Mới là nơi.

Quản lý.

- Local Run Queue.
- Work Stealing.
- Scheduling Context.
- Local Cache.
- Timer.

Có thể nói.

Nếu.

```
G

=

Task
```

```
M

=

Worker
```

Thì.

```
P

=

Manager
```

---

# Chuẩn bị cho Chapter 13

Trong chương tiếp theo.

Chúng ta sẽ.

Nghiên cứu.

Toàn bộ.

```
struct p
```

Bao gồm.

- Processor là gì?
- Vì sao Runtime cần Processor?
- Local Run Queue.
- Global Run Queue.
- RunNext.
- Work Stealing.
- Timer.
- Processor Lifecycle.

Đây sẽ là.

Mảnh ghép.

Quan trọng nhất.

Để.

Hiểu.

Toàn bộ.

G-M-P Scheduler.

---

# Góc nhìn của toàn bộ Volume II

Đến thời điểm này.

Chúng ta đã hoàn thành.

```text
Volume II

Go Runtime

│

├── Chapter 10

│     Scheduler

│

├── Chapter 11

│     Goroutine (G)

│

└── Chapter 12

      Machine (M)
```

Hai thành phần.

Đã rõ.

Chỉ còn.

```
Processor

(P)
```

Là.

Mảnh ghép.

Cuối cùng.

Trước khi.

Ghép.

Toàn bộ.

Kiến trúc.

```
G-M-P
```

---

# Hành trình phía trước

Roadmap.

Tiếp theo.

Của Volume II.

```text
Chapter 13

Processor (struct p)

↓

Chapter 14

G-M-P Scheduler

↓

Chapter 15

Channel

↓

Chapter 16

Select

↓

Chapter 17

Netpoll

↓

Chapter 18

Timer

↓

Chapter 19

Memory Allocator

↓

Chapter 20

Garbage Collector

↓

Chapter 21

Synchronization

↓

Chapter 22

Runtime Source Code Walkthrough
```

Đến cuối.

Volume II.

Bạn sẽ.

Không chỉ.

Biết.

Viết Go.

Mà còn.

Hiểu.

Go Runtime.

Ở mức.

Có thể.

Đọc.

Source Code.

Phân tích.

Performance.

Debug.

Production.

Và.

Tự tin.

Trong các buổi.

Phỏng vấn.

Senior.

Staff.

Principal.

---

# Lời kết

Nếu.

**Chapter 11**

giúp chúng ta hiểu.

> **Ai được thực thi?**

Thì.

**Chapter 12**

đã trả lời.

> **Ai thực thi?**

Machine (`struct m`) chính là hiện thân của **OS Thread** trong Go Runtime.

Nó là cầu nối giữa **Go Runtime** và **Kernel**, đồng thời là nơi chứa toàn bộ trạng thái cần thiết để Scheduler, Garbage Collector, Memory Allocator và System Call phối hợp với nhau một cách hiệu quả.

Tuy nhiên.

Machine vẫn chưa phải là trung tâm của Scheduler.

Thành phần giữ vai trò điều phối, quyết định Goroutine nào sẽ chạy và trên Machine nào chính là:

> **Processor (`struct p`)**

Đó sẽ là chủ đề của **Chapter 13**, cũng là chương quan trọng nhất để hoàn thiện bức tranh về kiến trúc **G-M-P** của Go Runtime.

**Chapter 12 kết thúc tại đây.**

---