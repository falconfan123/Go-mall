package valueobject

// AddressInfo 地址详情值对象
type AddressInfo struct {
	Province string // 省份
	City     string // 城市
	District string // 区域
	Detail   string // 详细地址
	ZipCode  string // 邮编
}

// NewAddressInfo 创建地址详情
func NewAddressInfo(province, city, district, detail, zipCode string) *AddressInfo {
	return &AddressInfo{
		Province: province,
		City:     city,
		District: district,
		Detail:   detail,
		ZipCode:  zipCode,
	}
}

// IsEmpty 判断地址是否为空
func (a AddressInfo) IsEmpty() bool {
	return a.Province == "" || a.City == "" || a.District == "" || a.Detail == ""
}

// Equals 比较两个地址是否相同
func (a AddressInfo) Equals(other AddressInfo) bool {
	return a.Province == other.Province &&
		a.City == other.City &&
		a.District == other.District &&
		a.Detail == other.Detail &&
		a.ZipCode == other.ZipCode
}

// FullAddress 获取完整地址字符串
func (a AddressInfo) FullAddress() string {
	return a.Province + a.City + a.District + a.Detail
}
