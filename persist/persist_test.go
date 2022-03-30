package persist

import (
	"fmt"
	"pro_mysql/system"
	"testing"
	"time"
)

var originDevs = []Device{
	{
		Cid: "pouwer", SN: "M749201012400366", Vendor: "IP-COM", Mode: "G3v3.6", SWVersion: "V15.11.0.17(9502)", MAC: "D8:38:0D:6F:67:38", Addr: "110.84.74.114:23185",
		Attribution: "湖南-长沙-芙蓉区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
		Tunnels: []Tunnel{{URLCgi: "http://M749201012400366-1.web.ip-com.com.cn:8080"},{URLCgi: "http://M749201012400366-2.web.ip-com.com.cn:8080"}},
	},	{
		Cid: "jkgtvb", SN: "F6ACF73724F56BED", Vendor: "Tenda", Mode: "M30V2.6", SWVersion: "V16.01.0.6(2700)", MAC: "B0:DF:C1:F8:D2:B8", Addr: "183.94.242.60:3107",
		Attribution: "湖南-长沙-开福区", State: 10, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
		Tunnels: []Tunnel{{URLCgi: "http://F6ACF73724F56BED-1.web.ip-com.com.cn:8080"},{URLCgi: "http://F6ACF73724F56BED-2.web.ip-com.com.cn:8080"}},
	},	{
		Cid: "ndfdgr", SN: "K7LMF89452F26LMK", Vendor: "Tenda", Mode: "M50V3.0", SWVersion: "V19.01.1.6(2689)", MAC: "C0:7K:1C:B8:J4:PK", Addr: "211.56.124.70:2385",
		Attribution: "湖南-长沙-雨花区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: false,
		Tunnels: []Tunnel{{URLCgi: "http://K7LMF89452F26LMK-1.web.ip-com.com.cn:8080"},{URLCgi: "http://K7LMF89452F26LMK-2.web.ip-com.com.cn:8080"}},
	},	{
		Cid: "uvwxyz", SN: "MC07101412800052", Vendor: "IP-COM", Mode: "M50V3.3", SWVersion: "17.11.1.16(2375)", MAC: "2B:B8:D2:B0:J4:5D", Addr: "124.242.65.70:8523",
		Attribution: "湖南-长沙-岳麓区", State: 20, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: false,
	},	{
		Cid: "acvbmb", SN: "1906747201001359", Vendor: "Tenda", Mode: "W15Ev2", SWVersion: "V71.1.1.59(2193)", MAC: "50:2B:73:BD:5D:98", Addr: "117.24.126.121:1030",
		Attribution: "湖南-长沙-天心区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
	}, {
		Cid: "huytto", SN: "2003938101002253", Vendor: "IP-COM", Mode: "W20Ev5", SWVersion: "V12.01.0.4(1576)", MAC: "C8:3A:35:6E:00:78", Addr: "36.62.56.112:27117",
		Attribution: "湖南-长沙-宁乡市", State: 367, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
	},
}

var againDevs = []Device{
	{
		Cid: "pouwer", SN: "M7492010124000366", Vendor: "IP-COM", Mode: "G4v4.7", SWVersion: "V16.22.1.20(9888)", MAC: "D8:38:0D:6F:67:38", Addr: "110.84.74.114:23185",
		Attribution: "湖南-长沙-开福区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
		Tunnels: []Tunnel{{URLCgi: "again-pouwer@example.com"},{URLCgi: "again-M7492010124000366@example.com"}},
	},	{
		Cid: "jkgtvb", SN: "F6ACF73724F56BED", Vendor: "Tenda", Mode: "M60V1.8", SWVersion: "V18.05.1.8(5555)", MAC: "B0:DF:C1:F8:D2:B8", Addr: "183.94.242.60:3107",
		Attribution: "湖南-长沙-芙蓉区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
	}}

var updateDevs = []Device{
	{   // dev-tun -> dev-tun
		Cid: "pouwer", SN: "upt-M749201012400366", Vendor: "upt-IP-COM", Mode: "upt-G3v3.6", SWVersion: "upt-V15.11.0.17(9502)", MAC: "upt-D8:38:0D:6F:67:38", Addr: "upt-110.84.74.114:23185",
		Attribution: "upt-湖南-长沙-芙蓉区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
		Tunnels: []Tunnel{{URLCgi: "http://upt-M749201012400366-1.web.ip-com.com.cn:8080"},{URLCgi: "http://upt-M749201012400366-2.web.ip-com.com.cn:8080"}},
	},	{// dev-tun -> tun
		Cid: "jkgtvb", SN: "F6ACF73724F56BED",
		Tunnels: []Tunnel{{URLCgi: "http://upt-F6ACF73724F56BED-1.web.ip-com.com.cn:8080"},{URLCgi: "http://upt-F6ACF73724F56BED-2.web.ip-com.com.cn:8080"}},
	}, {// dev-tun -> dev
		Cid: "ndfdgr", SN: "K7LMF89452F26LMK", Vendor: "upt-Tenda", Mode: "upt-M50V3.0", SWVersion: "upt-V19.01.1.6(2689)", MAC: "upt-C0:7K:1C:B8:J4:PK", Addr: "upt-211.56.124.70:2385",
		Attribution: "upt-湖南-长沙-雨花区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: false,
	}, {// dev -> dev-tun
		Cid: "uvwxyz", SN: "MC07101412800052", Vendor: "upt-IP-COM", Mode: "upt-M50V3.3", SWVersion: "upt-17.11.1.16(2375)", MAC: "upt-2B:B8:D2:B0:J4:5D", Addr: "upt-124.242.65.70:8523",
		Attribution: "upt-湖南-长沙-岳麓区", State: 100, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: false,
		Tunnels: []Tunnel{{URLCgi: "http://upt-MC07101412800052-1.web.ip-com.com.cn:8080"},{URLCgi: "http://upt-MC07101412800052-2.web.ip-com.com.cn:8080"}},
	}, {// dev -> tun
		Cid: "acvbmb", SN: "1906747201001359",
		Tunnels: []Tunnel{{URLCgi: "http://upt-1906747201001359-1.web.ip-com.com.cn:8080"},{URLCgi: "http://upt-1906747201001359-2.web.ip-com.com.cn:8080"}},
	}, {// dev -> dev
		Cid: "huytto", SN: "2003938101002253", Vendor: "upt-IP-COM", Mode: "upt-W20Ev5", SWVersion: "upt-V12.01.0.4(1576)", MAC: "upt-C8:3A:35:6E:00:78", Addr: "upt-36.62.56.112:27117",
		Attribution: "upt-湖南-长沙-宁乡市", State: 400, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
	},
}

var appendDevs = []Device{
	{   // dev-tun -> dev-tun
		Cid: "pouwer", SN: "M749201012400366",
		Tunnels: []Tunnel{{URLCgi: "http://apd-M749201012400366-1.web.ip-com.com.cn:8080"}},
	},	{// dev -> dev-tun
		Cid: "uvwxyz", SN: "MC07101412800052",
		Tunnels: []Tunnel{{URLCgi: "http://apd-MC07101412800052-1.web.ip-com.com.cn:8080"},{URLCgi: "http://apd-MC07101412800052-2.web.ip-com.com.cn:8080"}},
	}, {// dev -> tun
		Cid: "acvbmb", SN: "1906747201001359",
		Tunnels: []Tunnel{{URLCgi: "http://apd-1906747201001359-1.web.ip-com.com.cn:8080"}},
	}, {// dev -> dev
		Cid: "huytto", SN: "2003938101002253",
		Tunnels: []Tunnel{{URLCgi: "http://apd-2003938101002253-1.web.ip-com.com.cn:8080"},{URLCgi: "http://apd-2003938101002253-2.web.ip-com.com.cn:8080"}},
	},
}


func TestInitDb(t *testing.T) {
	err := system.LoadDbConfiguration("..\\config\\config.yaml")
	if err != nil {
		t.Log("load db config error:", err)
		return
	}

	db, err := OpenDb()
	if err != nil {
		t.Log("connect db error:", err)
		return
	}

	// 如果数据库中的表已经存在，调用AutoMigrate函数会怎样？
	err = InitDb(db)
	if err != nil {
		t.Log("init db error:", err)
	}

	// OpenDb后是否需要close，close后是否还能操作数据库

	// 执行open之后，继续open

	// 多个routine同时操作同一个打开的db

	CloseDB(db)
}

func TestInsertDevice(t *testing.T) {
	err := system.LoadDbConfiguration("..\\config\\config.yaml")
	if err != nil {
		t.Log("load db config error:", err)
		return
	}

	db, err := OpenDb()
	if err != nil {
		t.Log("connect db error:", err)
		return
	}

	// 如果数据库中的表已经存在，调用AutoMigrate函数会怎样？
	err = InitDb(db)
	if err != nil {
		t.Log("init db error:", err)
	}

	devices := originDevs
	for i := 0; i < len(devices); i++ {
		err = devices[i].InsertDevice(db)
		if err != nil {
			t.Log("InsertDevice error:", err)
		}
	}

	CloseDB(db)
}

func TestInsertDeviceAgain(t *testing.T) {
	err := system.LoadDbConfiguration("..\\config\\config.yaml")
	if err != nil {
		t.Log("load db config error:", err)
		return
	}

	db, err := OpenDb()
	if err != nil {
		t.Log("connect db error:", err)
		return
	}

	// 如果数据库中的表已经存在，调用AutoMigrate函数会怎样？
	err = InitDb(db)
	if err != nil {
		t.Log("init db error:", err)
	}

	devices := againDevs
	for i := 0; i < len(devices); i++ {
		err = devices[i].InsertDevice(db)
		if err != nil {
			t.Log("InsertDevice error:", err)
		}
	}

	CloseDB(db)
}

func TestUpdateDevice(t *testing.T) {
	err := system.LoadDbConfiguration("..\\config\\config.yaml")
	if err != nil {
		t.Log("load db config error:", err)
		return
	}

	db, err := OpenDb()
	if err != nil {
		t.Log("connect db error:", err)
		return
	}

	devices := updateDevs
	for i := 1; i < len(devices); i++ {
		err = devices[i].UpdateDeviceByID(db)
		if err != nil {
			t.Log("UpdateDeviceByID error:", err)
		}
	}
	CloseDB(db)
}

func TestAppendDeviceAssociation(t *testing.T) {
	err := system.LoadDbConfiguration("..\\config\\config.yaml")
	if err != nil {
		t.Log("load db config error:", err)
		return
	}

	db, err := OpenDb()
	if err != nil {
		t.Log("connect db error:", err)
		return
	}

	for	_, device := range appendDevs {
		dev := &device
		err = dev.AppendDeviceAssociation(db)
		if err != nil {
			t.Log("UpdateDeviceByID error:", err)
		}
	}

	CloseDB(db)
}

func TestDeletedDeviceAssociation(t *testing.T) {
	err := system.LoadDbConfiguration("..\\config\\config.yaml")
	if err != nil {
		t.Log("load db config error:", err)
		return
	}

	db, err := OpenDb()
	if err != nil {
		t.Log("connect db error:", err)
		return
	}

	dev := &Device {   // dev-tun -> dev-tun
		/*ClientID: "pouwer", SN: "M749201012400366",*/
		Tunnels: []Tunnel{{URLCgi: "http://apd-MC07101412800052-1.web.ip-com.com.cn:8080"},{URLCgi: "http://upt-MC07101412800052-1.web.ip-com.com.cn:8080"}},
	}
	/*,{URLCgi: "http://apd-M749201012400366-2.web.ip-com.com.cn:8080"}*/
	err = dev.DeletedDeviceAssociation(db)
	if err != nil {
		t.Log("DeletedDeviceAssociation error:", err)
	}


	CloseDB(db)
}

func TestQueryDevice(t *testing.T) {
	err := system.LoadDbConfiguration("..\\config\\config.yaml")
	if err != nil {
		t.Log("load db config error:", err)
		return
	}

	db, err := OpenDb()
	if err != nil {
		t.Log("connect db error:", err)
		return
	}

	devices, err := QueryAllDevices(db, "")
	//fmt.Println(devices)
	for _, dev := range devices {
		fmt.Println(dev)
	}

	fmt.Println("=============================")

	devices, err = QueryOnlineDevices(db, "")
	//fmt.Println(devices)
	for _, dev := range devices {
		fmt.Println(dev)
	}

	fmt.Println("=============================")

	devices, err = QueryDevicesByVendor(db, "Tenda", "")
	//fmt.Println(devices)
	for _, dev := range devices {
		fmt.Println(dev)
	}
	fmt.Println("=============================")

	device := &Device{Cid: "pouwer"}
	device.QueryDevicesByClientID(db)
	fmt.Println(*device)

	CloseDB(db)
}

func TestDeleteOffLineDevice(t *testing.T) {
	err := system.LoadDbConfiguration("..\\config\\config.yaml")
	if err != nil {
		t.Log("load db config error:", err)
		return
	}

	db, err := OpenDb()
	if err != nil {
		t.Log("connect db error:", err)
		return
	}

	err = DeleteOffLineDevicesByDay(db, 0)
	if err != nil {
		t.Log("DeleteOffLineDevice error:", err)
	}

/*	var device = &Device{}
	for i := 0; i < 5; i++ {
		device.ID = uint(i + 1)
		_ = device.DeleteDevice(db)
	}*/

	CloseDB(db)
}

func TestDeleteAll(t *testing.T) {
	err := system.LoadDbConfiguration("..\\config\\config.yaml")
	if err != nil {
		t.Log("load db config error:", err)
		return
	}

	db, err := OpenDb()
	if err != nil {
		t.Log("connect db error:", err)
		return
	}

	device := &Device{Cid: "jkgtvb"/*, SN: "M749201012400366"*/}
	_ = device.DeleteDeviceByClientID(db)

	device = &Device{Cid: "jkgtvb"/*, SN: "M749201012400366"*/}
	_ = device.DeletedDeviceAssociation(db)
	/*	dev,_ := device.QueryDevicesByClientID(db)
	fmt.Println(*dev)
	device.DeleteDeviceByClientID(db)
	err = DeleteAllDevices(db)
	if err != nil {
		t.Log("DeleteAllDevices error:", err)
	}

	err = DeleteAllTunnels(db)
	if err != nil {
		t.Log("DeleteAllTunnels error:", err)
	}*/

	CloseDB(db)
}