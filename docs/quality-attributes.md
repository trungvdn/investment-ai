# Quality Attributes

Version: 1.0

---

# 1. Purpose

Tài liệu này định nghĩa các thuộc tính chất lượng (Quality Attributes) của hệ thống Investment AI.

Quality Attributes được sử dụng để:

* Đánh giá kiến trúc.
* Đánh giá capability.
* Đánh giá workflow.
* Đánh giá AI model.
* Đánh giá phiên bản phát hành.

Mọi quyết định kỹ thuật cần cân nhắc tác động đến các thuộc tính này.

---

# 2. Priority

Đối với MVP, mức độ ưu tiên được xác định như sau:

| Attribute       | Priority  |
| --------------- | --------- |
| Accuracy        | Critical  |
| Explainability  | Critical  |
| Traceability    | Critical  |
| Reliability     | High      |
| Extensibility   | High      |
| Testability     | High      |
| Performance     | Medium    |
| Cost Efficiency | Medium    |
| Reproducibility | Medium    |
| Scalability     | Low (MVP) |

---

# 3. Accuracy

## Goal

Đưa ra kết quả phân tích chính xác và nhất quán.

## Principles

* Không suy đoán khi thiếu dữ liệu.
* Không tạo dữ liệu mới.
* Luôn sử dụng nguồn dữ liệu đáng tin cậy.
* Ưu tiên đúng hơn nhanh.

## Measurement

Ví dụ:

* % phân tích được chuyên gia chấp nhận.
* % recommendation hợp lý khi backtest.
* % dữ liệu được sử dụng đúng.

---

# 4. Explainability

## Goal

Mọi kết luận đều phải giải thích được.

## Requirement

Recommendation phải truy vết được:

Recommendation

↓

Investment Thesis

↓

Analysis

↓

Evidence

↓

Raw Data

Nếu không thể truy vết thì kết quả không hợp lệ.

---

# 5. Traceability

Mọi output phải biết được:

* Dữ liệu nào được sử dụng.
* Capability nào tạo ra.
* Workflow nào gọi.
* Model nào sinh kết quả.
* Prompt nào được sử dụng (nếu có).

Điều này giúp:

* Debug.
* Audit.
* So sánh các phiên bản.

---

# 6. Reliability

Hệ thống phải hoạt động ổn định ngay cả khi:

* Thiếu dữ liệu.
* API lỗi.
* Một capability thất bại.
* LLM timeout.

Yêu cầu:

* Retry hợp lý.
* Fallback.
* Graceful Degradation.

---

# 7. Extensibility

Có thể bổ sung capability mới mà không sửa workflow hiện có.

Ví dụ:

Thêm:

* ESG Analysis
* Insider Trading Analysis
* Options Analysis

không làm thay đổi các capability khác.

---

# 8. Testability

Mỗi capability phải kiểm thử độc lập.

Input cố định.

Output xác định.

Workflow có thể được kiểm thử bằng mock capability.

---

# 9. Performance

MVP không tối ưu tuyệt đối về tốc độ nhưng cần đảm bảo trải nghiệm người dùng.

Mục tiêu:

* Báo cáo cổ phiếu hoàn chỉnh trong thời gian chấp nhận được.
* Hạn chế các bước xử lý không cần thiết.
* Hỗ trợ chạy song song khi không có phụ thuộc dữ liệu.

---

# 10. Cost Efficiency

Chi phí AI phải được kiểm soát.

Ưu tiên:

* Rule-based trước.
* AI khi thật sự cần.
* Cache các kết quả ít thay đổi.
* Tái sử dụng dữ liệu.

---

# 11. Reproducibility

Cùng:

* Input
* Data
* Prompt
* Model
* Configuration

phải tạo ra kết quả nhất quán.

Điều này giúp:

* Regression Testing.
* Benchmark.
* Debug.

---

# 12. Scalability

Kiến trúc phải cho phép mở rộng:

* Capability.
* Workflow.
* Data Source.
* AI Model.

Không cần tối ưu hạ tầng ở giai đoạn MVP nhưng không được tạo rào cản cho các phiên bản sau.

---

# 13. Observability

Mọi workflow cần ghi nhận:

* Start Time.
* End Time.
* Duration.
* Input.
* Output.
* Error.
* Retry.
* Token Usage.
* Cost.

Những dữ liệu này phục vụ:

* Monitoring.
* Performance Tuning.
* Cost Optimization.

---

# 14. Security

Nguyên tắc:

* Không lưu trữ API Key trong mã nguồn.
* Kiểm soát quyền truy cập theo capability.
* Mọi thao tác quan trọng cần có audit log.
* Dữ liệu người dùng được tách biệt với dữ liệu thị trường.

---

# 15. Evaluation Checklist

Một capability được xem là đạt yêu cầu khi:

* Chính xác.
* Có thể giải thích.
* Có thể truy vết.
* Có thể kiểm thử.
* Có thể tái sử dụng.
* Có thể mở rộng.
* Có khả năng xử lý lỗi.
* Đáp ứng hiệu năng mong muốn.

Nếu không đạt các tiêu chí trên thì không được đưa vào workflow chính thức.
