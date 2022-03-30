package system

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
)

type DbConfiguration struct {
    UserName      string     `yaml:"UserName"`
    PassWord      string     `yaml:"PassWord"`
    Host          string     `yaml:"Host"`
    Port          int        `yaml:"Port"`
    DbName        string     `yaml:"Dbname"`
    DSN           string     `yaml:"Dsn"`
    MaxIdleConns  int        `yaml:"MaxIdleConns"`
    MaxOpenConns  int        `yaml:"MaxOpenConns"`
}

var dbConfiguration *DbConfiguration

func LoadDbConfiguration (path string) error {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return err
    }

    var dbConfig DbConfiguration
    err = yaml.Unmarshal(data, &dbConfig)
    if err != nil {
        return err
    }

    dbConfiguration = &dbConfig
    return err
}

func GetDbConfiguration() *DbConfiguration {
    return dbConfiguration
}