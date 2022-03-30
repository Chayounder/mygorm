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
	return db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='device info'").AutoMigrate(&Device{}, &Tunnel{})
	//tables := []interface{} {&Device{}, &TunnelUrl{}}
	//return db.AutoMigrate(&Device{} /*, &Tunnel{} */)
}

// InsertDevice Insert device info to 'devices' table
func (dev *Device) InsertDevice(db *gorm.DB) error {
	return insertDevice(db, dev)
}

// AppendDeviceAssociation append a dev's association into 'tunnels' table.
// If there is no dev in the 'devices' table, append fail because of "OnDelete:CASCADE".
func (dev *Device) AppendDeviceAssociation(db *gorm.DB) error {
	for _, tunnel := range dev.Tunnels {
		//if err := insertDeviceAssociation(db, dev, &tunnel); err != nil {
		if err := appendAssociation(db, dev, "Tunnels", &tunnel); err != nil {
			return err
		}
	}
	return nil
}

// DeletedDeviceAssociation Delete dev's association from 'tunnels' table.
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

// DeleteDeviceByClientID Delete a line device info from 'devices' table
// and delete the dev's association info from 'tunnels' table
func (dev *Device) DeleteDeviceByClientID(db *gorm.DB) error {
	if 0 == len(dev.Cid) {
		return errors.New(PkgName + "condition is nil")
	}
	//return deleteDeviceByClientID(db, dev.Cid)
	return deleteTable(db, &Device{}, "cid = ?", dev.Cid)
}

//UpdateDeviceByID Update device info and dev's association info
func (dev *Device) UpdateDeviceByID(db *gorm.DB) error {
	// 1.delete the dev's old association info in 'tunnels' table
	if err := deleteTable(db, &Tunnel{}, "fk_cid = ?",dev.Cid); err != nil {
		return err
	}
	// 2. update the dev's info in 'devices' table, and dev's new association info will be append to 'tunnels' table
	//return updateDeviceByID(db, dev)
	return updateTable(db, dev, "cid = ?", dev.Cid)
}

func (dev *Device) QueryDevicesByClientID(db *gorm.DB) (*Device, error) {
	// 获取全部匹配的记录
	if 0 == len(dev.Cid) {
		return nil, errors.New(PkgName + "condition is nil")
	}

	if err := queryTable(db, nil, dev, "cid = ?", dev.Cid); err != nil {
		return nil, err
	}
	return dev, nil
}

// DeleteOffLineDevicesByDay delete a line device info from 'devices' table
// and the corresponding association info from 'tunnels' table
func DeleteOffLineDevicesByDay(db *gorm.DB, days uint32) error {
	//return deleteOffLineDeviceByDays(db, days)
	return deleteTable(db, &Device{}, "state > ?", days)
}

func QueryAllDevices(db *gorm.DB, order string) (devices []Device, err error) {
	if 0 == len(order) {
		order = "cid"
	}

	if err = queryTable(db, order, &devices, nil); err != nil {
		return nil, err
	}
    return devices, nil
}

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

func QueryOnlineDevicesByVendor(db *gorm.DB, vendor string, order string) (devices []Device, err error) {
	// 获取指定厂商的全部设备，默认按dev_id排序
	if 0 == len(order) {
		order = "cid"
	}

	if err = queryTable(db, order, &devices,"Vendor = ? AND state = ?", vendor, 0); err != nil {
		return nil, err
	}
	return devices, nil
}

// insertDevice Insert device info to 'devices' table
// 插入一行设备信息:INSERT INTO <table_name> (col1, col2,...,colN) VALUES (v1, v2,...,vN);
// 如果设备sn冲突，该设备信息已经存在，不做任何刷新。
// 如果要刷新，请调用刷新接口
func insertDevice(db *gorm.DB, dev *Device) error {
	conflict := clause.OnConflict {
		Columns: []clause.Column{{Name: "sn"}},
		DoNothing: true,
	}

	//return db.Select("*").Create(dev).Error
	return db.Clauses(conflict).Create(dev).Error
	//return db.Clauses(conflict).Omit(clause.Associations).Create(dev).Error
}

func deleteAllDevices(db *gorm.DB) error {
	// DELETE FROM devices
	//return db.Exec("DELETE FROM devices").Error
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Device{}).Error
}

func deleteAllTunnels(db *gorm.DB) error {
	// DELETE FROM tunnels
	//return db.Exec("DELETE FROM tunnels").Error
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Tunnel{}).Error
}
// deleteTableByCond 根据条件删除表中的数据
func deleteTable(db *gorm.DB, tbl, cond interface{}, args ...interface{}) error {
	//return db.Where("fk_cid = ?", cid).Delete(&Tunnel{}).Error
	return db.Where(cond, args...).Delete(tbl).Error
}
// deleteDeviceByClientID DONE
func deleteDeviceByClientID(db *gorm.DB, cid string) error {
	// 删除一行设备信息
	return db.Where("cid = ?", cid).Delete(&Device{}).Error
}
// deleteOffLineDeviceByDays DONE
func deleteOffLineDeviceByDays(db *gorm.DB, days uint32) error {
	return db.Where("state > ?", days).Delete(&Device{}).Error
}

// appendAssociation 向tbl表的关联表中添加一行数据
func appendAssociation(db *gorm.DB, tbl interface{}, assName string, assTbl interface{}) error {
	// &assTbl 为什么要传耳机指针，否则的话一个 assTbl 会被添加两次
	return db.Model(tbl).Association(assName).Append(&assTbl)
}
func insertDeviceAssociation(db *gorm.DB, dev *Device, tunnel *Tunnel) error {
	// &tunnel 为什么要传耳机指针，否则的话一个tunnel会被添加两次
	return db.Model(dev).Association("Tunnels").Append(&tunnel)
}
func deleteDeviceAllAssociation(db *gorm.DB, cid string) error {
	//fmt.Println("count:", db.Model(dev).Association("Tunnels").Count())
	//
	//_ = db.Model(dev).Association("Tunnels").Clear()
	//_ = db.Model(&Device{}).Association("Tunnels").Delete(dev)
	return db.Where("fk_cid = ?", cid).Delete(&Tunnel{}).Error
	//fmt.Println("count:", db.Model(dev).Association("Tunnels").Count())
	//return nil
}
func deleteDeviceAssociationByUrl(db *gorm.DB, url string) error {
	//fmt.Println("count:", db.Model(dev).Association("Tunnels").Count())
	//_ = db.Model(dev).Association("Tunnels").Clear()
		//_ = db.Model(dev).Association("Tunnels").Delete(tunnel)
	return db.Where("url_cgi = ?", url).Delete(&Tunnel{}).Error
	//fmt.Println("count:", db.Model(dev).Association("Tunnels").Count())
	//return nil
}

// updateTable 根据条件更新表中的数据
func updateTable(db *gorm.DB, tbl, cond interface{}, arg interface{}) error {
	//return db.Where("cid = ?", dev.Cid).Session(&gorm.Session{FullSaveAssociations: true}).Updates(dev).Error
	return db.Where(cond, arg).Session(&gorm.Session{FullSaveAssociations: true}).Updates(tbl).Error
}
func updateDeviceByID(db *gorm.DB, dev *Device) error {
	/*    // 在冲突时，更新除主键以外的所有列都新值。
	      //return db.Clauses(clause.OnConflict{UpdateAll: true }).Create(&dev).Error

	      // 没有匹配规则，将所有行的dev_id字段更新成指定值
	      return db.Update("dev_id", "hhhhhhh").Error
	      return db.Model(&Device{}).Update("dev_id", "hhhhhhh").Error

	      // 按dev主键匹配行，并将该行的指定字段（dev_id）更新成指定值
	      return db.Model(dev).Update("dev_id", "hhhhhhh").Error             // 只更新指定的单个字段
	      return db.Model(dev).Select("dev_id", "dev_sn").Updates(dev).Error // 只更新指定的两个字段。当只指定一个字段时，与上面代码一样
	      return db.Model(dev).Select("*").Updates(dev).Error                // "*" 表示更新所有字段

	      // 将表格中与dev_id匹配的行的所有字段更新成dev指定值
	      return db.Where("dev_id = ?", dev.DevID).Updates(dev).Error

	      // 按dev的主键和dev_id匹配行，并改行所有字段更新成dev指定值（只会更新dev中非零值的字段）
	      return db.Model(dev).Where("dev_id = ?", dev.DevID).Updates(dev).Error
	      // 按dev的dev_id匹配行，并改行所有字段更新成dev指定值（只会更新dev中非零值的字段）
	      return db.Model(&Device{}).Where("dev_id = ?", dev.DevID).Updates(dev).Error
	*/

	// 1.只会更新dev中非零值的字段，且忽略dev_id字段的更新
	// 2.当使用了 Model 方法，且该对象主键有值，该值会被用于构建条件
	//return db.Model(dev).Where("cid = ?", dev.ClientID).Omit("DevID", clause.Associations).Updates(dev).Error

	return db.Where("cid = ?", dev.Cid).Session(&gorm.Session{FullSaveAssociations: true}).Updates(dev).Error
}

// queryTableAll 查询主表中的所有数据，包括主表中每项数据在从表中的关联数据
func queryTableAll(db *gorm.DB, order string, tbl interface{}) error {
	// 预加载
	return db.Order(order).Preload(clause.Associations).Find(tbl).Error
}
func queryAllDevices(db *gorm.DB, order string) ([]Device, error) {
	var devices []Device
	err := db.Order(order).Find(&devices).Error
	return devices, err
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

func queryOnlineDevices(db *gorm.DB, order string) ([]Device, error) {
	var devices []Device
	// state大于0表示设备离线天数，只获取该值为0的在线设备
	err := db.Where("state = ?", 0).Order(order).Find(&devices).Error
	return devices, err
}

func queryOnlineDevicesByVendor(db *gorm.DB, vendor string, order string) ([]Device, error) {
	var devices []Device
	// state大于0表示设备离线天数，只获取该值为0的在线设备
	err := db.Where("Vendor = ? AND state = ?", vendor, 0).Order(order).Find(&devices).Error
	return devices, err
}

func queryDeviceByClientID(db *gorm.DB, cid string) (*Device, error) {
	var device Device
	err := db.Where("cid = ?", cid).Find(&device).Error
	return &device, err
}

// queryTableAssociation Query association data
func queryTableAssociation(db *gorm.DB, tbl interface{}, assName string, out interface{}) error {
	return db.Model(tbl).Association(assName).Find(out)
}
func queryDeviceAllAssociation(db *gorm.DB, dev *Device) ([]Tunnel, error) {
	var tunnels []Tunnel
	err := db.Model(dev).Association("Tunnels").Find(&tunnels)
	return tunnels, err
}


/*
func queryAllTunnels(db *gorm.DB, order string) ([]Tunnel, error) {
	var tunnels []Tunnel
	err := db.Order(order).Find(&tunnels).Error
	return tunnels, err
}

func QueryAllTunnels(db *gorm.DB, order string) ([]Tunnel, error) {
	// 获取全部记录，默认按dev_id排序
	if 0 == len(order) {
		return queryAllTunnels(db, "dev_sn")
	} else {
		return queryAllTunnels(db, order)
	}
}*/

/*
// TblInsert insert 'obj' into 'tbl', returns error.
func (tbl *ObjTable)TblInsert(obj *Object) error {

    return nil
}

// TblDelete delete 'obj' from 'tbl', returns error.
func (tbl *ObjTable)TblDelete(obj *Object) error {

    return nil
}

// TblUpdate update 'old' in 'tbl' to 'new', returns error.
func (tbl *ObjTable)TblUpdate(old, new *Object) (err error) {

    return
}

// TblLookup lookup 'obj' from 'tbl', returns 'Obj' and error.
func (tbl *ObjTable)TblLookup(obj Object) (Obj *Object, err error) {

    //tblLookup(tbl.Obj)
    return
}*/
