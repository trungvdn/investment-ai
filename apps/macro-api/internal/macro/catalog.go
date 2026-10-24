package macro

func Catalog() []Indicator {
	return []Indicator{
		{"VN_M2_GROWTH", "Tăng trưởng M2", "Liquidity", "VN", "% YoY", "MONTHLY", "YoY", false, 75, "Chưa tích hợp NHNN; không thay bằng dữ liệu mẫu."},
		{"VN_CREDIT_GROWTH", "Tăng trưởng tín dụng", "Credit", "VN", "% YTD", "MONTHLY", "YTD", false, 75, "Chưa tích hợp NHNN."},
		{"VN_INTERBANK_ON", "Lãi suất liên ngân hàng O/N", "Liquidity", "VN", "%", "DAILY", "LEVEL", false, 7, "Chưa tích hợp NHNN."},
		{"VN_SBV_TBILL", "Tín phiếu NHNN", "Liquidity", "VN", "%", "DAILY", "LEVEL", false, 7, "Chưa xác minh nguồn tín phiếu NHNN; không dùng tín phiếu Kho bạc thay thế."},
		{"VN_USD_VND", "Tỷ giá trung tâm USD/VND", "External", "VN", "VND/USD", "DAILY", "LEVEL", false, 7, "Chưa tích hợp NHNN; tỷ giá trung tâm khác giá giao dịch ngân hàng."},
		{"VN_CPI_YOY", "CPI Việt Nam", "Inflation", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "CPI cả nước tháng so cùng kỳ; SDMX được tính từ cùng chuỗi chỉ số gốc."},
		{"VN_CORE_CPI_YOY", "Lạm phát cơ bản", "Inflation", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "Chỉ nhận số tháng so cùng kỳ, không nhận số bình quân lũy kế."},
		{"VN_PPI_YOY", "PPI công nghiệp", "Inflation", "VN", "% YoY", "QUARTERLY", "YoY", false, 120, "Để giai đoạn tiếp theo; không giả lập số tháng từ số quý."},
		{"VN_PMI", "PMI sản xuất", "Growth", "VN", "index", "MONTHLY", "LEVEL", false, 60, "Chưa tích hợp S&P Global; cần kiểm tra quyền sử dụng."},
		{"VN_IIP_YOY", "Sản xuất công nghiệp", "Growth", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "SDMX AIP_ISIC4_IX; tính YoY từ cùng chuỗi tổng toàn ngành."},
		{"VN_RETAIL_YOY", "Bán lẻ & dịch vụ tiêu dùng", "Growth", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "Tăng trưởng tháng theo giá hiện hành; không phải mức tăng thực đã loại trừ giá."},
		{"VN_EXPORT_YOY", "Xuất khẩu hàng hóa", "External", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "SDMX TXG_FOB_USD, tổng xuất khẩu hàng hóa tháng; tính YoY."},
		{"US_EFFR", "Lãi suất Fed hiệu dụng", "External", "US", "%", "DAILY", "LEVEL", true, 7, "FRED DFF; không phải trần hoặc sàn lãi suất mục tiêu Fed."},
		{"US_BROAD_USD", "Broad USD Index", "External", "US", "index Jan 2006=100", "DAILY", "LEVEL", true, 14, "FRED DTWEXBGS; chỉ số USD rộng, KHÔNG phải ICE DXY."},
		{"US_10Y", "Lợi suất US Treasury 10Y", "External", "US", "%", "DAILY", "LEVEL", true, 7, "FRED DGS10; lợi suất constant maturity 10 năm."},
	}
}
func Find(code string) (Indicator, bool) {
	for _, i := range Catalog() {
		if i.Code == code {
			return i, true
		}
	}
	return Indicator{}, false
}
