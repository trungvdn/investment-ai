# Knowledge Model

Version: 1.0

---

# 1. Purpose

Knowledge Model định nghĩa cách tri thức được hình thành, biến đổi và sử dụng trong Investment AI.

Đây là mô hình trung tâm kết nối:

* Data
* Capability
* Workflow
* AI
* Reporting

Knowledge Model là nền tảng cho:

* Analysis Engine
* RAG
* Reflection
* Multi-Agent
* Memory

---

# 2. Knowledge Transformation Pipeline

```text
External Sources
        │
        ▼
Raw Data
        │
        ▼
Normalized Data
        │
        ▼
Evidence
        │
        ▼
Signal
        │
        ▼
Analysis
        │
        ▼
Investment Thesis
        │
        ▼
Recommendation
        │
        ▼
Investment Report
```

Mỗi bước làm tăng giá trị của tri thức.

Không được bỏ qua bất kỳ bước nào.

---

# 3. Raw Data

Raw Data là dữ liệu gốc được thu thập từ bên ngoài.

Ví dụ:

* OHLCV
* Financial Statements
* GDP
* CPI
* PMI
* News
* Interest Rate

Đặc điểm:

* Chưa chuẩn hóa.
* Chưa có ngữ nghĩa.
* Không chứa kết luận.

---

# 4. Normalized Data

Raw Data sau khi được chuẩn hóa.

Ví dụ:

* Chuẩn hóa đơn vị tiền tệ.
* Chuẩn hóa múi giờ.
* Chuẩn hóa tên doanh nghiệp.
* Chuẩn hóa định dạng ngày.
* Chuẩn hóa mã cổ phiếu.

Đặc điểm:

* Có cấu trúc.
* Có schema thống nhất.
* Có thể sử dụng bởi mọi capability.

---

# 5. Evidence

Evidence là dữ liệu đã được diễn giải và có khả năng hỗ trợ một kết luận.

Ví dụ:

* ROE tăng liên tục trong 5 năm.
* CPI giảm trong 3 tháng gần nhất.
* Khối lượng giao dịch cao hơn trung bình 20 phiên.
* Doanh thu tăng trưởng 18% YoY.

Đặc điểm:

* Có nguồn gốc rõ ràng.
* Có thể truy vết về Raw Data.
* Chưa phải kết luận đầu tư.

---

# 6. Signal

Signal là mẫu hình hoặc hiện tượng được phát hiện từ Evidence.

Ví dụ:

* Golden Cross.
* Spring.
* SOS.
* No Supply.
* Earnings Acceleration.
* Margin Expansion.

Signal trả lời:

"Có điều gì đáng chú ý đang xảy ra?"

Signal không trả lời:

"Có nên đầu tư hay không?"

---

# 7. Analysis

Analysis là kết quả đánh giá của một capability.

Ví dụ:

Macro Analysis

Technical Analysis

Fundamental Analysis

Valuation Analysis

Mỗi Analysis phải có:

* Summary
* Score
* Confidence
* Evidence
* Signals
* Risks

Analysis chỉ đánh giá trong phạm vi capability của mình.

---

# 8. Investment Thesis

Investment Thesis là sự tổng hợp có cấu trúc từ nhiều Analysis.

Bao gồm:

* Opportunities.
* Risks.
* Catalysts.
* Assumptions.
* Key Drivers.

Investment Thesis trả lời:

"Tại sao nên hoặc không nên đầu tư?"

---

# 9. Recommendation

Recommendation là quyết định cuối cùng của hệ thống.

Giá trị chuẩn:

* Strong Buy
* Buy
* Hold
* Sell
* Strong Sell

Recommendation phải dựa trên:

Investment Thesis

↓

Analysis

↓

Evidence

↓

Raw Data

Recommendation không được sinh trực tiếp từ Raw Data.

---

# 10. Investment Report

Investment Report là biểu diễn cuối cùng của toàn bộ Knowledge Pipeline.

Bao gồm:

* Executive Summary.
* Investment Thesis.
* Recommendation.
* Supporting Analysis.
* Supporting Evidence.
* Appendix.

Report không tạo tri thức mới.

Report chỉ trình bày tri thức.

---

# 11. Knowledge Hierarchy

```text
Report
▲
Recommendation
▲
Investment Thesis
▲
Analysis
▲
Signal
▲
Evidence
▲
Normalized Data
▲
Raw Data
```

Càng lên cao:

* Giá trị tri thức càng lớn.
* Khối lượng dữ liệu càng nhỏ.
* Khả năng ra quyết định càng cao.

---

# 12. Knowledge Rules

## Rule 1

Raw Data không được dùng để sinh Recommendation.

---

## Rule 2

Mọi Recommendation phải có Investment Thesis.

---

## Rule 3

Mọi Investment Thesis phải có ít nhất một Analysis.

---

## Rule 4

Mọi Analysis phải có Evidence.

---

## Rule 5

Mọi Evidence phải truy vết được về Raw Data.

---

## Rule 6

Signal chỉ là đầu vào của Analysis.

Signal không phải Recommendation.

---

## Rule 7

Report không được tạo tri thức mới.

---

# 13. Traceability

Mọi kết luận phải truy vết được.

```text
Recommendation

↓

Investment Thesis

↓

Analysis

↓

Evidence

↓

Normalized Data

↓

Raw Data
```

Đây là nguyên tắc quan trọng nhất của Explainable AI.

---

# 14. Future Evolution

Knowledge Model được thiết kế để mở rộng.

Các phiên bản sau có thể bổ sung:

* Knowledge Graph.
* Reflection Knowledge.
* Episodic Memory.
* Semantic Memory.
* Cross-Market Knowledge.
* Portfolio Knowledge.

Không làm thay đổi Knowledge Pipeline hiện tại.
