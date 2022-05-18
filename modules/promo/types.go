package promo

type Promo struct {
	ID          string `json:"id" db:"id"`                     // *主键* 活动ID
	State       int    `json:"state" db:"state"`               // 活动状态 4=已结束 3=进行中 2=展示中 1=未开始 0=关闭 -1=异常
	Title       string `json:"title" db:"title"`               // 活动标题
	WebContent  string `json:"web_content" db:"web_content"`   // 活动内容web
	H5Content   string `json:"h5_content" db:"h5_content"`     // 活动内容h5
	Prefix      string `json:"prefix" db:"prefix"`             // 前缀
	Flag        string `json:"flag" db:"flag"`                 // 活动类型 deposit=存款活动 rescue=救援金活动 static=静态页活动
	Period      int    `json:"period" db:"period"`             // 活动周期 0=永久有效 1=在指定时间内有效
	Sort        int    `json:"sort" db:"sort"`                 // 展示权重 在列表展示中的权重值 1-255
	StartAt     int64  `json:"start_at" db:"start_at"`         // 活动开始时间 >= ts
	EndAt       int64  `json:"end_at" db:"end_at"`             // 活动结束时间 < ts
	ShowAt      int64  `json:"show_at" db:"show_at"`           // 展示始示时间 >= ts
	CreatedAt   int64  `json:"created_at" db:"created_at"`     // 活动创建时间
	CreatedUID  string `json:"created_uid" db:"created_uid"`   // 活动创建管理员uid
	CreatedName string `json:"created_name" db:"created_name"` // 活动创建管理员账号
	UpdatedAt   int64  `json:"updated_at" db:"updated_at"`     // 活动更新时间
	UpdatedUID  string `json:"updated_uid" db:"updated_uid"`   // 活动更新管理员uid
	UpdatedName string `json:"updated_name" db:"updated_name"` // 活动更新管理员账号
	ApplyTotal  int    `json:"apply_total" db:"apply_total"`   // 每个活动每个用户一共可参与的次数
	ApplyDaily  int    `json:"apply_daily" db:"apply_daily"`   // 每个活动每个用户每天可以参与的次数
	Platforms   string `json:"platforms" db:"platforms"`       // 流水稽查场馆 ["11","12","13","14"...]  ="" 全部场馆生效
	StaticJson  string `db:"static_json"`                      // 静态资源配置 json
	RulesJson   string `db:"rules_json"`                       // 规则配置 json
	ConfigJson  string `db:"config_json"`                      // 基础配置json
}

type PromoJson struct {
	ID         string `json:"id" db:"id"`
	State      int    `json:"state" db:"state"`
	Flag       string `json:"flag" db:"flag"`               //活动标识
	StaticJson string `json:"static_json" db:"static_json"` // 静态资源配置 json
	RulesJson  string `json:"rules_json" db:"rules_json"`   // 规则配置 json
	ConfigJson string `json:"config_json" db:"config_json"` // 基础配置json
}
