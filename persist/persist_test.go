package persist

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"gorm.io/gorm"
	mrand "math/rand"
	"pro_mysql/system"
	"strconv"
	"sync"
	"testing"
	"time"
)

var originDevs = []Device{
	{
		Cid: "pouwer", SN: "M749201012400366", Vendor: "IP-COM", Mode: "G3v3.6", SWVersion: "V15.11.0.17(9502)", MAC: "D8:38:0D:6F:67:38", Addr: "110.84.74.114:23185",
		Attribution: "湖南-长沙-芙蓉区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
		Tunnels: []Tunnel{{URLCgi: "http://M749201012400366-1.web.ip-com.com.cn:8080"},{URLCgi: "http://M749201012400366-2.web.ip-com.com.cn:8080"}},
	},{
		Cid: "jkgtvb", SN: "F6ACF73724F56BED", Vendor: "Tenda", Mode: "M30V2.6", SWVersion: "V16.01.0.6(2700)", MAC: "B0:DF:C1:F8:D2:B8", Addr: "183.94.242.60:3107",
		Attribution: "湖南-长沙-开福区", State: 10, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
		Tunnels: []Tunnel{{URLCgi: "http://F6ACF73724F56BED-1.web.ip-com.com.cn:8080"}, {URLCgi: "http://F6ACF73724F56BED-2.web.ip-com.com.cn:8080"}},
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

var againDevs = [...]Device{
	/*{// dev-tun1 ==> dev-tun2
		Cid: "pouwer", SN: "M749201012400366", Vendor: "IP-COM", Mode: "G3v3.6", SWVersion: "V15.11.0.17(9502)", MAC: "D8:38:0D:6F:67:38", Addr: "110.84.74.114:23185",
		Attribution: "湖南-长沙-芙蓉区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
		Tunnels: []Tunnel{{URLCgi: "agn-http://M749201012400366-1.web.ip-com.com.cn:8080"},{URLCgi: "agn-http://M749201012400366-2.web.ip-com.com.cn:8080"}},
	},*/ {// dev1-tun ==> dev2-tun, Cid和SN相同（增加不了）
		Cid: "jkgtvb", SN: "F6ACF73724F56BED", Vendor: "agn-Tenda", Mode: "agn-M30V2.6", SWVersion: "agn-V16.01.0.6(2700)", MAC: "agn-B0:DF:C1:F8:D2:B8", Addr: "agn-183.94.242.60:3107",
		Attribution: "agn-湖南-长沙-开福区", State: 10, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: true,
		Tunnels: []Tunnel{{URLCgi: "http://F6ACF73724F56BED-1.web.ip-com.com.cn:8080"},{URLCgi: "http://F6ACF73724F56BED-2.web.ip-com.com.cn:8080"}},
	}, {// dev1-tun ==> dev2-tun, Cid和SN也不相同（tun冲突，重点关注tun增加情况）
		Cid: "ndfdgr", SN: "K7LMF89452F26LMK-agn", Vendor: "agn-Tenda", Mode: "agn-M50V3.0", SWVersion: "agn-V19.01.1.6(2689)", MAC: "agn-C0:7K:1C:B8:J4:PK", Addr: "agn-211.56.124.70:2385",
		Attribution: "agn-湖南-长沙-雨花区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: false,
		Tunnels: []Tunnel{{URLCgi: "http://K7LMF89452F26LMK-1.web.ip-com.com.cn:8080"},{URLCgi: "http://K7LMF89452F26LMK-2.web.ip-com.com.cn:8080"}},
	}, /*{// dev1-tun1 ==> dev2-tun2, Cid和SN也不相同
		Cid: "ndfdgr-agn-agn", SN: "K7LMF89452F26LMK-agn-agn", Vendor: "agn-agn-Tenda", Mode: "agn-agn-M50V3.0", SWVersion: "agn-agn-V19.01.1.6(2689)", MAC: "agn-agn-C0:7K:1C:B8:J4:PK", Addr: "agn-agn-211.56.124.70:2385",
		Attribution: "agn-agn-湖南-长沙-雨花区", State: 0, CreateTime: time.Now().Format("2006/01/02 15:04:05"), LoginTime: time.Now().Format("2006/01/02 15:04:05"), IsHole: false,
		Tunnels: []Tunnel{{URLCgi: "agn-agn-http://K7LMF89452F26LMK-1.web.ip-com.com.cn:8080"},{URLCgi: "agn-agn-http://K7LMF89452F26LMK-2.web.ip-com.com.cn:8080"}},
	},*/
}

var updateDevs = [...]Device{
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

var appendDevs = [...]Device{
	{
		Cid: "pouwer", SN: "M749201012400366",
		Tunnels: []Tunnel{{URLCgi: "apd-http://M749201012400366-1.web.ip-com.com.cn:8080"},{URLCgi: "apd-http://M749201012400366-2.web.ip-com.com.cn:8080"}},
	},/*{
		Cid: "jkgtvb", SN: "F6ACF73724F56BED",
		Tunnels: []Tunnel{{URLCgi: "http://F6ACF73724F56BED-1.web.ip-com.com.cn:8080"}, {URLCgi: "http://F6ACF73724F56BED-2.web.ip-com.com.cn:8080"}},
	},	*//*{// dev -> dev-tun
		Cid: "mkhytg", SN: "MB07101412800096",
		Tunnels: []Tunnel{{URLCgi: "http://apd-MC07101412800052-1.web.ip-com.com.cn:8080"},{URLCgi: "http://apd-MC07101412800052-2.web.ip-com.com.cn:8080"}},
	}, *//*{// dev -> tun
		Cid: "acvbmb", SN: "1906747201001359",
		Tunnels: []Tunnel{{URLCgi: "http://apd-1906747201001359-1.web.ip-com.com.cn:8080"}},
	}, {// dev -> dev
		Cid: "huytto", SN: "2003938101002253",
		Tunnels: []Tunnel{{URLCgi: "http://apd-2003938101002253-1.web.ip-com.com.cn:8080"},{URLCgi: "http://apd-2003938101002253-2.web.ip-com.com.cn:8080"}},
	},*/
}

func TestInsertDevice(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()
	_ = DestroyModel(db)
	_ = MigrateModel(db)

	devices := originDevs
	for i := 0; i < len(devices); i++ {
		err := devices[i].InsertDevice(db)
		if err != nil {
			t.Log("InsertDevice error:", err)
		}
	}

/*	device, err := getDeviceInfo("tenda")
	if err != nil {
		t.Error("getDeviceInfo error")
		return
	}

	err = device.InsertDevice(db)
	if err != nil {
		t.Log("InsertDevice error:", err)
	}*/

	CloseDB(db)
}

func TestInsertDeviceAgain(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()
	_ = DestroyModel(db)
	_ = MigrateModel(db)


	devices := againDevs
	for i := 0; i < len(devices); i++ {
		err := devices[i].InsertDevice(db)
		if err != nil {
			t.Log("InsertDevice error:", err)
		}
	}

	CloseDB(db)
}

func TestUpdateDevice(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()

	devices := updateDevs
	for i := 1; i < len(devices); i++ {
		err := devices[i].UpdateDeviceByCID(db)
		if err != nil {
			t.Log("UpdateDeviceByID error:", err)
		}
	}

	var device = &Device{}
	device.Cid = updateDevs[0].Cid
	device.State = 321
	err := device.UpdateDeviceStateByCID(db)
	if err != nil {
		t.Log("UpdateDeviceStateByCID error:", err)
	}
	CloseDB(db)
}

func TestAppendDeviceAssociation(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()
	_ = MigrateModel(db)

	for	_, device := range appendDevs {
		dev := &device
		err := dev.AppendDeviceAssociation(db)
		if err != nil {
			t.Log("UpdateDeviceByID error:", err)
		}
	}

/*	for i := 0; i < 2; i++ {

		devices :=make([]Device, len(originDevs))
		copy(devices, originDevs)

		for _, device := range devices {
			dev := &device
			dev.Cid = strconv.Itoa(i + 1) + "-" + dev.Cid
			dev.SN = strconv.Itoa(i + 1)  + "-" + dev.SN
			dev.Tunnels[0].FkCid = dev.Cid
			dev.Tunnels[1].FkCid = dev.Cid
			//dev.Tunnels[0].URLCgi = strconv.Itoa(i + 1)  + "-" + dev.Tunnels[0].URLCgi
			//dev.Tunnels[1].URLCgi = strconv.Itoa(i + 1)  + "-" + dev.Tunnels[1].URLCgi
			_ = dev.AppendDeviceAssociation(db)
		}
		//fmt.Println("sleep")
		//time.Sleep(time.Second * 10)
	}*/

	CloseDB(db)
}

func TestQueryDevice(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()

	devices, err := QueryAllDevices(db, "")
	if err != nil {
		t.Log("QueryAllDevices error:", err)
	}
	for _, dev := range devices {
		fmt.Println(dev)
	}

	fmt.Println("=============================")

	devices, err = QueryOnlineDevices(db, "")
	if err != nil {
		t.Log("QueryOnlineDevices error:", err)
	}
	for _, dev := range devices {
		fmt.Println(dev)
	}

	fmt.Println("=============================")

	devices, err = QueryDevicesByVendor(db, "Tenda", "")
	if err != nil {
		t.Log("QueryOnlineDevices error:", err)
	}
	for _, dev := range devices {
		fmt.Println(dev)
	}
	fmt.Println("=============================")

	device := &Device{Cid: "pouwer"}
	err = device.QueryDevicesByClientID(db)
	if err != nil {
		t.Log("QueryOnlineDevices error:", err)
	}
	fmt.Println(*device)

	CloseDB(db)
}

func TestDeleteOffLineDevicesByDay(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()

	t.Log("All Devices:")
	devices, _ := QueryAllDevices(db, "")
	for _, dev := range devices {
		fmt.Println(dev)
	}

	t.Log("DeleteOffLineDevicesByDay")
	err := DeleteOffLineDevicesByDay(db, 0)
	if err != nil {
		t.Log("DeleteOffLineDevicesByDay error:", err)
	}

	t.Log("All Devices:")
	devices, _ = QueryAllDevices(db, "")
	for _, dev := range devices {
		fmt.Println(dev)
	}
	CloseDB(db)
}

func TestDeleteDeviceByCID(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()

	t.Log("All Devices:")
	devices, _ := QueryAllDevices(db, "")
	for _, dev := range devices {
		fmt.Println(dev)
	}
	device := &Device{Cid: "jkgtvb"/*, SN: "M749201012400366"*/}
	err := device.DeleteDeviceByCID(db)
	if err != nil {
		t.Log("DeleteDeviceByCID error:", err)
	}

	t.Log("All Devices:")
	devices, _ = QueryAllDevices(db, "")
	for _, dev := range devices {
		fmt.Println(dev)
	}
	CloseDB(db)
}

func TestDeletedDeviceAssociation(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()



	CloseDB(db)
}

func TestDevice_DeletedDeviceAssociation(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()

	t.Log("All Devices:")
	devices, _ := QueryAllDevices(db, "")
	for _, dev := range devices {
		fmt.Println(dev)
	}

	dev := &Device {   // dev-tun -> dev-tun
		Tunnels: []Tunnel{{URLCgi: "http://apd-MC07101412800052-1.web.ip-com.com.cn:8080"},{URLCgi: "http://upt-MC07101412800052-1.web.ip-com.com.cn:8080"}},
	}
	err := dev.DeletedDeviceAssociation(db)
	if err != nil {
		t.Log("DeletedDeviceAssociation error:", err)
	}

	t.Log("All Devices:")
	devices, _ = QueryAllDevices(db, "")
	for _, dev := range devices {
		fmt.Println(dev)
	}

	CloseDB(db)
}

func TestDestroyModel(t *testing.T) {
	_ = system.LoadDbConfiguration("..\\config\\config.yaml")
	db, _ := OpenDb()

	_ = DestroyModel(db)

	CloseDB(db)
}
//===============================并发测试===========================================
type Control struct {
	IsTunnel bool
	State    bool
}

type ControlClient struct {
	controls map[string]*Control
	sync.RWMutex
}

type Vendor struct {
	Name   string
	DevNum int
}

var db  *gorm.DB
const DeviceNumMax = 100
var ctlClient *ControlClient
const RunTime = 5 // 分钟
var stop int64

var gAttributions = []string {
	"湖南-长沙-岳麓区","湖南-长沙-天心区","湖南-长沙-开福区","湖南-长沙-雨花区",
	"广东-深圳-南山区","广东-深圳-福田区","广东-深圳-盐田区","广东-深圳-龙华区","广东-深圳-龙岗区", "广东-深圳-宝安区",
}
var gMode = []string {
	"G2v3.5", "G3v6.6","G6v6.2", "G8v8.6",
	"M30V2.6", "M50V1.8", "M60V2.6", "M80V6.0",
	"W15Ev2", "W16Ev1", "W18Ev5", "W18Ev5",
}
var gVendor = []string{"Tenda", "IP-COM", "HaiKang"}

func TestConcurrencyInsert(t *testing.T) {
	var err error
	var wg sync.WaitGroup
	if err = system.LoadDbConfiguration("..\\config\\config.yaml"); err != nil {
		panic(errors.New("LoadDbConfiguration err:" + err.Error()))
	}

	db, err = OpenDb()
	if err != nil {
		panic(errors.New("OpenDb err:" + err.Error()))
	}

	_ = DestroyModel(db)
	if err = MigrateModel(db); err != nil {
		panic(errors.New("MigrateModel err:" + err.Error()))
	}

	defer func() {
		//_ = DestroyModel(db)
		CloseDB(db)
	}()

	seed, err := randomSeed()
	if err != nil {
		panic(errors.New("RandomSeed err:" + err.Error()))
	}
	mrand.Seed(seed)
	ctlClient = newControlClient()
	stop = time.Now().Unix() + RunTime * 60

	wg.Add(1)
	go testInsert(t, &wg)
	time.Sleep(time.Second * 2)
/*	wg.Add(1)
	go testQuery(t, &wg)*/
	wg.Add(1)
	go testUpdate(t, &wg)
	wg.Add(1)
	go testDelete(t, &wg)

	wg.Wait()
	t.Log("all goroutine finish")
}

func testInsert(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()

	var ch = make(chan *Vendor)
	go func() {
		itr := len(gVendor)
		for i := 0; i < DeviceNumMax && time.Now().Unix() < stop; i += itr {
			for j, ven := range gVendor {
				ch <- &Vendor{Name: ven, DevNum: i + j + 1}
			}
		}
		close(ch)  // close 后发生什么，会终止 for range channel 循环
	}()

	var iWg sync.WaitGroup
	for v := range ch {
		iWg.Add(1)
		go func(dev *Vendor) {
			defer iWg.Done()
			device, err := getDeviceInfo(dev.Name)
			if err != nil || len(device.Cid) == 0 {
				t.Error("getDeviceInfo error")
				return
			}
			var clintID = make(chan string)

			// 使用单独的协程去保存
			go saveDeviceInfo(t, device, dev.DevNum, clintID)
			// 注意：tunnels为从表，devices为主表，只有设备数据被插入主表后，才能将关联数据插入从表，否则报错；这里使用通道同步两个协成
			cid, ok := <- clintID
			if ok {
				iWg.Add(1)
				go appendDeviceAssociation(t, cid, &iWg)
				t.Log("appendDeviceAssociation finished!")
			} else {
				t.Log("appendDeviceAssociation error!")
			}
			close(clintID)
		}(v)
	}
	iWg.Wait()
	t.Log("testInsert finished!")
}

func testQuery(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()

	var qWg sync.WaitGroup
	qWg.Add(1)
	go func() {
		defer qWg.Done()
		for {
			devices, err := QueryAllDevices(db, "")
			if err != nil {
				panic(errors.New("QueryAllDevices error:" + err.Error()))
			}
			t.Log(fmt.Errorf("QueryAllDevices: %p", &devices))

			if time.Now().Unix() > stop {
				break
			}
			time.Sleep(time.Second * 2)
		}
	}()

	qWg.Add(1)
	go func() {
		defer qWg.Done()
		for {
			devices, err := QueryOnlineDevices(db, "")
			if err != nil {
				panic(errors.New("QueryOnlineDevices error:" + err.Error()))
			}
			t.Log(fmt.Errorf("QueryOnlineDevices: %p", &devices))

			if time.Now().Unix() > stop {
				break
			}
			time.Sleep(time.Second * 2)
		}
	}()

	qWg.Add(1)
	go func() {
		defer qWg.Done()
		for {
			for _,  ven := range gVendor {
				devices, err := QueryDevicesByVendor(db, ven, "")
				if err != nil {
					panic(errors.New("QueryDevicesByVendor error:" + err.Error()))
				}
				t.Log(fmt.Errorf("query devices by vendor[%s]: %p", ven, &devices))
				time.Sleep(time.Second * 2)
			}

			if time.Now().Unix() > stop {
				break
			}
			time.Sleep(time.Second * 2)
		}
	}()

	qWg.Add(1)
	go func() {
		defer qWg.Done()
		for {
			ctlClient.RLock()
			for cid, _ := range ctlClient.controls {
				device := &Device{Cid: cid}
				err := device.QueryDevicesByClientID(db)
				ctlClient.RUnlock()
				if err != nil {
					panic(errors.New("QueryDevicesByClientID error:" + err.Error()))
				}

				t.Log(fmt.Errorf("query devices by cid[%s]-sn[%s]", cid, device.SN))
				if time.Now().Unix() > stop {
					break
				}
				time.Sleep(time.Second * 2)
				ctlClient.RLock()
			}
			ctlClient.RUnlock()
			time.Sleep(time.Second * 2)
		}
	}()
	qWg.Wait()
	t.Log("testUpdate finished!")
}

func testUpdate(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	var uWg sync.WaitGroup

	uWg.Add(1)
	go func() {
		defer uWg.Done()
		for {
			for _, vendor := range gVendor {
				device, err := getDeviceInfo(vendor)
				if err != nil {
					panic(errors.New("getDeviceInfo" + err.Error()))
				}
				tunnels, err := getTunnelInfo()
				if err != nil {
					panic(errors.New("getTunnelInfo" + err.Error()))
				}
				device.Tunnels = tunnels

				ctlClient.Lock()
				num := mrand.Intn(len(ctlClient.controls))
				idx := 0
				for cid, c := range ctlClient.controls {
					if idx == num {
						if c.State {
							device.Cid = cid
							break
						}
						num++  // 如果不在线，取下一个
					}
					idx++
				}
				if err = device.UpdateDeviceByCID(db); err != nil {
					ctlClient.Unlock()
					panic(errors.New("UpdateDeviceByID" + err.Error()))
				}
				ctlClient.controls[device.Cid].IsTunnel = true
				ctlClient.Unlock()

				t.Log("id[",device.Cid,"] device info updated")
				if time.Now().Unix() > stop {
					break
				}
				time.Sleep(time.Second * 15)
			}
			if time.Now().Unix() > stop {
				break
			}
			time.Sleep(time.Second * 15)
		}
	}()

	uWg.Add(1)
	go func() {
		defer uWg.Done()
		for {
			device := &Device{State: uint(mrand.Intn(366))}

			ctlClient.Lock()
			num := mrand.Intn(len(ctlClient.controls))
			idx := 0
			for cid, _ := range ctlClient.controls {
				if idx == num {
					device.Cid = cid
				}
				idx++
			}
			if err := device.UpdateDeviceStateByCID(db); err != nil {
				ctlClient.Unlock()
				panic(errors.New("UpdateDeviceStateByCID" + err.Error()))
			}
			ctlClient.controls[device.Cid].State = true
			ctlClient.Unlock()

			t.Log("id[",device.Cid,"] device off-line")
			if time.Now().Unix() > stop {
				break
			}
			time.Sleep(time.Second * 10)
		}
	}()

	uWg.Wait()
	t.Log("testUpdate finished!")
}

func testDelete(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
}

//func createDevice(t *testing.T, vendor string, num int, clintID chan string) () {
func saveDeviceInfo(t *testing.T, device *Device, num int, clintID chan string) {
	control := &Control{IsTunnel: false, State: true}

	if ctlClient.get(device.Cid) != nil {
		if err := device.UpdateDeviceByCID(db); err != nil {
			t.Errorf("[%d]st [%s]device err >> UpdateDeviceByCID err >> %s", num, device.Vendor, err.Error())
			return
		}
		if err := ctlClient.del(device.Cid); err != nil {
			t.Errorf("delete client control err:" + err.Error())
			t.Errorf("[%d]st [%s]device err >> del client ctl err >> %s", num, device.Vendor, err.Error())

			return
		}
		t.Log("[",num,"]st", device.Vendor, "device updated")
	} else {
		if err := device.InsertDevice(db); err != nil {
			t.Errorf("[%d]st [%s]device err >> InsertDevice err >> %s", num, device.Vendor, err.Error())
			return
		}
		t.Log("[",num,"]st", device.Vendor, "device created")
	}

	ctlClient.add(device.Cid, control)
	clintID <- device.Cid
}

func appendDeviceAssociation(t *testing.T, cid string, wg *sync.WaitGroup) {
	defer wg.Done()
	var err error
	var device = &Device{}
	device.Tunnels, err = getTunnelInfo()
	if err != nil {
		t.Error("getTunnelInfo error")
		return
	}

	device.Cid = cid
	if err = device.AppendDeviceAssociation(db); err != nil {
		t.Error("AppendDeviceAssociation error")
		return
	}
	ctlClient.Lock()
	ctlClient.controls[cid].IsTunnel = true
	ctlClient.Unlock()
	t.Log("[",cid,"]device association appended")
}

func getDeviceInfo(vendor string) (device *Device, err error) {
	device = &Device {
		Vendor: vendor,
		Mode: gMode[mrand.Intn(len(gMode))],  // 生成 [0,n)区间的一个随机数（注意：不包括n）
		Attribution: gAttributions[mrand.Intn(len(gAttributions))],
		State: 0,
		CreateTime: time.Now().Format("2006/01/02 15:04:05"),
		LoginTime: time.Now().Format("2006/01/02 15:04:05"),
		IsHole: true,
		//Tunnels: make([]Tunnel, 2),
	}

	if device.Cid, err = secureRandId(4); err != nil {
		return nil, errors.New("get id error")
	}

	if device.SN, err = secureRandId(8); err != nil {
		return nil, errors.New("get sn error")
	}
	if vendor == "Tenda" {
		device.SN = "TE" + device.SN
	} else if vendor == "IP-COM" {
		device.SN = "IC" + device.SN
	} else if vendor == "HaiKang" {
		device.SN = "HK" + device.SN
	} else {
		device.SN = "OT" + device.SN
	}

	mac, err := secureRandId(12)
	if err != nil {
		return nil, errors.New("get mac error")
	}
	device.MAC = mac[0:2] + ":" + mac[2:4] + ":" + mac[4:6] + ":" + mac[6:8] + ":" + mac[8:10] + ":" + mac[10:12] // "D8:38:0D:6F:67:38"
	device.Addr = strconv.Itoa(mrand.Intn(255)) + "." + strconv.Itoa(mrand.Intn(255)) + "." +
		          strconv.Itoa(mrand.Intn(255)) + "." + strconv.Itoa(mrand.Intn(255)) + ":" + strconv.Itoa(mrand.Intn(65535))  // "110.84.74.114:23185",
	device.SWVersion = "V" + strconv.Itoa(mrand.Intn(30)) + "." + strconv.Itoa(mrand.Intn(20)) + "." +
		               "01" + "." + strconv.Itoa(mrand.Intn(100)) + "(" + strconv.Itoa(mrand.Intn(65535)) + ")"
	return
}

func getTunnelInfo() (tunnels []Tunnel, err error) {
	var str string
	str, err = secureRandId(4)
	if err != nil {
		return nil, errors.New("get id error")
	}

	tunnels = []Tunnel {
		{URLCgi: "http://" + str[:4] + ".web.ip-com.com.cn:8080"},
		{URLCgi: "http://" + str[4:] + ".web.ip-com.com.cn:8080"},
	}
	return
}

func randomSeed() (seed int64, err error) {
	err = binary.Read(rand.Reader, binary.LittleEndian, &seed)
	return
}

func secureRandId(idlen int) (id string, err error) {
	b := make([]byte, idlen)
	n, err := rand.Read(b)

	if n != idlen {
		err = fmt.Errorf("Only generated %d random bytes, %d requested", n, idlen)
		return
	}

	if err != nil {
		return
	}

	id = fmt.Sprintf("%x", b)
	return
}

func newControlClient() *ControlClient {
	return &ControlClient{
		controls: make(map[string]*Control),
	}
}


func (c *ControlClient) get(clientId string) *Control {
	c.RLock()
	defer c.RUnlock()
	return c.controls[clientId]
}

func (c *ControlClient) add(clientId string, new *Control) (old *Control){
	c.Lock()
	defer c.Unlock()

	old = c.controls[clientId]
	if old != nil {
		delete(c.controls, clientId)
	}
	c.controls[clientId] = new
	return
}

func (c *ControlClient) del(clientId string) error {
	c.Lock()
	defer c.Unlock()
	if c.controls[clientId] == nil {
		return fmt.Errorf("no found client, id: %s", clientId)
	} else {
		delete(c.controls, clientId)
		return nil
	}
}
