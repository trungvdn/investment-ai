# Glossary

Version: 1.0

---

# 1. Purpose

Glossary định nghĩa ngôn ngữ chung (Ubiquitous Language) cho toàn bộ hệ thống Investment AI.

Mọi tài liệu, capability, workflow, API và agent phải sử dụng các thuật ngữ trong tài liệu này.

Nếu xuất hiện thuật ngữ mới, phải bổ sung vào Glossary trước khi sử dụng trong hệ thống.

---

# 2. Core Business Concepts

## Stock

Một mã cổ phiếu được niêm yết trên thị trường chứng khoán.

Ví dụ:

* FPT
* VCB
* HPG

---

## Company

Doanh nghiệp phát hành cổ phiếu.

Một Company có thể có:

* Business Model
* Financial Statements
* Management
* Competitive Advantage
* Corporate Actions

---

## Sector

Nhóm doanh nghiệp hoạt động trong cùng một ngành.

Ví dụ:

* Banking
* Securities
* Steel
* Real Estate
* Retail

---

## Market

Toàn bộ thị trường chứng khoán.

Ví dụ:

* VN-Index
* HNX
* UPCoM

---

# 3. Investment Concepts

## Investment Thesis

Luận điểm đầu tư được xây dựng từ nhiều bằng chứng.

Investment Thesis trả lời:

* Tại sao nên đầu tư?
* Điều gì tạo ra lợi nhuận?
* Rủi ro là gì?

---

## Recommendation

Khuyến nghị cuối cùng của hệ thống.

Giá trị chuẩn:

* Strong Buy
* Buy
* Hold
* Sell
* Strong Sell

---

## Confidence Score

Độ tin cậy của Recommendation.

Khoảng giá trị:

0–100

Confidence phản ánh mức độ chắc chắn của hệ thống dựa trên chất lượng dữ liệu và sự đồng thuận giữa các capability.

Không phản ánh xác suất giá cổ phiếu sẽ tăng hoặc giảm.

---

## Fair Value

Giá trị nội tại của doanh nghiệp được ước tính thông qua các phương pháp định giá.

---

## Margin of Safety

Khoảng chênh lệch giữa Fair Value và Market Price.

Được sử dụng để đánh giá mức độ an toàn khi đầu tư.

---

# 4. Macro Concepts

## GDP

Tổng sản phẩm quốc nội.

Đại diện cho tốc độ tăng trưởng của nền kinh tế.

---

## CPI

Chỉ số giá tiêu dùng.

Đại diện cho mức lạm phát.

---

## Interest Rate

Lãi suất điều hành hoặc lãi suất thị trường.

Có ảnh hưởng lớn đến:

* Thanh khoản
* Định giá
* Chu kỳ kinh tế

---

## PMI

Purchasing Managers' Index.

Được sử dụng để đánh giá hoạt động sản xuất.

---

## Liquidity

Lượng tiền đang lưu thông trên thị trường.

Không chỉ phản ánh khối lượng giao dịch mà còn phản ánh khả năng hấp thụ dòng tiền.

---

# 5. Market Analysis Concepts

## Trend

Xu hướng chính của thị trường hoặc cổ phiếu.

Giá trị chuẩn:

* Bullish
* Neutral
* Bearish

---

## Support

Vùng giá có khả năng xuất hiện lực mua.

---

## Resistance

Vùng giá có khả năng xuất hiện lực bán.

---

## Breakout

Giá vượt qua vùng kháng cự hoặc vùng tích lũy.

---

## Pullback

Giá điều chỉnh sau một đợt tăng hoặc giảm.

---

# 6. Wyckoff Concepts

## Composite Operator (CO)

Thực thể đại diện cho dòng tiền lớn trên thị trường.

Không phải là một cá nhân cụ thể mà là mô hình khái quát hóa hành vi của các tổ chức.

---

## Accumulation

Giai đoạn hấp thụ nguồn cung để chuẩn bị cho xu hướng tăng.

---

## Markup

Giai đoạn giá tăng sau khi hoàn tất tích lũy.

---

## Distribution

Giai đoạn phân phối cổ phiếu cho thị trường.

---

## Markdown

Giai đoạn giá giảm sau khi phân phối.

---

## Spring

Hiện tượng phá vỡ giả xuống dưới vùng hỗ trợ nhằm kiểm tra hoặc hấp thụ nguồn cung.

---

## Upthrust

Hiện tượng phá vỡ giả lên trên vùng kháng cự nhằm kiểm tra hoặc phân phối nguồn cầu.

---

# 7. VSA Concepts

## No Supply

Thanh khoản thấp trong nhịp giảm.

Cho thấy áp lực bán suy yếu.

---

## No Demand

Thanh khoản thấp trong nhịp tăng.

Cho thấy lực mua yếu.

---

## Stopping Volume

Khối lượng lớn xuất hiện sau xu hướng giảm.

Có thể báo hiệu sự hấp thụ lực bán.

---

## Buying Climax

Khối lượng cực lớn xuất hiện sau xu hướng tăng mạnh.

Có thể báo hiệu kết thúc xu hướng tăng.

---

## Selling Climax

Khối lượng cực lớn xuất hiện sau xu hướng giảm mạnh.

Có thể báo hiệu kết thúc xu hướng giảm.

---

# 8. Analysis Concepts

## Analysis

Kết quả đánh giá của một capability.

Ví dụ:

* Macro Analysis
* Technical Analysis
* Fundamental Analysis

Mỗi Analysis phải có:

* Summary
* Evidence
* Score
* Confidence

---

## Evidence

Thông tin hoặc dữ liệu hỗ trợ cho một kết luận.

Mọi kết luận trong hệ thống phải liên kết với ít nhất một Evidence.

---

## Signal

Một tín hiệu được phát hiện từ dữ liệu.

Ví dụ:

* Golden Cross
* Spring
* No Supply
* Earnings Growth

Signal không phải là Recommendation.

---

# 9. Reporting Concepts

## Executive Summary

Tóm tắt ngắn gọn kết quả phân tích.

---

## Investment Report

Báo cáo đầu tư hoàn chỉnh bao gồm:

* Executive Summary
* Macro Analysis
* Sector Analysis
* Company Analysis
* Financial Analysis
* Valuation
* Technical Analysis
* Risk Analysis
* Recommendation

---

# 10. Naming Rules

Để đảm bảo tính nhất quán trong toàn hệ thống:

* Một thuật ngữ chỉ có một định nghĩa.
* Không sử dụng từ đồng nghĩa cho cùng một khái niệm.
* Recommendation khác Signal.
* Confidence khác Probability.
* Fair Value khác Market Price.
* Investment Thesis khác Recommendation.
* Capability khác Agent.
* Workflow khác Capability.
* Analysis khác Report.

