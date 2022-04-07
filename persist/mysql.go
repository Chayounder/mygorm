package persist

import (
    "errors"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
    "pro_mysql/system"
    "strconv"
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

// MigrateModel 创建 Device 和 Tunnel 模型
func MigrateModel(db *gorm.DB) error {
    return migrateTableModel(db, &Device{}, &Tunnel{})
}

// DestroyModel 销毁 Device 和 Tunnel 模型
func DestroyModel(db *gorm.DB) error {
    return destroyTableModel(db, &Device{}, &Tunnel{})
}

// InsertDevice 插入一行device数据
func (dev *Device) InsertDevice(db *gorm.DB) error {
    //如果设备cid冲突，该设备信息已经存在，不做任何刷新;如果要刷新，请调用刷新接口
    conflict := clause.OnConflict {
        Columns: []clause.Column{{Name: "cid"}},
        DoNothing: true,
    }
    return insertTable(db, dev, &conflict)
}

// AppendDeviceAssociation 添加dev的关联数据。
func (dev *Device) AppendDeviceAssociation(db *gorm.DB) error {
    /*
     * 1. 如果模型dev不在数据表中，则其关联数据添加失败（设置了约束：OnDelete:CASCADE）
     * 2. 目前没有针对Tunnels模型设置unique字段和冲突限制，只要主键值不冲突，均可插入
     */
    return appendAssociation(db, dev, "Tunnels", dev.Tunnels)
}

// DeletedDeviceAssociation 删除设备的关联数据。
func (dev *Device) DeletedDeviceAssociation(db *gorm.DB) error {
    if 0 == len(dev.Tunnels) {
        return errors.New(PkgName + "condition is nil")
    }

    size := len(dev.Tunnels)
    tunnels := dev.Tunnels
    tx := db.Begin()
    tx.SavePoint("DelDevAss")

    for i := 0; i < size; i++ {
        err := deleteTableByCond(tx, &Tunnel{}, "url_cgi = ?", tunnels[i].URLCgi)
        if err != nil {
            tx.RollbackTo("DelDevAss")
            return errors.New(PkgName +
                "delete[" + strconv.Itoa(i) + "]tunnel err:" + err.Error())
        }
    }
    return tx.Commit().Error
}

// DeleteDeviceByCID 根据客户端ID删除一行设备数据，其关联数据同时被删除
func (dev *Device) DeleteDeviceByCID(db *gorm.DB) error {
    if 0 == len(dev.Cid) {
        return errors.New(PkgName + "condition is nil")
    }
    return deleteTableByCond(db, &Device{}, "cid = ?", dev.Cid)
}

//UpdateDeviceByCID 更新设备数据，dev.Cid不能为nil
func (dev *Device) UpdateDeviceByCID(db *gorm.DB) error {
    if len(dev.Cid) == 0 {
        return errors.New(PkgName + "condition [id] is nil")
    }

    tx := db.Begin()
    tx.SavePoint("UpdDev")

    // 1.从 'tunnels' 表中删除该设备已有的关联数据
    if err := deleteTableByCond(tx, &Tunnel{}, "fk_cid = ?",dev.Cid); err != nil {
        tx.RollbackTo("UpdDev")
        return err
    }

    // 2. 将新的dev数据更新到 'devices' 表中，将Tunnels更新到 'tunnels' 表中
    if err := updateTable(tx, dev, "cid = ?", dev.Cid); err != nil {
        tx.RollbackTo("UpdDev")
        return err
    }
    return tx.Commit().Error
}

func (dev *Device) UpdateDeviceStateByCID(db *gorm.DB) error {
    //dev := Device{State: dev.State}
    return updateTableColumns(db, dev, "cid = ?", dev.Cid, "state")
}

// QueryDevicesByClientID 根据ID查询设备数据，dev.Cid不能为nil
func (dev *Device) QueryDevicesByClientID(db *gorm.DB) error {
    // 获取全部匹配的记录
    if len(dev.Cid) == 0 {
        return errors.New(PkgName + "condition [id] is nil")
    }

    if err := queryAssociationTable(db, nil, dev, "cid = ?", dev.Cid); err != nil {
        return  err
    }
    return nil
}

// QueryDevicesByClientID 根据ID查询设备数据，dev.Cid不能为nil
func (dev *Device) QueryDeviceTunnels(db *gorm.DB) error {
    if len(dev.Cid) == 0 {
        return errors.New(PkgName + "condition [id] is nil")
    }

    if err := queryAssociation(db, dev, "Tunnels", &dev.Tunnels); err != nil {
        return  err
    }
    return nil
}

// DeleteOffLineDevicesByDay 根据离线天数删除离线设备数据，对应的关联数据也会从 'tunnels' 中删除
func DeleteOffLineDevicesByDay(db *gorm.DB, days uint32) error {
    return deleteTableByCond(db, &Device{}, "state > ?", days)
}

// QueryAllDevices 查询所有设备数据
func QueryAllDevices(db *gorm.DB, order string) (devices []Device, err error) {
    if 0 == len(order) {
        order = "cid"
    }

    if err = queryAssociationTable(db, order, &devices, nil); err != nil {
        return nil, err
    }
    return devices, nil
}

func QueryAllDevicesVer2(db *gorm.DB, order string, n int) (devices []Device, err error) {
    if 0 == len(order) {
        order = "cid"
    }

    tx := db.Begin()
    tx.SavePoint("QueryAllDev")
    if err = queryTable(db, order, &devices, nil); err != nil {
        tx.RollbackTo("QueryAllDev")
        return nil, err
    }

    for i, device := range devices {
        if err = queryTable(db, order, &devices[i].Tunnels, map[string]interface{}{"fk_cid": device.Cid}); err != nil {
            tx.RollbackTo("QueryAllDev")
            return nil, err
        }
    }

/*  var tunnels []Tunnel
    if err = queryTable(db, order, &tunnels, nil); err != nil {
        tx.RollbackTo("QueryAllDev")
        return nil, err
    }*/

    tx.Commit()

/*  for i, device := range devices {
        for j, tunnel := range tunnels {
            if device.Cid == tunnel.FkCid {
                tunnels = append(devices[i].Tunnels, tunnel)
                tunnels = append(tunnels[:j], tunnels[j + 1 :]...) // 删除当前元素
            }
        }
    }*/

    return devices, nil
}


func QueryAllDevicesBatches(db *gorm.DB, order string) (devices []Device, err error) {
    if 0 == len(order) {
        order = "cid"
    }

    return queryDeviceBatches(db, order, 75535, nil)
}

// QueryOnlineDevices 查询在线设备数据
func QueryOnlineDevices(db *gorm.DB, order string) (devices []Device, err error) {
    if 0 == len(order) {
        order = "cid"
    }

    // state大于0表示设备离线天数，只获取该值为0的在线设备
    if err = queryAssociationTable(db, order, &devices, "state = ?", 0); err != nil {
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

    if err = queryAssociationTable(db, order, &devices,"Vendor = ?", vendor); err != nil {
        return nil, err
    }
    return devices, nil
}

// migrateTableModel 自动迁移数据表模型
func migrateTableModel(db *gorm.DB, dst ...interface{}) (err error) {
    /*
     * 1. 根据模型创建表结构，同时自动约束，列和索引
     * 2. 根据配置（gorm.Config.DisableForeignKeyConstraintWhenMigrating）是否创建外键约束，
     * 3. 每次调用会根据传入的模型更改现有的模型，包括：列类型（如果其大小、精度、是否为空可更改），但不会删除未使用的列。
     */
    para := "ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='device info'"
    return db.Set("gorm:table_options", para).AutoMigrate(dst...)
}

// destroyTableModel 销毁数据表模型
func destroyTableModel(db *gorm.DB, dst ...interface{}) error {
    return db.Migrator().DropTable(dst...)
}

// insertTable 向表中插入数据，并且存在字段冲突时，按 conflict 指定的方式操作
func insertTable(db *gorm.DB, tbl interface{}, conflict *clause.OnConflict) error {
    // 不禁止同时创建关联模型，当创建主模型时，即使主模型因冲突没做任务操作，关联模型数据也会被插入对应的数据表中。
    return db.Clauses(conflict).Create(tbl).Error
}

// insertTableOmitAssociation 向表中插入数据，并且存在字段冲突时，按 conflict 指定的方式操作
func insertTableOmitAssociation(db *gorm.DB, tbl interface{}, conflict *clause.OnConflict) (int64, error) {
    /*
     * 禁止同时创建关联模型：当创建主模型时，如果其存在关联，则同时将关联模型的数据插入关联模型对应的表中；
     * 如果不禁止同时创建关联模型，当创建主模型时，即使主模型因冲突不做任务操作，关联模型数据也会被插入对应的数据表中。
     */
    result := db.Clauses(conflict).Omit(clause.Associations).Create(tbl)
    return result.RowsAffected, result.Error
}

// deleteTable 删除数据表中全部数据
func deleteTable(db *gorm.DB, tbl interface{}) error {
    // return db.Delete(&User{}).Error  // gorm.ErrMissingWhereClause
    //return db.Exec("DELETE FROM users").Error
    return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(tbl).Error
}

// deleteTableByCond 根据条件删除表中的数据
func deleteTableByCond(db *gorm.DB, tbl, cond interface{}, args ...interface{}) error {
    return db.Where(cond, args...).Delete(tbl).Error
}

// appendAssociation 向tbl表的关联表中添加一行数据
func appendAssociation(db *gorm.DB, tbl interface{}, assName string, assTbl interface{}) error {
    // 为什么还要取地址？某些情况会覆盖原来的数据行，某些情况一条数据会插入两次
    return db.Model(tbl).Association(assName).Append(&assTbl)
}

// updateTable 根据条件更新表中的数据
func updateTable(db *gorm.DB, tbl, cond interface{}, arg ...interface{}) error {
    // Session(&gorm.Session{FullSaveAssociations: true})是否需要？？
    return db.Where(cond, arg...).Session(&gorm.Session{FullSaveAssociations: true}).Updates(tbl).Error
}

func updateTableColumns(db *gorm.DB, tbl, cond, arg interface{}, cols interface{}) error {
    return db.Where(cond, arg).Select(cols).Updates(tbl).Error
}

// queryAssociationTable 根据条件查询数据表，并按指定字段排序查询结果
func queryAssociationTable(db *gorm.DB, odr, tbl, cond interface{}, args ...interface{}) error {
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

// queryTableLimit 根据条件查询数据表，并按指定字段排序查询结果
func queryTableLimit(db *gorm.DB, n int, odr, tbl, cond interface{}, args ...interface{}) error {
    if tbl == nil {
        return errors.New(PkgName + "parameter is nil")
    }

    if cond == nil {
        if odr == nil {
            return db.Limit(n).Preload(clause.Associations).Find(tbl).Error
        } else {
            return db.Order(odr).Limit(n).Preload(clause.Associations).Find(tbl).Error
        }
    } else {
        if odr == nil {
            return db.Where(cond, args...).Limit(n).Preload(clause.Associations).Find(tbl).Error
        } else {
            return db.Where(cond, args...).Order(odr).Preload(clause.Associations).Limit(n).Find(tbl).Error
        }
    }
}
// queryTable 根据条件查询数据表，并按指定字段排序查询结果
func queryTable(db *gorm.DB, odr, tbl, cond interface{}, args ...interface{}) error {
    if tbl == nil {
        return errors.New(PkgName + "parameter is nil")
    }

    if cond == nil {
        if odr == nil {
            return db.Find(tbl).Error
        } else {
            return db.Order(odr).Find(tbl).Error
        }
    } else {
        if odr == nil {
            return db.Where(cond, args...).Find(tbl).Error
        } else {
            return db.Where(cond, args...).Find(tbl).Error
        }
    }
}


func queryDeviceBatches(db *gorm.DB, odr string, num int, cond interface{}, args ...interface{}) (devices []Device, err error) {
    var results []Device

    function := func(tx *gorm.DB, batch int) error {
        devices = append(devices, results...)
        return nil  // 如果返回错误会终止后续批量操作
    }

    if cond == nil {
        if len(odr) == 0 {
            err = db.FindInBatches(&results, num, function).Error
        } else {
            err = db.Order(odr).FindInBatches(&results, num, function).Error
        }
    } else {
        if len(odr) == 0 {
            err = db.Where(cond, args...).FindInBatches(&results, num, function).Error
        } else {
            err = db.Where(cond, args...).Order(odr).FindInBatches(&results, num, function).Error
        }
    }
    return
}

func queryAssociation(db *gorm.DB, mode interface{}, name string,  results interface{}) error {
    return db.Model(mode).Association(name).Find(results)
}
