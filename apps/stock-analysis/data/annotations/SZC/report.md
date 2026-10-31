# SZC — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn để minh họa cách điền mẫu. Trạng thái in_review; chưa dùng làm mẫu chuẩn cho AI. Nhãn trên ảnh chưa được đối chiếu; mọi nhãn dưới đây suy luận từ OHLCV.

## 1. Thông tin và dữ liệu

- Mã: SZC; khung 1D; ngày phân tích: 2026-10-30.
- OHLCV: [candles_SZC_1D.csv](candles_SZC_1D.csv). 497 phiên từ 31/10/2024 đến 30/10/2026.
- Ảnh tham chiếu để người duyệt đối chiếu: [SZC_nhan.png](SZC_nhan.png).
- Commit dữ liệu: ebec9ce72b4f3ce1fea6908e27ca3c59d104db6c.
- Close cuối: 19.3 theo đơn vị CSV; chưa xác minh đơn vị/điều chỉnh giá với nguồn.
- Không trùng ngày; OHLC hợp lệ, volume không âm. Chưa kiểm tra đủ phiên theo lịch sở giao dịch. Cột event trống.
- RV20 = volume / trung bình 20 phiên TRƯỚC phiên đang xét. Không tính phiên hiện tại.

## 2. Kết luận nhanh

Ưu tiên giả thuyết **tích lũy, pha D đang được kiểm tra**, tin cậy trung bình theo đánh giá định tính. Nhịp tăng 22–23/10 vượt đỉnh hồi tháng 8 nhưng đóng cửa 29–30/10 quay xuống dưới vùng breakout, nên chưa có LPS được xác nhận. Nếu không giữ được breakout, cần hạ đánh giá về dao động pha B hoặc hồi kỹ thuật trong xu hướng giảm.

## 3. Bối cảnh trước cấu trúc

Close giảm từ 35.003 ngày 05/02/2026 xuống 16.268 ngày 27/07. Phiên 29/05 volume rất cao nhưng không chấm dứt xu hướng giảm: các tháng sau vẫn tạo đáy thấp hơn. Cụm 27–28/07 được chọn làm ứng viên dừng giảm cho vùng đang xét. Chưa bổ sung bối cảnh ngành/vĩ mô vì không có dữ liệu đó trong mẫu.

## 4. Trading Range

| Thành phần | Vùng/ngày | Bằng chứng |
|---|---|---|
| structure_id | SZC_TR_202607 | Vùng ứng viên, chưa được người duyệt xác nhận |
| Bắt đầu | 27–28/07/2026 | Low 16.268 rồi 16.031 |
| Hỗ trợ | 16.031–16.268 | Cụm đáy tháng 7; 06/10 low 16.2 |
| Kháng cự | 19.446–19.588 | High ngày 10/08 và 12/08 |
| Kết thúc | Chưa chốt | Breakout 22–23/10 đang bị kiểm tra |

## 5. Nhật ký nhãn sự kiện

Các trạng thái candidate/invalidated dưới đây là đánh giá AI; **label_verdict của mọi sự kiện vẫn unreviewed**, final_label để null. “Invalidated” không đồng nghĩa người dùng đã duyệt nhãn sai.

### E001 — PS (2026-07-23)

- structure_id: SZC_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 16.885, close 17.881, vị trí đóng cửa 95.5% biên độ.
- Bằng chứng volume: 496100; RV20 1.61.
- Phản ứng đã quan sát đến ngày phân tích: 27–28/07 tiếp tục tạo đáy thấp hơn.
- Điều kiện xác nhận: Phản ứng cầu tại 23/07 và chuyển sang cụm SC/AR tiếp theo.
- Điều kiện bác bỏ: Nếu chỉ là một nhịp hồi nhỏ, thiếu thay đổi hành vi thì bỏ nhãn PS.
- Trạng thái AI đề xuất: candidate; kết quả duyệt nhãn: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E002 — SC (2026-07-27 → 2026-07-28)

- structure_id: SZC_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 27/07 O/H/L/C=17.217/17.454/16.268/16.268. 28/07 low 16.031, close 16.505; cụm hai phiên, không coi 27/07 là đáy cuối cùng.
- Bằng chứng volume: 27/07 volume 693400, RV20 2.13; 28/07 volume 446500.
- Phản ứng đã quan sát đến ngày phân tích: Hồi đến 19.588 ngày 12/08; 06/10 low 16.2 chưa phá đáy cụm.
- Điều kiện xác nhận: Cụm dừng giảm có AR và kiểm tra đáy về sau.
- Điều kiện bác bỏ: Nếu không tạo được vùng đi ngang và tiếp tục giảm bền dưới 16.031, bác bỏ giả thuyết cụm SC mở đầu vùng.
- Trạng thái AI đề xuất: candidate; kết quả duyệt nhãn: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E003 — AR (2026-08-10 → 2026-08-12)

- structure_id: SZC_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 10/08 close=high=19.446; 12/08 high 19.588.
- Bằng chứng volume: 10/08 volume 1278300, RV20 3.53.
- Phản ứng đã quan sát đến ngày phân tích: Giá điều chỉnh sau đỉnh, hình thành cạnh trên tham chiếu.
- Điều kiện xác nhận: Đỉnh hồi sau SC được theo sau bởi phản ứng giảm và dao động trong vùng.
- Điều kiện bác bỏ: Nếu xác định một vùng khác, phải xem lại AR và biên trên tương ứng.
- Trạng thái AI đề xuất: candidate; kết quả duyệt nhãn: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E004 — ST (2026-10-06)

- structure_id: SZC_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=16.55/16.8/16.2/16.7; low cao hơn 16.031 khoảng 1.05%.
- Bằng chứng volume: Volume 293000, bằng 42.3% volume 27/07; RV20 1.03, không phải thấp hơn trung bình 20 phiên.
- Phản ứng đã quan sát đến ngày phân tích: Giá hồi lên 17.75 high ngày 16/10 rồi tăng mạnh 20–23/10.
- Điều kiện xác nhận: Giữ đáy cụm tháng 7 và có phản ứng hồi; đánh giá ST muộn cần người duyệt.
- Điều kiện bác bỏ: Nếu dùng vùng hỗ trợ khác hoặc không có quan hệ kiểm tra với cụm SC, bỏ nhãn ST.
- Trạng thái AI đề xuất: candidate; kết quả duyệt nhãn: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E005 — SOS (2026-10-22 → 2026-10-23)

- structure_id: SZC_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 22/10 close 19.6 vượt 19.588 chỉ 0.012; 23/10 close 20.2 và high 20.25 củng cố.
- Bằng chứng volume: 22/10 volume 1003700 RV20 2.87; 23/10 volume 1286300 RV20 3.41.
- Phản ứng đã quan sát đến ngày phân tích: 29–30/10 đóng 19.4 và 19.3, quay dưới vùng đỉnh AR.
- Điều kiện xác nhận: Cần phục hồi và giữ trên 19.588; bộ ngưỡng vận hành cần kiểm định.
- Điều kiện bác bỏ: Hai phiên đóng dưới 19.446 và không lấy lại trong 3 phiên sau làm suy yếu breakout. Đến 30/10 mới có phần đầu điều kiện, chưa biết 3 phiên sau.
- Trạng thái AI đề xuất: candidate; kết quả duyệt nhãn: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E006 — LPS (2026-10-26 → 2026-10-30)

- structure_id: SZC_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Pullback từ high 20.5 ngày 26/10 xuống low 19.1 ngày 30/10; close 19.3 dưới vùng breakout.
- Bằng chứng volume: 29/10 volume 992000, 30/10 912600: chưa cạn cung so với trước breakout.
- Phản ứng đã quan sát đến ngày phân tích: Chưa có phiên hồi xác nhận sau 30/10 trong dữ liệu.
- Điều kiện xác nhận: Chờ giữ vùng pullback, volume giảm và hồi lấy lại 19.588.
- Điều kiện bác bỏ: Nếu tiếp tục giảm và không lấy lại vùng breakout thì không xác nhận LPS.
- Trạng thái AI đề xuất: uncertain; kết quả duyệt nhãn: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E007 — Spring (2026-10-06)

- structure_id: SZC_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 16.2 > đáy TR 16.031; không có xuyên thủng hỗ trợ đã chọn.
- Bằng chứng volume: Volume không thay đổi điều kiện thiếu xuyên đáy.
- Phản ứng đã quan sát đến ngày phân tích: Hồi phục sau phiên không tự biến nó thành Spring.
- Điều kiện xác nhận: Spring phải xuyên hỗ trợ của TR đã xác định rồi phục hồi.
- Điều kiện bác bỏ: Thiếu xuyên 16.031 nên bác bỏ nhãn Spring theo TR này.
- Trạng thái AI đề xuất: invalidated; kết quả duyệt nhãn: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E008 — SC (2026-05-29)

- structure_id: SZC_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close 21.485 gần low 21.438, nhưng tháng 6–7 tạo nhiều đáy thấp hơn.
- Bằng chứng volume: Volume 5023400, RV20 8.92; volume cực lớn riêng lẻ chưa đủ.
- Phản ứng đã quan sát đến ngày phân tích: 08/06 low 21.011; 28/07 low 16.031.
- Điều kiện xác nhận: Cần stopping action, AR và vùng giữ đáy; không chỉ xét volume.
- Điều kiện bác bỏ: Bác bỏ việc dùng 29/05 làm SC của TR tháng 7 hiện tại; không kết luận mọi cách diễn giải lịch sử đều sai.
- Trạng thái AI đề xuất: invalidated; kết quả duyệt nhãn: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Bằng chứng phản đối |
|---|---|---|
| 06/10 volume 293,000 so với 693,400 ngày 27/07 | Ít cung hơn khi kiểm tra vùng đáy; ST ứng viên | RV20 vẫn 1.03, không gọi là cạn cung so với trung bình gần nhất |
| 23/10 close 20.2, volume 1,286,300; RV20 3.41 | Nhịp tăng có kết quả vượt 19.588 | Breakout 22/10 chỉ vượt 0.012; cần xét 23/10 để củng cố |
| 29/10 close 19.4, vị trí đóng cửa 9.1% biên độ | Cung quay lại rõ trên pullback | Chưa đủ dữ liệu để kết luận breakout thất bại hoàn toàn |
| 30/10 low 19.1, close 19.3 | LPS chưa xác nhận | Không có phản ứng hồi sau 30/10 trong dữ liệu |

Không gán Spring cho 06/10 theo đáy TR 16.031. Không ép gán đủ chuỗi PS–SC–AR–ST–Spring–SOS–LPS. Nếu người duyệt chọn ranh giới TR khác, phải đánh giá lại toàn bộ quan hệ giữa các nhãn.

## 7. Kịch bản

- **S1: Giữ được breakout, củng cố pha D.** Xác nhận: Đóng lại trên 19.588 trong ít nhất 2 phiên, pullback sau đó giữ vùng với volume giảm. Bác bỏ: Không lấy lại vùng, hình thành chuỗi đỉnh/đáy thấp hơn.

- **S2: Quay lại dao động trong TR.** Xác nhận: Giao dịch lại dưới 19.446 nhưng vẫn giữ đáy 16.031. Bác bỏ: Thoát vùng bền vững theo một hướng.

- **S3: Giả thuyết tích lũy thất bại.** Xác nhận: Hai phiên liên tiếp đóng dưới 16.031. Bác bỏ: Phục hồi nhanh vào vùng và kiểm tra giữ được đáy.

Các ngưỡng số phiên trên là quy ước minh họa để điều kiện có thể kiểm tra, không phải quy tắc Wyckoff bắt buộc hay ngưỡng đã được backtest.

## 8. Duyệt toàn bộ mẫu

- Trạng thái: in_review; eligible_for_ai: false.
- Người duyệt và ngày duyệt: chưa điền.
- Nhãn đúng đã được người dùng xác nhận: chưa có.
- Nhãn sai AI đề xuất: E007 Spring 06/10; E008 SC 29/05 khi dùng cho TR tháng 7. Chờ anh xác nhận hoặc sửa.
- Cần kiểm tra: SC một phiên hay cụm 27–28/07; vị trí PS; AR kết thúc ngày nào; quan hệ ST muộn; pha D hay B sau 30/10; điều chỉnh giá và nhãn trên ảnh.
- Anh sửa proposed_label/final_label, label_verdict, correction_reason và thông tin reviewer trong annotation.json tương ứng với report.md. Chỉ chuyển approved sau khi duyệt xong.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Chưa có dữ liệu sau 30/10/2026 trong bộ mẫu. Ngày đánh giá, diễn biến thực tế và bài học: để trống; không suy diễn kết quả tương lai.
