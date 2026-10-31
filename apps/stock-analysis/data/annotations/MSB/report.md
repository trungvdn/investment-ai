# MSB — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn, chờ người dùng duyệt. review.status=in_review; eligible_for_ai=false. Nhận định từ OHLCV; chưa đối chiếu nhãn trên ảnh.

## 1. Thông tin và dữ liệu

- Khung 1D; cutoff 2026-10-30.
- [candles_MSB_1D.csv](candles_MSB_1D.csv): 497 phiên từ 2024-10-31 đến 2026-10-30; cột event trống.
- [MSB_nhan.png](MSB_nhan.png): ảnh để người duyệt đối chiếu.
- source_commit: 2ed467825a5c943f946304ecd0ef712d6c8fc02d.
- Close cuối 15.8 theo đơn vị CSV; chưa xác minh đơn vị/điều chỉnh giá, lịch giao dịch.
- RV20=volume/trung bình 20 phiên liền trước, loại phiên hiện tại.
- Cần xác minh chính sách điều chỉnh quanh 26–28/08 trước khi duyệt nhãn Spring; không suy đoán nguyên nhân doanh nghiệp.

## 2. Kết luận nhanh

**Ưu tiên giả thuyết tái tích lũy đã thoát vùng, pha E đang phát triển; tin cậy trung bình định tính.** Cụm 26–27/08 là Spring ứng viên, 23/09 là SOS và 05/10 là LPS ứng viên. Giá sau đó tạo đỉnh cao hơn tới high 16.15 ngày 29/10. Phiên giảm 30/10 chưa đủ để kết luận phân phối, đồng thời chưa đủ xác nhận một LPS mới.

## 3. Bối cảnh

- Close 09/03=9.011, 30/06=13.517; trước TR là nhịp tăng, phù hợp giả thuyết tái tích lũy.\n- Vùng đi ngang tháng 7–8 có đáy 22/07=12.808 và đỉnh 15–17/07=13.726.\n- 30/10 close 15.8 cao hơn vùng cũ; không có bằng chứng buộc phải gán phân phối chỉ vì một phiên giảm.

Không ép đủ SC–AR–ST: vùng tái tích lũy sau tăng không nhất thiết có Selling Climax như vùng tích lũy sau markdown. Chưa có dữ liệu ngành/vĩ mô trong mẫu để kết luận thêm.

## 4. Trading Range

| Thành phần | Kết quả | Cơ sở |
|---|---|---|
| structure_id | MSB_TR_202607 | Vùng tháng 7–8, kiểm tra tháng 9 |
| Bắt đầu | 01/07/2026 | Đi ngang sau nhịp tăng tháng 3–6 |
| Hỗ trợ trước xuyên đáy | 12.808–12.975 | Low 22/07 và các phản ứng tháng 8 |
| Kháng cự | 13.601–13.726 | Đỉnh 15–17/07, kiểm tra tháng 8 |
| Đáy Spring ứng viên | 12.14 | Cụm 26–27/08; không hồi tố sửa biên cũ |
| Breakout ứng viên | 23/09 | Close 14.15 trên 13.726 |

TR vẫn provisional theo trạng thái duyệt; end_date trong JSON là ngày breakout ứng viên, không xác nhận cấu trúc đã được người dùng duyệt.

## 5. Nhật ký nhãn sự kiện

Mọi label_verdict=unreviewed, final_label=null. invalidated là đề xuất bác bỏ của AI. Sử dụng kiến thức đến cutoff, không giả vờ biết phản ứng sau sự kiện ngay tại ngày xảy ra.

### E001 — Spring: 2026-08-26 → 2026-08-27

- structure_id: MSB_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 26/08 low 12.14, close 12.808; 27/08 low 12.14, close=high=13.1 trở lại trên 12.808–12.975.
- Bằng chứng volume: 26/08 volume 8724400 RV20 1.56; 27/08 volume 13851000 RV20 2.37.
- Phản ứng đã quan sát: 28/08 close 13.35; kiểm tra tháng 9 không xuyên lại 12.14.
- Bằng chứng phản đối/giới hạn: Volume cao ở phiên phục hồi; đây không phải ví dụ Spring volume thấp. Điều chỉnh giá chưa được xác minh.
- Điều kiện xác nhận: Xuyên hỗ trợ đã hình thành, hồi lại vùng và giữ đáy; dữ liệu cần cùng cơ sở điều chỉnh.
- Điều kiện bác bỏ: Nếu biến động do điều chỉnh dữ liệu chưa xử lý hoặc giá chấp nhận dưới 12.14 bền vững, xem lại nhãn Spring.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E002 — Test: 2026-09-15

- structure_id: MSB_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 12.65 cao hơn Spring low 12.14 khoảng 4.2%; close 12.8 hơi dưới biên 12.808.
- Bằng chứng volume: Volume 2244600 RV20 0.44, bằng 16.2% volume 27/08.
- Phản ứng đã quan sát: 16–22/09 hồi và 23/09 breakout.
- Bằng chứng phản đối/giới hạn: Cách Spring gần ba tuần; kiểm tra không sát đáy và close chưa lấy lại hẳn biên cũ.
- Điều kiện xác nhận: Cần đánh giá quan hệ với Spring và cung giảm; có thể chỉ là kiểm tra bên trong.
- Điều kiện bác bỏ: Nếu không chấp nhận khoảng cách/thời gian hoặc hành vi giữ hỗ trợ, bỏ nhãn Test thay vì ép đủ chuỗi.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E003 — SOS: 2026-09-23

- structure_id: MSB_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=13.55/14.35/13.4/14.15; close vượt 13.726 khoảng 3.09%. 22/09 close 13.5 vẫn dưới kháng cự; không gán breakout vùng lớn ở phiên này.
- Bằng chứng volume: 23/09 volume 28997100 RV20 5.52; 22/09 volume 19351700 RV20 4.35.
- Phản ứng đã quan sát: 24/09 close 14.25; pullback 05/10 low 13.85 vẫn trên 13.726; sau đó giá tiếp tục tăng.
- Bằng chứng phản đối/giới hạn: Một phiên volume cao riêng lẻ không đủ; đánh giá kết hợp phản ứng sau breakout.
- Điều kiện xác nhận: Thoát vùng rõ, follow-through và kiểm tra giữ kháng cự cũ.
- Điều kiện bác bỏ: Quay bền dưới 13.601 và không lấy lại 13.726 làm suy yếu SOS thành công.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E004 — LPS: 2026-10-05

- structure_id: MSB_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 13.85, close 13.9 vẫn trên biên breakout 13.726.
- Bằng chứng volume: Volume 7284400 RV20 0.93, bằng 25.1% volume SOS 23/09.
- Phản ứng đã quan sát: 06/10 close 14.4; 07/10 14.65; 08/10 14.75, sau đó tạo đỉnh cao hơn.
- Bằng chứng phản đối/giới hạn: Volume chỉ hơi thấp hơn trung bình 20 phiên, không gọi cạn cung hoàn toàn.
- Điều kiện xác nhận: Pullback sau SOS giữ vùng và có phản ứng cầu; không cần gán LPS trước SOS.
- Điều kiện bác bỏ: Nếu chuỗi sau đó mất vùng breakout bền vững, đánh giá lại LPS thành công.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E005 — SC: 2026-07-22

- structure_id: MSB_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 12.808, close 13.267; trước đó là xu hướng tăng và đi ngang, không phải kết thúc markdown rõ.
- Bằng chứng volume: Volume 21176000 RV20 3.03; volume cao không đủ tạo SC.
- Phản ứng đã quan sát: Giá tiếp tục dao động; tháng 8 xuyên thấp hơn 12.808.
- Bằng chứng phản đối/giới hạn: Có thể là shakeout/kiểm tra cung trong vùng, chưa gán nhãn thay thế.
- Điều kiện xác nhận: SC cần bối cảnh dừng xu hướng giảm của đúng cấu trúc, không chỉ phiên giảm volume cao.
- Điều kiện bác bỏ: Đề xuất bác bỏ SC theo giả thuyết tái tích lũy đang xét; không khẳng định mọi phân tích khung khác đều sai.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E006 — SOW: 2026-08-26

- structure_id: MSB_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 12.14 dưới hỗ trợ nhưng close 12.808, ngày 27/08 phục hồi close 13.1.
- Bằng chứng volume: 26/08 RV20 1.56; 27/08 volume lớn hơn và close ở high.
- Phản ứng đã quan sát: 28/08 close 13.35; sau đó breakout trên vùng tháng 9.
- Bằng chứng phản đối/giới hạn: Một lần xuyên hỗ trợ chưa đủ chứng minh breakdown được chấp nhận; cần xem phản ứng hồi.
- Điều kiện xác nhận: SOW cần suy yếu có bối cảnh phân phối và follow-through phù hợp.
- Điều kiện bác bỏ: Phục hồi nhanh vào vùng cùng diễn biến sau đó phản đối nhãn SOW thành công ở cutoff; ưu tiên Spring ứng viên E001.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E007 — LPS: 2026-10-30

- structure_id: MSB_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=16/16.1/15.75/15.8; đây là pullback sau chuỗi đỉnh mới, xa biên TR cũ.
- Bằng chứng volume: Volume 10312400 cao hơn 29/10=5863300; chưa phải pullback có volume giảm so phiên liền trước.
- Phản ứng đã quan sát: Không có phiên sau 30/10 trong bộ mẫu.
- Bằng chứng phản đối/giới hạn: Không dùng giá ở cao hơn vùng cũ để tự xác nhận mọi pullback là LPS.
- Điều kiện xác nhận: Cần vùng hỗ trợ thích hợp và phản ứng hồi xác nhận; nếu là cấu trúc mới phải tạo structure_id riêng.
- Điều kiện bác bỏ: Tiếp tục giảm không có cầu hỗ trợ thì không xác nhận LPS mới.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Giới hạn |
|---|---|---|
| 27/08 close=high=13.1, RV20 2.37 | Phục hồi sau xuyên đáy | Không gọi Spring volume thấp; điều chỉnh giá cần xác minh |
| 15/09 volume 2,244,600, RV20 0.44 | Ít nỗ lực bán hơn cụm xuyên đáy | Test còn uncertain về vị trí/thời gian |
| 23/09 RV20 5.52, close vượt 13.726 | Nỗ lực tăng có kết quả thoát vùng | Cần phản ứng sau breakout, không xét riêng volume |
| 05/10 low 13.85 giữ trên biên cũ, sau đó tăng | LPS ứng viên được phản ứng sau hỗ trợ | Không chứng minh danh tính/ý định tổ chức |
| 30/10 close 15.8, volume tăng so 29/10 | Có cung trên nhịp điều chỉnh | Không đủ kết luận phân phối hoặc LPS mới |

Pha E là cách diễn giải thoát vùng cũ tại cutoff; không có nghĩa xu hướng tăng sẽ tiếp tục chắc chắn. Nếu trở lại bền trong vùng, phải cập nhật nhận định.

## 7. Kịch bản

- **S1 — Tiếp diễn markup sau tái tích lũy.** structure_id: MSB_TR_202607. Xác nhận: Giữ được các đáy pullback mới và vượt 16.15 với follow-through; kiểm tra lại có cầu hỗ trợ. Bác bỏ: Chuỗi đỉnh/đáy thấp hơn và mất hỗ trợ breakout; không coi một phiên giảm là đủ bác bỏ. Bằng chứng nền: E003, E004, E007.

- **S2 — Điều chỉnh sau tăng, chưa mất vùng cũ.** structure_id: MSB_TR_202607. Xác nhận: Giá giảm/đi ngang nhưng vẫn giữ 13.601–13.726 khi kiểm tra; theo dõi chất lượng cung–cầu trước khi đặt vùng mới. Bác bỏ: Tiếp tục tăng có follow-through hoặc mất bền vùng breakout. Bằng chứng nền: E004, E007.

- **S3 — Breakout thất bại và quay vào TR.** structure_id: MSB_TR_202607. Xác nhận: Hai close liên tiếp dưới 13.601, không lấy lại 13.726 trong ba phiên sau; hồi bị từ chối củng cố. Bác bỏ: Nhanh lấy lại biên trên và kiểm tra giữ được. Mất 12.14 bền vững sẽ làm xấu hơn giả thuyết Spring. Bằng chứng nền: E001, E003.

Ngưỡng hai/ba phiên là quy ước minh họa, chưa backtest; các kịch bản tương lai chưa xảy ra tại cutoff.

## 8. Duyệt mẫu

- Trạng thái: in_review; eligible_for_ai=false.
- Người duyệt/ngày duyệt: chưa điền. Chưa có nhãn đúng/sai được người dùng xác nhận.
- Cần duyệt: bối cảnh tái tích lũy, biên TR, dữ liệu điều chỉnh quanh tháng 8, Spring, Test, SOS, LPS và pha E.
- Đề xuất bác bỏ: SC 22/07 do thiếu bối cảnh kết thúc markdown; SOW 26/08 do phục hồi nhanh. LPS mới 30/10 còn uncertain.
- Sửa final_label, label_verdict, correction_reason, reviewer trong annotation.json cùng báo cáo; chỉ approved sau khi người dùng xác nhận.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Không có dữ liệu sau 30/10 trong bộ mẫu; evaluation_date=null và observations=[].
