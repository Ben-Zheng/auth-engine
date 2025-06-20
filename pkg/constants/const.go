package constants

const (
	ProductName            = "auth-engine" // ProductName 表示产品名称
	DefaultPageSize        = 10            // DefaultPageSize 表示默认的每页数量
	AesCryptionKey         = "53436832ecc4eeda3769c19400395208"
	ImportTokenFileMaxSize = 1024 * 1024 * 512 // 512MB
	DaliyPolicyType        = "DAILY"
	WeeklyPolicyType       = "WEEKLY"
	DateRangePolicyType    = "DATERANGE"
	DaliyTimeFormat        = "15:04:05"   // time.Now().Format("08:00:00")
	DateRangeTimeFormat    = "2006-01-02" // time.Now().Format("2006-01-02")
	TimeFormat             = "2006-01-02 15:04:05"
	MONDAY                 = "MONDAY" // time.Now().Weekday()
	TUESDAY                = "TUESDAY"
	WEDNESDAY              = "WEDNESDAY"
	THURSDAY               = "THURSDAY"
	FRIDAY                 = "FRIDAY"
	SATURDAY               = "SATURDAY"
	SUNDAY                 = "SUNDAY"
)

var (
	WeeklyDayMap = map[string]int{
		"MONDAY":    1,
		"TUESDAY":   2,
		"WEDNESDAY": 3,
		"THURSDAY":  4,
		"FRIDAY":    5,
		"SATURDAY":  6,
		"SUNDAY":    7,
	}
)
