// backend/internal/model/article.go
// Article model：文章資料表定義，作為垂直切片範例的資料層

package model

// Article 文章 model
type Article struct {
	Base
	Title   string `gorm:"size:255;not null" json:"title"`
	Content string `gorm:"type:text" json:"content"`
	Tag     string `gorm:"size:64;index" json:"tag"`
}

func init() {
	Register(&Article{})
}
