# NVL — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn, chờ người dùng duyệt. review.status=in_review; eligible_for_ai=false. Phân tích từ OHLCV, chưa đối chiếu nhãn trên ảnh.

## 1. Thông tin và dữ liệu

- Khung: 1D; cutoff: 2026-10-30.
- [candles_NVL_1D.csv](candles_NVL_1D.csv): 497 phiên từ 31/10/2024 đến 30/10/2026.
- [NVL_nhan.png](NVL_nhan.png): ảnh tham chiếu để người duyệt đối chiếu.
- source_commit: d74161d9e6b549236f55eb9419cbf32879e1653a.
- Close cuối 10.4 theo đơn vị CSV; chưa xác minh đơn vị và điều chỉnh giá.
- Không trùng ngày; OHLC hợp lệ, volume không âm. Chưa xác minh đủ phiên theo lịch thị trường. Cột event trống.
- RV20 = volume / trung bình 20 phiên liền trước, loại phiên hiện tại. CLV=(close-low)/(high-low); không tính khi high=low như 29/09.

## 2. Kết luận nhanh

**Cấu trúc chưa rõ; ưu tiên giả thuyết dừng giảm và tích lũy sớm ở pha A, mức tin cậy thấp.** SC 30/09 và AR đến 12–13/10 là các nhãn ứng viên. Pullback cuối tháng 10 có thể là ST nhưng còn thiếu hồi xác nhận. Chưa có SOS hay LPS hợp lệ của vùng đang xét. Không sao chép kết luận pha D từ mẫu SZC.

## 3. Bối cảnh

- High 10/08/2026 là 14.181; close 30/09 là 10.2.\n- Vùng đáy cũ 11.394 ngày 16/07 và 11.593 ngày 14/09 không giữ được: 28/09 close 10.996, 29/09 close 10.25.\n- Phân tích tách vùng tháng 9–10 khỏi cấu trúc tháng 7–8; không nối ST tháng 10 vào đáy cũ đã mất.

Chưa có dữ liệu ngành, vĩ mô hoặc nguyên nhân doanh nghiệp trong bộ mẫu; không suy đoán nguyên nhân giảm giá. Vùng mới có thể là tích lũy, cũng có thể chỉ là nghỉ trong markdown.

## 4. Trading Range ứng viên

| Thành phần | Vùng/ngày | Cơ sở |
|---|---|---|
| structure_id | NVL_TR_202609 | Tách biệt khỏi vùng tháng 7–8 |
| Bắt đầu | 30/09/2026 | SC ứng viên sau breakdown |
| Hỗ trợ | 9.6–10.05 | Low 30/09=9.6; 02/10=9.91; 08–09/10=10.05 |
| Kháng cự | 11.15–11.3 | Close và high AR 12–13/10 |
| Vùng kiểm tra bên trong | 10.15–10.25 | Pullback cuối tháng 10; không thay đáy TR |
| Kết thúc | Chưa xác định | Chưa thoát vùng bền vững |

## 5. Nhật ký nhãn sự kiện

Mọi label_verdict đều unreviewed, final_label=null. Trạng thái invalidated là đề xuất bác bỏ của AI, chưa phải kết quả người dùng duyệt. Dùng dữ liệu đến cutoff; không coi kiến thức có ở 30/10 là đã biết tại từng phiên sự kiện.

### E001 — PS: 2026-09-14

- structure_id: NVL_TR_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 11.593; close 11.892; sau đó hồi đến high 13.136 ngày 21/09.
- Bằng chứng volume: Volume 22465000, RV20 2.00.
- Phản ứng đã quan sát: Có phản ứng cầu tạm thời trước nhịp bán mạnh cuối tháng 9.
- Bằng chứng phản đối/giới hạn: Chưa chắc có quan hệ PS–SC; có thể chỉ là hồi kỹ thuật trong vùng cũ.
- Điều kiện xác nhận: Cần đánh giá thay đổi hành vi từ giữa đến cuối tháng 9.
- Điều kiện bác bỏ: Nếu không có vai trò hỗ trợ sơ bộ trong cụm dừng giảm thì bỏ PS.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E002 — SC: 2026-09-30

- structure_id: NVL_TR_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=9.6/10.6/9.6/10.2; close ở 60% biên độ; tạo đáy mới sau các phiên giảm 28–29/09.
- Bằng chứng volume: 97059000; trung bình 20 phiên trước 13935145; RV20 6.97.
- Phản ứng đã quan sát: Sau đó low 02/10=9.91 chưa phá 9.6; hồi tới 11.3 ngày 12–13/10.
- Bằng chứng phản đối/giới hạn: Volume lớn có thể phản ánh bán cưỡng bức; OHLCV không xác định được nguyên nhân hay người mua.
- Điều kiện xác nhận: Cần AR và giữ vùng đáy qua các lần kiểm tra để củng cố giả thuyết SC.
- Điều kiện bác bỏ: Nếu markdown tiếp diễn bền dưới 9.6 thì xem lại SC của vùng hiện tại.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E003 — AR: 2026-10-01 → 2026-10-13

- structure_id: NVL_TR_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Nhịp hồi sau 30/09: close 09/10=10.95; high 12–13/10=11.3; close hai phiên=11.15.
- Bằng chứng volume: 09/10 volume 73096900 RV20 3.12; 12/10=18672700; 13/10=21625300.
- Phản ứng đã quan sát: Từ 14/10 điều chỉnh; close 30/10=10.4.
- Bằng chứng phản đối/giới hạn: Nhịp hồi kéo dài nhiều phiên; người duyệt cần xác nhận phạm vi AR phù hợp.
- Điều kiện xác nhận: Hồi sau SC rồi xuất hiện pullback tạo cạnh trên tham chiếu.
- Điều kiện bác bỏ: Nếu người duyệt chọn cấu trúc khác, phải xác định lại AR và cạnh trên.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E004 — ST: 2026-10-26 → 2026-10-30

- structure_id: NVL_TR_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 26/10=10.2; low 30/10=10.15, vẫn cao hơn đáy SC 9.6 khoảng 5.73%.
- Bằng chứng volume: 26/10=28763200 RV20 1.03; 30/10=14738700 RV20 0.61; bằng 15.2% volume SC.
- Phản ứng đã quan sát: Chưa có dữ liệu hồi sau 30/10.
- Bằng chứng phản đối/giới hạn: Đây là kiểm tra hỗ trợ bên trong; chưa quay sát đáy 9.6 và chưa có hồi xác nhận.
- Điều kiện xác nhận: Chờ giữ đáy pullback và hồi phục; đối chiếu test bên trong hay ST của SC.
- Điều kiện bác bỏ: Nếu giá tiếp tục giảm không có phản ứng cầu thì không xác nhận ST; thủng 9.6 cần đánh giá lại vùng.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E005 — SOS: 2026-10-09

- structure_id: NVL_TR_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close=high=10.95, vượt high 30/09=10.6 nhưng chưa vượt biên trên AR hình thành sau đó là 11.3.
- Bằng chứng volume: 73096900; RV20 3.12; volume lớn riêng lẻ không đủ xác nhận SOS.
- Phản ứng đã quan sát: 12–13/10 high 11.3, sau đó quay xuống; đến 30/10 chưa có close trên 11.3.
- Bằng chứng phản đối/giới hạn: Không dùng đỉnh tương lai để giả vờ biết biên AR tại 09/10: đây là đánh giá hồi cứu ở cutoff 30/10.
- Điều kiện xác nhận: Muốn gán SOS cho vùng đang xét cần breakout rõ khỏi TR đã hình thành và follow-through.
- Điều kiện bác bỏ: Tại cutoff, bác bỏ gán SOS 09/10 cho NVL_TR_202609; ưu tiên xem đây là một phần AR.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E006 — Spring: 2026-10-02

- structure_id: NVL_TR_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 9.91 cao hơn hỗ trợ đáy TR 9.6; không xuyên đáy SC.
- Bằng chứng volume: 40823300 RV20 2.22; volume không bù được điều kiện thiếu xuyên hỗ trợ.
- Phản ứng đã quan sát: Giá hồi sau phiên, nhưng không biến phiên thành Spring theo đáy TR đã chọn.
- Bằng chứng phản đối/giới hạn: Nếu dùng một vùng nhỏ khác cần structure_id riêng.
- Điều kiện xác nhận: Phải xác định TR trước và có xuyên hỗ trợ rồi phục hồi.
- Điều kiện bác bỏ: Bác bỏ Spring 02/10 đối với vùng có đáy 9.6.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E007 — LPS: 2026-10-26 → 2026-10-30

- structure_id: NVL_TR_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Pullback về low 10.15 ở dưới đỉnh AR 11.3; chưa có SOS hợp lệ trước đó.
- Bằng chứng volume: Volume giảm so SC nhưng giảm volume không tự tạo nhãn LPS.
- Phản ứng đã quan sát: Chưa có dữ liệu sau 30/10 để xác nhận hồi phục.
- Bằng chứng phản đối/giới hạn: Có thể là ST ứng viên theo E004, không đồng thời gán LPS cho cùng cấu trúc.
- Điều kiện xác nhận: Cần SOS trước, sau đó pullback giữ breakout và có bằng chứng cầu.
- Điều kiện bác bỏ: Bác bỏ LPS hiện tại do thiếu SOS của NVL_TR_202609.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Giới hạn |
|---|---|---|
| 30/09 volume 97,059,000; RV20 6.97; CLV 60% | Bán mạnh và có phản ứng cầu trong phiên | Không chứng minh bên mua là tổ chức hoặc cung đã hết |
| 09/10 close ở high, RV20 3.12 | Nhịp hồi mạnh, thuộc AR ứng viên | Không tự suy ra SOS của vùng lớn |
| 30/10 volume 14,738,700; RV20 0.61 | Nỗ lực bán giảm so SC và trung bình gần nhất | Có thể thiếu cầu; cần hồi xác nhận ST |

Pha A là giả thuyết làm việc; không khẳng định đã hoàn tất pha A hoặc vào pha B. Cần kiểm tra đáy có chất lượng trước khi nâng mức tin cậy. Pha D chưa được hỗ trợ vì thiếu breakout khỏi 11.3.

## 7. Kịch bản

- **S1: Củng cố tích lũy sớm.** Xác nhận: Giữ vùng 10.15–10.25, có nhịp hồi và kiểm tra lại với cung giảm; sau đó mới xem xét phá 11.3. Bác bỏ: Mất đáy pullback liên tục và không có hồi; thủng bền 9.6 bác bỏ vùng hiện tại.

- **S2: Dao động trong vùng, chưa rõ cấu trúc.** Xác nhận: Giá tiếp tục giao dịch giữa 9.6 và 11.3, chưa có breakout có follow-through. Bác bỏ: Thoát vùng bền vững theo một hướng.

- **S3: Markdown tiếp diễn hoặc tái phân phối.** Xác nhận: Hai close liên tiếp dưới 9.6 và không phục hồi trong ba phiên tiếp theo. Bác bỏ: Nhanh chóng lấy lại vùng và giữ được hỗ trợ khi kiểm tra.

Ngưỡng hai/ba phiên là quy ước vận hành minh họa, không phải quy tắc Wyckoff bắt buộc hoặc ngưỡng đã backtest.

## 8. Duyệt mẫu

- Trạng thái: in_review; eligible_for_ai=false.
- Người duyệt/ngày duyệt: chưa điền. Chưa có nhãn đúng hoặc sai được con người xác nhận.
- Cần duyệt: SC 30/09; phạm vi AR 01–13/10; PS 14/09 có nên bỏ; ST cuối tháng 10 hay chưa đủ điều kiện.
- Đề xuất bác bỏ: SOS 09/10 (một phần AR), Spring 02/10 (không xuyên 9.6), LPS cuối tháng 10 (thiếu SOS trước đó).
- Kiểm tra nhãn trên ảnh và chính sách điều chỉnh giá; sửa final_label, label_verdict, correction_reason và reviewer trong annotation.json cùng báo cáo.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Chưa có dữ liệu sau 30/10 trong bộ mẫu. Giữ evaluation_date=null và observations=[]; không giả định kịch bản đã xảy ra.
