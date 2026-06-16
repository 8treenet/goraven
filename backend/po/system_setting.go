package po

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

const (
	ValueTypeString		= "string"
	ValueTypeInt		= "int"
	ValueTypeFloat		= "float"
	ValueTypeBool		= "bool"
	ValueTypeDate		= "date"
	ValueTypeDatetime	= "datetime"
)

type SystemSetting struct {
	ID		int		`gorm:"primaryKey;autoIncrement"`
	Key		string		`gorm:"uniqueIndex;column:config_key;type:varchar(128);not null"`
	Value		string		`gorm:"column:config_value;type:text"`
	ValueType	string		`gorm:"column:value_type;type:varchar(16);default:string"`
	GroupName	string		`gorm:"index;column:group_name;type:varchar(32)"`
	Updated		time.Time	`gorm:"not null;column:updated"`
}

func (s *SystemSetting) TableName() string {
	return "system_setting"
}

func (s *SystemSetting) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

func (s *SystemSetting) BeforeCreate(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

func (s *SystemSetting) Int() int {
	v, _ := strconv.Atoi(s.Value)
	return v
}

func (s *SystemSetting) Bool() bool {
	v, _ := strconv.ParseBool(s.Value)
	return v
}
