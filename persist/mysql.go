package persist

import (
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"pro_mysql/system"
	"time"
)

type ObjTable struct {
	DbName  string
	TblName string
	Obj     *Object
}

type SqlDb struct {
	gormDb *gorm.DB
	user   string
	passwd string
	host   string
	port   int
	dbname string
}

const (
	PkgName  string = "[persist]"
)

//func init() {
func OpenDb() (gormDb *gorm.DB, err error) {
	dbConfig := system.GetDbConfiguration()

	sqlCfg := mysql.Config{
		DSN:                       dbConfig.DSN,
		DefaultStringSize:         128,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}
	gormConfig := gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: false, // 此处禁止AutoMigrate自动创建数据库外键约束
	}

	gormDb, err = gorm.Open(mysql.New(sqlCfg), &gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDb, err := gormDb.DB()
	if err != nil {
		return nil, err
	}

	sqlDb.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDb.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDb.SetConnMaxLifetime(time.Minute)

	err = sqlDb.Ping()
	if err != nil {
		return nil, err
	}

	return gormDb, err
}

func CloseDB(gormDb *gorm.DB) {
	sqlDb, err := gormDb.DB()
	if err != nil {
		return
	}

	_ = sqlDb.Close() // 关闭链接
}

func InitDb(db *gorm.DB) error {
	para := "ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='device info'"
	return db.Set("gorm:table_options", para).AutoMigrate(&Device{}, &Tunnel{})
}

// InsertDevice 插入一行device数据
func (dev *Device) InsertDevice(db *gorm.DB) error {
	//如果设备sn冲突，该设备信息已经存在，不做任何刷新;如果要刷新，请调用刷新接口
	conflict := clause.OnConflict {
		Columns: []clause.Column{{Name: "sn"}},
		DoNothing: true,
	}
	return insertTable(db, dev, &conflict)
}

// AppendDeviceAssociation 添加dev的关联数据。
// 如果主表中不存在dev，则其关联数据添加失败（设置了约束：OnDelete:CASCADE）
func (dev *Device) AppendDeviceAssociation(db *gorm.DB) error {
	for _, tunnel := range dev.Tunnels {
		if err := appendAssociation(db, dev, "Tunnels", &tunnel); err != nil {
			return err
		}
	}
	return nil
}

// DeletedDeviceAssociation 删除设备的关联数据。
func (dev *Device) DeletedDeviceAssociation(db *gorm.DB) error {
	if 0 == len(dev.Tunnels) {
		return errors.New(PkgName + "condition is nil")
	}

	var err error = nil
	size := len(dev.Tunnels)
	tunnels := dev.Tunnels
	for i := 0; i < size && err == nil; i++ {
		err = deleteTable(db, &Tunnel{}, "url_cgi = ?", tunnels[i].URLCgi)
	}
	return err
}

// DeleteDeviceByClientID 根据客户端ID删除一行设备数据，其关联数据同时被删除
func (dev *Device) DeleteDeviceByClientID(db *gorm.DB) error {
	if 0 == len(dev.Cid) {
		return errors.New(PkgName + "condition is nil")
	}
	return deleteTable(db, &Device{}, "cid = ?", dev.Cid)
}

//UpdateDeviceByID 更新设备数据，dev.Cid不能为nil
func (dev *Device) UpdateDeviceByID(db *gorm.DB) error {
	if len(dev.Cid) == 0 {
		return errors.New(PkgName + "condition [id] is nil")
	}
	// 1.从 'tunnels' 表中删除该设备已有的关联数据
	if err := deleteTable(db, &Tunnel{}, "fk_cid = ?",dev.Cid); err != nil {
		return err
	}
	// 2. 将新的dev数据更新到 'devices' 表中，将Tunnels更新到 'tunnels' 表中
	return updateTable(db, dev, "cid = ?", dev.Cid)
}

// QueryDevicesByClientID 根据ID查询设备数据，dev.Cid不能为nil
func (dev *Device) QueryDevicesByClientID(db *gorm.DB) (*Device, error) {
	// 获取全部匹配的记录
	if len(dev.Cid) == 0 {
		return nil, errors.New(PkgName + "condition [id] is nil")
	}

	if err := queryTable(db, nil, dev, "cid = ?", dev.Cid); err != nil {
		return nil, err
	}
	return dev, nil
}

// DeleteOffLineDevicesByDay 根据离线天数删除离线设备数据，对应的关联数据也会从 'tunnels' 中删除
func DeleteOffLineDevicesByDay(db *gorm.DB, days uint32) error {
	return deleteTable(db, &Device{}, "state > ?", days)
}

// QueryAllDevices 查询所有设备数据
func QueryAllDevices(db *gorm.DB, order string) (devices []Device, err error) {
	if 0 == len(order) {
		order = "cid"
	}

	if err = queryTable(db, order, &devices, nil); err != nil {
		return nil, err
	}
    return devices, nil
}

// QueryOnlineDevices 查询在线设备数据
func QueryOnlineDevices(db *gorm.DB, order string) (devices []Device, err error) {
	if 0 == len(order) {
		order = "cid"
	}

	// state大于0表示设备离线天数，只获取该值为0的在线设备
	if err = queryTable(db, order, &devices, "state = ?", 0); err != nil {
		return nil, err
	}
	return devices, nil
}

// QueryDevicesByVendor 根据厂商查询在线设备数据
func QueryDevicesByVendor(db *gorm.DB, vendor string, order string) (devices []Device, err error) {
	// 获取指定厂商的全部设备，默认按dev_id排序
	if 0 == len(order) {
		order = "cid"
	}

	if err = queryTable(db, order, &devices,"Vendor = ?", vendor); err != nil {
		return nil, err
	}
	return devices, nil
}

// insertTable 向表中插入数据，并且存在字段冲突时，按 conflict 指定的方式操作
func insertTable(db *gorm.DB, tbl interface{}, conflict *clause.OnConflict) error {
	return db.Clauses(conflict).Create(tbl).Error
}

// deleteTable 根据条件删除表中的数据
func deleteTable(db *gorm.DB, tbl, cond interface{}, args ...interface{}) error {
	return db.Where(cond, args...).Delete(tbl).Error
}

// appendAssociation 向tbl表的关联表中添加一行数据
func appendAssociation(db *gorm.DB, tbl interface{}, assName string, assTbl interface{}) error {
	return db.Model(tbl).Association(assName).Append(&assTbl)
}

// updateTable 根据条件更新表中的数据
func updateTable(db *gorm.DB, tbl, cond interface{}, arg interface{}) error {
	return db.Where(cond, arg).Session(&gorm.Session{FullSaveAssociations: true}).Updates(tbl).Error
}

// queryTable 根据条件查询数据表，并按指定字段排序查询结果
func queryTable(db *gorm.DB, odr, tbl, cond interface{}, args ...interface{}) error {
	if tbl == nil {
		return errors.New(PkgName + "parameter is nil")
	}

	if cond == nil {
		if odr == nil {
			return db.Preload(clause.Associations).Find(tbl).Error
		} else {
			return db.Order(odr).Preload(clause.Associations).Find(tbl).Error
		}
	} else {
		if odr == nil {
			return db.Where(cond, args...).Preload(clause.Associations).Find(tbl).Error
		} else {
			return db.Where(cond, args...).Order(odr).Preload(clause.Associations).Find(tbl).Error
		}
	}
}
