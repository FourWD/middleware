package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type Vehicle struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	SourceID              string    `json:"source_id" query:"source_id" gorm:"type:varchar(36)"`
	SKU                   string    `json:"sku" query:"sku" gorm:"type:varchar(20)"`
	VehicleSubmodelID     string    `json:"vehicle_submodel_id" query:"vehicle_submodel_id" gorm:"type:varchar(36)"`
	VehicleColorID        string    `json:"vehicle_color_id" query:"vehicle_color_id" gorm:"type:varchar(36)"`
	ChassisNo             string    `json:"chassis_no" query:"chassis_no" gorm:"type:varchar(20)"`
	EngineNo              string    `json:"engine_no" query:"engine_no" gorm:"type:varchar(20)"`
	Mile                  int       `json:"mile" query:"mile" gorm:"type:int(11)"`
	YearManufacturing     string    `json:"year_manufacturing" query:"year_manufacturing" gorm:"type:varchar(4)"`
	YearRegister          string    `json:"year_register" query:"year_register" gorm:"type:varchar(4)"`
	License               string    `json:"license" query:"license" gorm:"type:varchar(10)"`
	LicenseProviceID      string    `json:"license_province_id" query:"license_province_id" gorm:"type:varchar(36)"`
	VehicleGradeID        string    `json:"vehicle_grade_id" query:"vehicle_grade_id" gorm:"type:varchar(36)"`
	BranchID              string    `json:"branch_id" query:"branch_id" gorm:"type:varchar(36)"`
	ImagePreviewPath      string    `json:"image_preview_path" query:"image_preview_path" gorm:"type:varchar(400)"`
	CRPPrice              int       `json:"crp_price" query:"crp_price" gorm:"type:int"`
	Remark                string    `json:"remark" query:"remark" gorm:"type:text"`
	IsRecommend           bool      `json:"is_recommend" query:"is_recommend" gorm:"type:bool"`
	//IsSpecialInterest     bool      `json:"is_special_interest" query:"is_special_interest" gorm:"type:bool"`
	EngineCapacity        int       `json:"engine_capacity" query:"engine_capacity" gorm:"type:int"`
	EngineSize            int       `json:"engine_size" query:"engine_size" gorm:"type:int"`
	EngineSizeActual      int       `json:"engine_size_actual" query:"engine_size_actual" gorm:"type:int"`
	VehicleModelID        string    `json:"vehicle_model_id" query:"vehicle_model_id" gorm:"type:varchar(36)"`
	VehicleDriveTypeID    string    `json:"vehicle_drive_type_id" query:"vehicle_drive_type_id" gorm:"type:varchar(2)"`
	VehicleGearID         string    `json:"vehicle_gear_id" query:"vehicle_gear_id" gorm:"type:varchar(2)"`
	VehicleFuelID         string    `json:"vehicle_fuel_id" query:"vehicle_fuel_id" gorm:"type:varchar(2)"`
	Seat                  int       `json:"seat" query:"seat" gorm:"type:int(2)"`
	VehicleTypeID         string    `json:"vehicle_type_id" query:"vehicle_type_id" gorm:"type:varchar(10)"`
	VehicleSubTypeID      string    `json:"vehicle_sub_type_id" query:"vehicle_sub_type_id" gorm:"type:varchar(2)"`
	VehicleBrandID        string    `json:"vehicle_brand_id" query:"vehicle_brand_id" gorm:"type:varchar(36)"`
	RegisterTypeCode      string    `json:"register_type_code" query:"register_type_code" gorm:"type:varchar(36)"`
	RegisterType          string    `json:"register_type" query:"register_type" gorm:"type:varchar(36)"`
	LicenseReceiveDate    time.Time `json:"license_receive_data" query:"license_receive_data" `
	LicenseExpireDate     time.Time `json:"license_expire_data" query:"license_data_expire" `
	TaxExpireDate         time.Time `json:"tax_expire_data" query:"license_data_expire" `
	ContractNumber        string    `json:"contract_number" query:"contract_number" gorm:"type:varchar(20)"`
	ContractDate          time.Time `json:"contract_date" query:"contract_date"`
	//VehiclePickupDate     time.Time `json:"vehicle_pickup_date" query:"vehicle_pickup_date"`
	//VehicleAuctionReceipt string    `json:"vehicle_auction_receipt" query:"vehicle_auction_receipt" gorm:"type:varchar(20)"` // เลขที่ใบรับรถขายทอดตลาด
	Accessories           string    `json:"accessories" query:"accessories" gorm:"type:varchar(500)"`
	PeriodOfUse           int       `json:"period_of_use" query:"period_of_use" gorm:"type:int(2)"`
	Display               bool      `json:"display" query:"display" gorm:"type:bool"`
	AdditionalInfo        string    `json:"additional_info" query:"additional_info" gorm:"type:text"`
	// A3FirstName           string    `json:"a3_first_name" query:"a3_first_name" gorm:"type:varchar(50)"`
	// A3LastName            string    `json:"a3_last_name" query:"a3_last_name" gorm:"type:varchar(50)"`
	// A3CarOwnerNumber      int       `json:"a3_car_owner_number" query:"a3_car_owner_number" gorm:"type:int(2)"`
	// A3AccidentHistory     string    `json:"a3_accident_history" query:"a3_accident_history" gorm:"type:varchar(50)"`
	// A3VehicleLienExists   string    `json:"a3_vehicle_lien_exists" query:"a3_vehicle_lien_exists" gorm:"type:varchar(50)"`
	// IsVehicleManual       bool      `json:"is_vehicle_manual" query:"is_vehicle_manual" gorm:"type:bool"`
	// IsObligate            bool      `json:"is_obligate" query:"is_obligate" gorm:"type:bool"`

	// ImgStrFront        string `json:"img_str_front" query:"img_str_front" gorm:"type:varchar(400)"`
	// ImgStrBack         string `json:"img_str_back" query:"img_str_back" gorm:"type:varchar(400)"`
	// ImgStrRight        string `json:"img_str_right" query:"img_str_right" gorm:"type:varchar(400)"`
	// ImgStrLeft         string `json:"img_str_left" query:"img_str_left" gorm:"type:varchar(400)"`
	// ImgFrontLeft45     string `json:"img_front_left_45" query:"img_front_left_45" gorm:"column:img_front_left_45;type:varchar(400)"`
	// ImgFrontRight45    string `json:"img_front_right_45" query:"img_front_right_45" gorm:"column:img_front_right_45;type:varchar(400)"`
	// ImgBackLeft45      string `json:"img_back_left_45" query:"img_back_left_45" gorm:"column:img_back_left_45;type:varchar(400)"`
	// ImgBackRight45     string `json:"img_back_right_45" query:"img_back_right_45" gorm:"column:img_back_right_45;type:varchar(400)"`
	// ImgInFront         string `json:"img_in_front" query:"img_in_front" gorm:"type:varchar(400)"`
	// ImgInBack          string `json:"img_in_back" query:"img_in_back" gorm:"type:varchar(400)"`
	// ImgConsole         string `json:"img_console" query:"img_console" gorm:"type:varchar(400)"`
	// ImgMileage         string `json:"img_mileage" query:"img_mileage" gorm:"type:varchar(400)"`
	// ImgVehTools        string `json:"img_veh_tools" query:"img_veh_tools" gorm:"type:varchar(400)"`
	// ImgEngineRoom      string `json:"img_engine_room" query:"img_engine_room" gorm:"type:varchar(400)"`
	// ImgGas             string `json:"img_gas" query:"img_gas" gorm:"type:varchar(400)"`
	// ImgOut360          string `json:"img_out_360" query:"img_out_360" gorm:"column:img_out_360;type:varchar(400)"`
	// ImgIn360           string `json:"img_in_360" query:"img_in_360" gorm:"column:img_in_360;type:varchar(400)"`
	// ImgAct             string `json:"img_act" query:"img_act" gorm:"type:varchar(400)"`
	// ImgInsurance       string `json:"img_insurance" query:"img_insurance" gorm:"type:varchar(400)"`
	// ImgInspectionFront string `json:"img_inspection_front" query:"img_inspection_front" gorm:"type:varchar(400)"`
	// ImgInspectionBack  string `json:"img_inspection_back" query:"img_inspection_back" gorm:"type:varchar(400)"`
}

/*
CREATE VIEW `vehicles` AS
SELECT vehicles.* ,
s.name AS source_name,
b.name AS branch_name,
vt.id AS vehicle_type_id1,
vt.name AS vehicle_type_name,
vm.id AS vehicle_model_id1,
vm.name AS vehicle_model_name,
vb.id AS vehicle_brand_id1,
vb.name AS vehicle_brand_name,
vsm.name AS vehicle_submodel_name,
vdt.id AS vehicle_drive_type_id1,
vdt.name AS vehicle_drive_type_name,
vg.name AS vehicle_gear_name,
vft.id AS vehicle_fuel_type_id1,
vft.name AS vehicle_fuel_type_name,
vc.name AS vehicle_color_name,
vgr.name AS vehicle_grade_name,
vgr.row_order AS vehicle_grade_value,
p.name AS license_province_name,
ac.auction_id
FROM (
select * from `auction-vehicle-0001`.vehicles
UNION
select * from `auction-vehicle-0002`.vehicles
UNION
select * from `auction-vehicle-0003`.vehicles
UNION
select * from `auction-vehicle-0004`.vehicles
UNION
select * from `auction-vehicle-0005`.vehicles
UNION
select * from `auction-vehicle-0006`.vehicles
UNION
select * from `auction-vehicle-0007`.vehicles
UNION
select * from `auction-vehicle-0008`.vehicles
UNION
select * from `auction-vehicle-0009`.vehicles
UNION
select * from `auction-vehicle-0010`.vehicles
UNION
select * from `auction-vehicle-0011`.vehicles
UNION
select * from `auction-vehicle-0012`.vehicles
UNION
select * from `auction-vehicle-0013`.vehicles
UNION
select * from `auction-vehicle-0014`.vehicles
UNION
select * from `auction-vehicle-0015`.vehicles
UNION
select * from `auction-vehicle-0016`.vehicles
) AS vehicles
LEFT JOIN auction_vehicles ac ON ac.vehicle_id = vehicles.id
LEFT JOIN vehicle_sub_models vsm ON vsm.id = vehicles.vehicle_submodel_id
LEFT JOIN vehicle_models vm ON vm.id = vsm.vehicle_model_id
LEFT JOIN vehicle_types vt ON vt.id = vm.vehicle_type_id
LEFT JOIN vehicle_brands vb ON vb.id = vm.vehicle_brand_id
LEFT JOIN vehicle_drive_types vdt ON vdt.id = vsm.vehicle_drive_type_id
LEFT JOIN vehicle_gears vg ON vg.id = vsm.vehicle_gear_id
LEFT JOIN vehicle_fuel_types vft ON vft.id = vsm.vehicle_fuel_id
LEFT JOIN vehicle_colors vc ON vc.id = vehicles.vehicle_color_id
LEFT JOIN vehicle_grades vgr ON vgr.id = vehicles.vehicle_grade_id
LEFT JOIN sources s ON s.id = vehicles.source_id
LEFT JOIN branches b ON b.id = vehicles.branch_id
LEFT JOIN provinces p ON p.id = vehicles.license_provice_id;
*/
