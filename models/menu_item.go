package models

//MenuItem type contains menu item info
type MenuItem struct {
	ID       uint64     `gorm:"primary_key" form:"id"`
	Title    string     `form:"title" binding:"required"`
	URL      string     `form:"url"`
	Ord      int        `form:"ord"`
	ParentID *uint64    `form:"parent_id"`
	Children []MenuItem `gorm:"foreignkey:ParentID"`
}

//GetParent returns parent item
func (m *MenuItem) GetParent() MenuItem {
	parent := MenuItem{}
	if m.ParentID != nil {
		db.First(&parent, *m.ParentID)
	}
	return parent
}
