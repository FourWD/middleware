package orm

import "github.com/FourWD/middleware/model"

type YearGroup struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel

	Name     string `json:"name" query:"name" gorm:"type:varchar(100)"`
	YearList string `json:"year_list" query:"year_list" gorm:"type:varchar(200)"`
}

/*
view "vehicle_year_groups"
select v.id as vehicle_id, v.year_manufacturing, y.id as year_group_id, y.name as year_group_name from vehicles v
LEFT JOIN year_groups y ON y.year_list LIKE CONCAT('%', v.year_manufacturing ,'%'); */
