package system

import (
    "fmt"
    "testing"
)

func TestSystem(t *testing.T) {
    err := LoadDbConfiguration("..\\config\\config.yaml")
    if err != nil {
        t.Log("load db config error:", err)
        return
    }

    config := GetDbConfiguration()
    fmt.Println(*config)
}
