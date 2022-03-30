package persist

import (
	"time"
)

type Object interface{}

type Table interface {
	TblInsert(obj *Object) error
	TblDelete(obj *Object) error
	TblUpdate(obj *Object) error
	TblLookup(obj *Object) error
}

type BaseModel struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Tunnel tunnel url table
type Tunnel struct {
	ID          uint     `gorm:"column:id;primaryKey"`
	FkCid       string   `gorm:"column:fk_cid;index:idx_cid"`
	URLCgi      string   `gorm:"column:url_cgi;type:varchar(128)" json:"url_cgi" `
}

// Device device info table
type Device struct {
	//gorm.Model
	//BaseModel
	//ID          uint     `gorm:"column:id;primaryKey" json:"id,omitempty"` // test
	Cid         string   `gorm:"column:cid;index:idx_sn;primaryKey"  json:"client_id,omitempty"`
	SN          string   `gorm:"column:sn" json:"sn,omitempty"`
	Vendor      string   `gorm:"column:vendor;index:idx_vendor" json:"vendor,omitempty"`
	Mode        string   `gorm:"column:mode" json:"mode,omitempty"`
	SWVersion   string   `gorm:"column:version" json:"sw_version,omitempty"`
	MAC         string   `gorm:"column:mac" json:"mac,omitempty"`
	Addr        string   `gorm:"column:addr" json:"addr,omitempty"`
	Attribution string   `gorm:"column:attribution" json:"attribution,omitempty"`
	State       uint     `gorm:"column:state;index:idx_state" json:"state,omitempty"` // 设备状态：0表示在线，>0表示离线天数
	CreateTime  string   `gorm:"column:create_time" json:"create_time,omitempty"`     // first login time
	LoginTime   string   `gorm:"column:login_time" json:"login_time,omitempty"`       // login time
	IsHole      bool     `gorm:"column:is_hole" json:"is_hole,omitempty"`
	Tunnels     []Tunnel `gorm:"foreignKey:FkCid;references:Cid;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"tunnels,omitempty"` // 以DevSN作为从表Tunnel的外键，引用主表Device的SN字段
}

type Vendor struct {
	Name   string
	DevNum uint32
}

type Area struct {
	Province string
	Country  string
	DevNum   uint32
}

type User struct {
	Name   string
	Passwd string
	Status bool
}

// AreaA 表示一个国家设备地理分布统计
type AreaA struct {
	Country string
	// 国家设备总数 = 其所有省份设备总和；key:value => "Changsha":125
	Province map[string]uint32
}

// Summarization 对设备厂商和设备地理分布汇总
type Summarization struct {
	VendorSet map[string]int  `json:"vendorSet"` // 方便排序，考虑链表
	AreaSet   map[string]Area `json:"areaSet"`
}

type SysResource struct {
	MemTotal  int32 `json:"memTotal"`
	MemUsed   int32 `json:"memUsed"`
	DiskTotal int32 `json:"diskTotal"`
	DiskUsed  int32 `json:"diskUsed"`
}
