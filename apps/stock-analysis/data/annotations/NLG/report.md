# NLG — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn, chờ người dùng duyệt. review.status=in_review; eligible_for_ai=false. Phân tích từ OHLCV; chưa đối chiếu nhãn trên ảnh.

## 1. Thông tin và dữ liệu

- Khung 1D; cutoff 2026-10-30.
- [candles_NLG_1D.csv](candles_NLG_1D.csv): 497 phiên từ 31/10/2024 đến 30/10/2026; cột event trống.
- [NLG_nhan.png](NLG_nhan.png): ảnh tham chiếu để người duyệt đối chiếu.
- source_commit: a86721517b4490cb9731326dcce97d3dee27f344.
- Close cuối 22 theo đơn vị CSV; đơn vị và điều chỉnh giá chưa xác minh.
- Không trùng ngày; OHLC hợp lệ, volume không âm. Chưa đối chiếu đủ phiên theo lịch thị trường.
- RV20 = volume / trung bình 20 phiên liền trước, không tính phiên hiện tại. CLV=(close-low)/(high-low), bỏ khi high=low.

## 2. Kết luận nhanh

**Cấu trúc chưa rõ; giả thuyết ưu tiên là tích lũy nhỏ thử chuyển C→D, tin cậy thấp.** Cụm 14–15/10 là Spring ứng viên của vùng nhỏ 21.3–22.5; 21/10 là SOS ứng viên. Tuy nhiên, 30/10 đóng ở đáy phiên 22, dưới vùng breakout: **chưa có LPS xác nhận**. Pha D không phải kết luận chắc chắn, và không áp dụng cho toàn bộ vùng tháng 7–9.

## 3. Bối cảnh và phạm vi

- Vùng bối cảnh rộng: low 28/07=19.15; hồi đến high 07/09=25.95; sau đó giảm đến low 28/09=21.35.\n- Vùng chính trong ví dụ là vùng nhỏ cuối tháng 9–10, không phải toàn bộ vùng tháng 7–9.\n- Cụm 14–15/10 xuyên 21.3 nhưng không xuyên 19.15; nhãn Spring chỉ có thể áp dụng cho vùng nhỏ đã chọn.

Không mặc định đây là tái tích lũy: cần bối cảnh xu hướng tăng lớn hơn để phân biệt tích lũy, tái tích lũy và hồi kỹ thuật. Chưa bổ sung nhận định ngành/vĩ mô vì không có dữ liệu tương ứng trong mẫu.

## 4. Vùng giao dịch

| Phạm vi | Hỗ trợ | Kháng cự | Trạng thái |
|---|---|---|---|
| NLG_TR_202609_SHORT, từ 28/09 | 21.3–21.4 | 22.2–22.5 | TR ngắn, provisional |
| NLG_CONTEXT_202607, từ 28/07 | 19.15–20.0 | 25.7–25.95 | Dải bối cảnh, chưa xác nhận TR |

Biên nhỏ có trước Spring: 28–29/09 low 21.35, 01/10 low 21.3; 08/10 high 22.2, 09/10 high 22.5. Sau cụm xuyên đáy, 21.0 là đáy Spring ứng viên; không hồi tố sửa biên trước sự kiện xuống 21.0. TR ngắn nên không ép có đủ PS/SC/AR/ST.

## 5. Nhật ký nhãn sự kiện

Tất cả label_verdict=unreviewed và final_label=null. invalidated là đề xuất bác bỏ của AI. Nhận định sử dụng kiến thức đến cutoff, không giả vờ biết phản ứng tương lai ngay tại ngày sự kiện.

### E001 — SC: 2026-09-28

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=22.45/22.45/21.35/21.65; close ở 27.3% biên độ sau nhịp giảm từ 25.95.
- Bằng chứng volume: Volume 2265200; RV20 1.24, không phải volume cực đoan.
- Phản ứng đã quan sát: 29/09 giữ low 21.35; 01/10 low 21.3; tiếp theo dao động quanh 21.3–22.5.
- Bằng chứng phản đối/giới hạn: Climax không rõ; nên cho phép no_event/stopping action thay vì ép SC.
- Điều kiện xác nhận: Cần xác định nhịp bán cuối và thay đổi hành vi, không chỉ chọn đáy.
- Điều kiện bác bỏ: Nếu không đủ bằng chứng climax thì bỏ SC nhưng vẫn giữ vùng đi ngang ứng viên.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E002 — AR: 2026-10-01 → 2026-10-09

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 01/10 close=high=21.85 sau low 21.3; 09/10 high 22.5, close 21.9.
- Bằng chứng volume: 01/10 volume 2687800 RV20 1.42; 09/10 volume 1248000 RV20 0.73.
- Phản ứng đã quan sát: Sau đỉnh 09/10, giá quay xuống cụm 14–15/10.
- Bằng chứng phản đối/giới hạn: SC còn uncertain; nếu SC không được duyệt, đây chỉ là nhịp hồi tạo biên trên, nhãn AR cần xét lại.
- Điều kiện xác nhận: Có hồi sau stopping action và phản ứng tạo cạnh trên.
- Điều kiện bác bỏ: Nếu vùng hoặc SC gốc bị bác bỏ, đánh giá lại nhãn AR.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E003 — Spring: 2026-10-14 → 2026-10-15

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 14/10 low 21.15, close 21.2 dưới hỗ trợ 21.3; 15/10 low 21.0 rồi close 21.4 trở lại vùng.
- Bằng chứng volume: 14/10 volume 1102700; 15/10 921400 RV20 0.64.
- Phản ứng đã quan sát: 16–20/10 low không thấp hơn 21.2; 21/10 vượt vùng phía trên.
- Bằng chứng phản đối/giới hạn: Chỉ là Spring của TR ngắn; 15/10 low vẫn cao hơn đáy bối cảnh rộng 19.15.
- Điều kiện xác nhận: Xuyên biên đã tồn tại trước sự kiện, hồi vào vùng rồi giữ đáy 21.0.
- Điều kiện bác bỏ: Nếu chấp nhận giá dưới 21.0 kéo dài, bác bỏ giả thuyết Spring thành công.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E004 — Test: 2026-10-20

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 21.2 cao hơn Spring low 21.0; close 21.4; biên độ 0.3.
- Bằng chứng volume: Volume 1091400 RV20 0.79 nhưng cao hơn volume 15/10=921400 khoảng 18.5%.
- Phản ứng đã quan sát: 21/10 tăng mạnh close 22.85.
- Bằng chứng phản đối/giới hạn: Volume không giảm so phiên Spring; hồi sau đó không tự chứng minh đây là Test chất lượng.
- Điều kiện xác nhận: Đánh giá đồng thời biên độ, giữ đáy và nguồn cung; không chỉ so với RV20.
- Điều kiện bác bỏ: Nếu yêu cầu Test volume thấp hơn Spring thì nhãn này không đạt; có thể để no_event.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E005 — SOS: 2026-10-21

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=21.65/22.85/21.65/22.85; vượt 22.5 khoảng 1.56%, đóng ở đỉnh phiên.
- Bằng chứng volume: Volume 2765500; RV20 1.97. Ngày 22/10 volume 3140500 RV20 2.17.
- Phản ứng đã quan sát: 22–23/10 close 23.15, 23.3; nhưng 30/10 đóng 22 trở lại dưới 22.5.
- Bằng chứng phản đối/giới hạn: Chỉ là breakout TR nhỏ; chưa vượt đỉnh bối cảnh rộng 25.95.
- Điều kiện xác nhận: Cần giữ hoặc lấy lại 22.5, có follow-through và pullback được cầu hỗ trợ.
- Điều kiện bác bỏ: Hai close dưới 22.2 và không lấy lại 22.5 trong ba phiên sau làm suy yếu breakout; đến cutoff chưa đủ diễn biến sau 30/10.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E006 — LPS: 2026-10-27 → 2026-10-30

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 28/10 low 22.45, close 23; 30/10 low=close=22, mất vùng breakout 22.2–22.5.
- Bằng chứng volume: 28/10 volume 904800 RV20 0.60; 30/10 1044600 RV20 0.75, tăng so 28/10.
- Phản ứng đã quan sát: Không có dữ liệu hồi sau 30/10.
- Bằng chứng phản đối/giới hạn: Close đáy phiên 30/10 và mất vùng breakout không phù hợp LPS đã xác nhận.
- Điều kiện xác nhận: Chờ hồi lấy lại 22.5 rồi kiểm tra giữ vùng với cung giảm.
- Điều kiện bác bỏ: Nếu tiếp tục giảm dưới vùng và không phục hồi, không xác nhận LPS.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E007 — Test: 2026-10-12

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 21.75, close 21.95; ngày này xảy ra trước Spring 14–15/10.
- Bằng chứng volume: Volume 322900 RV20 0.20: volume thấp không sửa được sai thứ tự.
- Phản ứng đã quan sát: Sau đó 14–15/10 mới xuất hiện xuyên hỗ trợ.
- Bằng chứng phản đối/giới hạn: Chỉ bác bỏ nhãn Test sau Spring đối với cấu trúc này, không bác bỏ mọi loại kiểm tra cung.
- Điều kiện xác nhận: Test sau Spring phải xảy ra sau Spring và có quan hệ với sự kiện đó.
- Điều kiện bác bỏ: 12/10 trước 14–15/10 nên không phải Test của Spring này.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E008 — SOS: 2026-10-21

- structure_id: NLG_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close 22.85 vẫn dưới đỉnh bối cảnh rộng 25.95.
- Bằng chứng volume: Volume lớn không thay thế điều kiện breakout khỏi đúng vùng.
- Phản ứng đã quan sát: Nhịp hồi cao nhất 26/10 high 24.5 vẫn dưới 25.95.
- Bằng chứng phản đối/giới hạn: Đây là ví dụ nhãn SOS sai ở phạm vi rộng; E005 vẫn là SOS ứng viên vùng nhỏ.
- Điều kiện xác nhận: Phải nêu đúng structure_id và biên cần vượt.
- Điều kiện bác bỏ: Không gán breakout trên 25.95 cho phiên đóng 22.85.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Giới hạn |
|---|---|---|
| 15/10 low 21, close 21.4; RV20 0.64 | Xuyên biên nhỏ rồi hồi, Spring ứng viên | Chưa xuyên đáy rộng 19.15; TR ngắn |
| 20/10 volume 1,091,400 cao hơn 15/10 khoảng 18.5% | Test còn nghi ngờ | RV20 thấp không đồng nghĩa volume thấp hơn Spring |
| 21/10 close=high=22.85; RV20 1.97 | Nỗ lực tăng có kết quả vượt vùng nhỏ | Không phải breakout trên 25.95 |
| 26/10 high 24.5, close 23, CLV 6.25% | Cầu không giữ được thành quả trong phiên | Không suy ra danh tính hoặc ý định bên bán |
| 30/10 close=low=22; RV20 0.75 | Giá giảm dù volume thấp, cần theo dõi thiếu cầu/cung còn lại | Không gọi LPS đã xác nhận chỉ vì volume giảm |

Pha D chỉ là giả thuyết đang bị kiểm tra. Nếu mất breakout và không lấy lại, hạ về dao động chưa rõ pha hoặc hồi kỹ thuật; không duy trì nhãn D bất chấp bằng chứng mới.

## 7. Kịch bản

- **S1: Giữ được breakout vùng nhỏ, củng cố C→D.** Xác nhận: Lấy lại 22.5 và giữ qua ít nhất hai close, pullback sau đó giữ vùng với cung giảm. Bác bỏ: Giảm tiếp, không lấy lại vùng; mất 21.0 làm suy yếu Spring.

- **S2: Breakout thất bại, quay lại dao động.** Xác nhận: Giá ở dưới 22.2 nhưng vẫn giữ 21.0; chưa có đà theo một hướng. Bác bỏ: Thoát vùng bền vững.

- **S3: Markdown tiếp diễn.** Xác nhận: Hai close liên tiếp dưới 21.0 và không hồi trong ba phiên sau. Bác bỏ: Hồi nhanh lấy lại vùng và kiểm tra giữ đáy.

Các ngưỡng hai/ba phiên là quy ước minh họa để điều kiện kiểm tra được, chưa backtest và không phải quy tắc Wyckoff bắt buộc.

## 8. Duyệt mẫu

- Trạng thái: in_review; eligible_for_ai=false.
- Người duyệt/ngày duyệt: chưa điền. Chưa có nhãn đúng/sai được người dùng xác nhận.
- Cần duyệt: TR nhỏ có đủ ý nghĩa hay không; SC 28/09 và AR 01–09/10; Spring cụm 14–15/10; Test 20/10; SOS 21/10; LPS chưa xác nhận.
- Nhãn sai AI đề xuất: E007 Test sau Spring ngày 12/10 sai thứ tự; E008 SOS breakout vùng rộng ngày 21/10 sai phạm vi. E005 vẫn là SOS ứng viên của vùng nhỏ.
- Đối chiếu ảnh và điều chỉnh giá; cập nhật final_label, label_verdict, correction_reason, reviewer cùng JSON. Chỉ chuyển approved sau khi người dùng duyệt.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Chưa có dữ liệu sau 30/10/2026. Giữ evaluation_date=null, observations=[]; không tự kết luận các kịch bản đã xảy ra.
