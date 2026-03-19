package model

// User 用户表
type User struct {
	BaseModel
	OpenID    string  `gorm:"uniqueIndex;size:64;not null;comment:微信openid" json:"open_id"`
	UnionID   string  `gorm:"index;size:64;comment:微信unionid" json:"union_id,omitempty"`
	Nickname  string  `gorm:"size:64;comment:昵称" json:"nickname"`
	AvatarURL string  `gorm:"size:512;comment:头像URL" json:"avatar_url,omitempty"`
	// Phone 使用指针类型（SQL NULL），uniqueIndex 对 NULL 不做唯一性约束，
	// 允许多个未绑定手机号的用户，同时保证已绑定用户的手机号全局唯一。
	Phone     *string `gorm:"uniqueIndex;size:20;comment:手机号" json:"phone,omitempty"`
	Gender    int8    `gorm:"default:0;comment:性别 0未知 1男 2女" json:"gender"`
	Status    int8    `gorm:"default:1;comment:状态 1正常 2禁用" json:"status"`
	Role      string  `gorm:"size:32;default:user;comment:角色 user|admin" json:"role"`
	LastLogin int64   `gorm:"comment:最后登录时间戳" json:"last_login,omitempty"`

	Sessions []UserSession `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// UserSession 用户会话表（存储微信session_key等）
type UserSession struct {
	BaseModel
	UserID     uint   `gorm:"index;not null;comment:用户ID" json:"user_id"`
	SessionKey string `gorm:"size:128;comment:微信session_key" json:"-"`
	Platform   string `gorm:"size:32;default:miniprogram;comment:平台" json:"platform"`
	DeviceInfo string `gorm:"size:256;comment:设备信息" json:"device_info,omitempty"`
	ClientIP   string `gorm:"size:64;comment:客户端IP" json:"client_ip,omitempty"`
	ExpiredAt  int64  `gorm:"comment:过期时间戳" json:"expired_at"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}
