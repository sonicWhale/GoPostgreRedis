package main

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var DB *sql.DB
var Myc MyConfig

// подключаемся к БД по данным из конфигурации toml
func init() {

	if _, err := toml.DecodeFile("myconf.toml", &Myc); err != nil {
		fmt.Println(err)
		return
	}

	var err error
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", Myc.DBuser, Myc.DBpassword, Myc.DBname)
	DB, err = sql.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}
	//defer DB.Close() //базу не закрываем она нам еще понадобится

}

//получаем все записи из таблицы устройств
func GetAllDevicesFromDB(out chan DevicesStruct) {

	rows, err := DB.Query("SELECT * FROM devices")
	if err != nil {
		//panic(err)
	}
	defer rows.Close()

	go func() {
		for rows.Next() {
			var newDevice DevicesStruct
			err := rows.Scan(&newDevice.Id, &newDevice.Name, &newDevice.Userid)
			if err != nil {
				//panic(err)
			}
			log.Println(newDevice)
			out <- newDevice
		}
	}()

}

func CreateMetric(d DevicesStruct) DevicesMetricStruct {

	var newMetric DevicesMetricStruct

	//получаем уникальный ID метрики для БД
	newMetric.Id = TableIDs("device_metrics")
	newMetric.Deviceid = d.Id
	//генерируем случайные значения метрик
	for i := 0; i < len(newMetric.Metric); i++ {
		newMetric.Metric[i] = rand.Intn(50)
	}
	newMetric.LocalTime = time.Now().AddDate(0, 0, -1)
	newMetric.ServerTime = time.Now()
	log.Println(newMetric)

	//записываем метрику в БД
	var stringQ = "INSERT INTO device_metrics (Id, device_Id, metric_1, metric_2, metric_3, metric_4, metric_5, local_time, server_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	_, err := DB.Exec(stringQ,
		newMetric.Id,
		newMetric.Deviceid,
		newMetric.Metric[0],
		newMetric.Metric[1],
		newMetric.Metric[2],
		newMetric.Metric[3],
		newMetric.Metric[4],
		newMetric.LocalTime,
		newMetric.ServerTime)
	if err != nil {
		fmt.Println(err.Error())
		//return
	}

	return newMetric
}

//проверка метрик
func checkMetrics(m DevicesMetricStruct) {

	var newAlert DeviceAlertStruct
	for i := 0; i < len(m.Metric); i++ {
		//если одно из значений метрик плохое создаем новый Alert
		if m.Metric[i] == Myc.BadMetricParam {
			go func(m DevicesMetricStruct) {
				newAlert.Id = TableIDs("device_alerts")
				newAlert.Deviceid = m.Deviceid
				newAlert.Message = "Bad metric param on device " + strconv.Itoa(m.Deviceid)
				// Пишем в Redis по ключу deviceid (будет перезаписан при наличии)
				setValues(newAlert.Deviceid, newAlert.Message)
				// Отправка из Redis
				SendEmail("test0151@yandex.ru", getValues(newAlert.Deviceid))
				log.Println(getValues(newAlert.Deviceid)) //для проверки
				//пишем алерты в БД
				_, err := DB.Exec("INSERT INTO device_alerts (id, device_id, message) VALUES ($1, $2, $3)", newAlert.Id, newAlert.Deviceid, newAlert.Message)
				if err != nil {
					fmt.Println(err.Error())
					//return
				}
			}(m)
		}
	}
}

//пуолчаем уникальный ID для таблицы
func TableIDs(nameT string) (lastID int) {
	stringQ := "SELECT COUNT(ID) FROM " + nameT + ";"
	rows, err := DB.Query(stringQ)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&lastID)
		if err != nil {
			fmt.Println(err.Error())
			//return
		}
	}
	if err = rows.Err(); err != nil {
		fmt.Println(err.Error())
	}

	lastID++
	return lastID
}
