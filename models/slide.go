package models

//Slide type contains carousel slide info
type Slide struct {
	ID            uint64 `gorm:"primary_key" form:"id"`
	Title         string `form:"title" binding:"required"`
	NavigationURL string `form:"navigation_url"`
	FileURL       string `form:"file_url"`
	Ord           int    `form:"ord"`
}
