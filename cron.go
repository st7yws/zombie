// 定时任务,每秒请求一次Open Hardwar Monitor,并把数据缓存在redis中
// Copyright (C) 2020  Thenagi<thenagi@ruiko.net>  https://www.thenagi.com/
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"github.com/bitly/go-simplejson"
	gocron "github.com/robfig/cron"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// initCron 初始化定时任务
func initCron() {
	httpClient := http.Client{Timeout: 1 * time.Second}
	cron := gocron.New()

	// 每秒请求一次Open Hardware Monitor的数据并存入redis
	if errGetData := cron.AddFunc("*/1 * * * * ?", func() {
		if resp, httpGetErr := httpClient.Get("http://192.168.1.100:8085/data.json"); httpGetErr == nil {
			//noinspection GoUnhandledErrorResult
			defer resp.Body.Close()

			if body, readBodyErr := ioutil.ReadAll(resp.Body); readBodyErr == nil {
				if data, err := simplejson.NewJson(body); err == nil {
					name, _ := data.Get("Children").GetIndex(0).Get("Text").String()
					info := data.Get("Children").GetIndex(0).Get("Children")
					cpu := info.GetIndex(0)
					mem := info.GetIndex(1)
					gpu := info.GetIndex(2)

					cpuTemp, _ := cpu.Get("Children").GetIndex(1).Get("Children").GetIndex(4).Get("Value").String()
					memLoad, _ := mem.Get("Children").GetIndex(0).Get("Children").GetIndex(0).Get("Value").String()
					gpuTemp, _ := gpu.Get("Children").GetIndex(1).Get("Children").GetIndex(0).Get("Value").String()
					gpuFan, _ := gpu.Get("Children").GetIndex(3).Get("Children").GetIndex(0).Get("Value").String()

					redisClient.HSet("nio_sensors", "name", name)
					redisClient.HSet("nio_sensors", "cpu_temp", strings.Replace(cpuTemp, " °C", "", 1))
					redisClient.HSet("nio_sensors", "mem_load", strings.Replace(memLoad, " %", "", 1))
					redisClient.HSet("nio_sensors", "gpu_temp", strings.Replace(gpuTemp, " °C", "", 1))
					redisClient.HSet("nio_sensors", "gpu_fan", strings.Replace(gpuFan, " RPM", "", 1))
				} else {
					errHandle()
				}
			} else {
				errHandle()
			}
		} else {
			errHandle()
		}
	}); errGetData != nil {
		errHandle()
	}

	cron.Start()
}

// errHandle 定时任务执行错误时向redis写入空数据
func errHandle() {
	redisClient.HSet("nio_sensors", "cpu_temp", "--.-")
	redisClient.HSet("nio_sensors", "mem_load", "--.-")
	redisClient.HSet("nio_sensors", "gpu_temp", "--.-")
	redisClient.HSet("nio_sensors", "gpu_fan", "----")
}
