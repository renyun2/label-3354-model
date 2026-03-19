package model

// Article 示例内容模型（演示CRUD接口）
type Article struct {
	BaseModel
	Title    string `gorm:"size:256;not null;comment:标题" json:"title"`
	Content  string `gorm:"type:text;comment:内容" json:"content"`
	AuthorID uint   `gorm:"index;not null;comment:作者ID" json:"author_id"`
	Status   int8   `gorm:"default:1;comment:状态 1发布 0草稿" json:"status"`
	ViewCount int   `gorm:"default:0;comment:浏览次数" json:"view_count"`

	// 关联（查询时预加载）
	Author *User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

func (Article) TableName() string {
	return "articles"
}
