package main

import (
	"encoding/json"
	"net/http"

	"github.com/adamaho/realtime/pkg/realtime"
	"github.com/labstack/echo/v4"
)

const SESSION_ID = "count"

type data struct {
	Count int `json:"count"`
}

func main() {
	var count data = data{0}

	e := echo.New()
	rt := realtime.New()

	e.GET("/count", func(c echo.Context) error {
		json, _ := json.Marshal(count)
		return rt.Response(c.Response().Writer, c.Request(), json, SESSION_ID, realtime.ResponseOptions())
	})

	e.POST("/count/increment", func(c echo.Context) error {
		count.Count += 1
		json, _ := json.Marshal(count)
		rt.SendMessage(json, SESSION_ID)
		return c.String(http.StatusOK, "count updated.")
	})

	e.POST("/count/decrement", func(c echo.Context) error {
		count.Count -= 1
		json, _ := json.Marshal(count)
		rt.SendMessage(json, SESSION_ID)
		return c.String(http.StatusOK, "count updated.")
	})

	e.Logger.Fatal(e.StartTLS(":3000", "cert.pem", "key.pem"))
}
