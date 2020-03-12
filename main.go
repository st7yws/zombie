// 读取数据并显示到OLED上
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
	"fmt"
	"time"

	graphicsEngine "github.com/fogleman/gg"
	"github.com/go-redis/redis/v7"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"

	"zombie/bitmapfont"
)

type stage struct {
	Name string
	Time int
	Work func()
}

var textHeight = 12 // 文本高度
var oledWidth = 128 // oled屏幕宽度
var oledHeight = 64 // oled屏幕高度

var refeshTime = 1 * time.Second // 屏幕刷新时间
var stages []stage               // 场景
var currentStageIndex = 0        // 当前场景序号
var n = 0                        // 刷新计次器

var onChangeStyle = true

var geContext *graphicsEngine.Context // graphicsEngine 上下文环境
var redisClient *redis.Client         // redis客户端

// main 主函数
func main() {
	redisClient = redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})
	//noinspection GoUnhandledErrorResult
	defer redisClient.Close()

	initCron()

	r := raspi.NewAdaptor()
	oled := i2c.NewSSD1306Driver(r)
	sht3x := i2c.NewSHT3xDriver(r)

	geContext = graphicsEngine.NewContext(oledWidth, oledHeight)
	geContext.SetFontFace(bitmapfont.Zpix)
	textHeight = int(geContext.FontHeight())

	stages = append(stages, stage{Name: "Ruiko Info", Time: 45, Work: ruikoStageHandler})
	stages = append(stages, stage{Name: "RaspPi Info", Time: 30, Work: raspPiStageHandler})
	stages = append(stages, stage{Name: "Room Temp Info", Time: 15, Work: roomTempStageHandler})

	stageMaxIndex := len(stages) - 1

	work := func() {
		oled.Clear()
		_ = oled.Display()

		_ = sht3x.Start()

		gobot.Every(refeshTime, func() {
			now := time.Now()

			if now.Unix()%10 == 0 {
				if temp, rh, err := sht3x.Sample(); err == nil {
					redisClient.HSet("nio_sensors", "room_temp", fmt.Sprintf("%.1f", temp))
					redisClient.HSet("nio_sensors", "room_rh", fmt.Sprintf("%.1f", rh))
				} else {
					redisClient.HSet("nio_sensors", "room_temp", "--.-")
					redisClient.HSet("nio_sensors", "room_rh", "--.-")
				}
			}

			geContext.SetRGB(0, 0, 0)
			geContext.Clear()
			geContext.SetRGB(1, 1, 1)

			if n == 0 {
				onChange(oled)
			} else {
				currentStage := stages[currentStageIndex]
				currentStage.Work()
				if n == currentStage.Time+1 {
					if currentStageIndex == stageMaxIndex {
						currentStageIndex = 0
					} else {
						currentStageIndex++
					}

					n = -1
				}

				_ = oled.ShowImage(geContext.Image())
			}

			n++
		})
	}

	_ = gobot.NewRobot("ZombieSensors", []gobot.Connection{r}, []gobot.Device{oled, sht3x}, work).Start()
}

// drawTitle 绘制标题
func drawTitle(title string) {
	geContext.DrawRectangle(0, 0, 128, 13)
	geContext.Fill()
	geContext.SetRGB(0, 0, 0)
	geContext.DrawStringWrapped(fmt.Sprintf("-- %s --", title), 0, 13, 0, 1, float64(oledWidth), 1, graphicsEngine.AlignCenter)
	geContext.SetRGB(1, 1, 1)
}

// countTextLinePoint 转换行数到上下文环境的像素点
func countTextLinePoint(line int) float64 {
	return float64((line * textHeight) + 4)
}

// drawText 在上下文环境中绘制文本到某行
func drawText(text string, line int) {
	geContext.DrawStringAnchored(text, 4, countTextLinePoint(line), 0, 1)
}

// onChange 切换场景动作,防止OLED像素点长时间不变导致烧屏
func onChange(oled *i2c.SSD1306Driver) {
	oled.Clear()

	if onChangeStyle {
		for x := 0; x < oledWidth; x += 4 {
			for y := 0; y < oledHeight; y++ {
				oled.Set(x, y, 1)
			}
		}
	} else {
		for y := 0; y < oledHeight; y += 4 {
			for x := 0; x < oledWidth; x++ {
				oled.Set(x, y, 1)
			}
		}
	}

	onChangeStyle = !onChangeStyle

	_ = oled.Display()
}

func ruikoStageHandler() {
	sensors := redisClient.HGetAll("nio_sensors").Val()

	drawTitle(sensors["name"])
	drawText(fmt.Sprintf("CPU 温度: %-5s℃", sensors["cpu_temp"]), 1)
	drawText(fmt.Sprintf("MEM 已用: %-5s％", sensors["mem_load"]), 2)
	drawText(fmt.Sprintf("GPU 温度: %-5s℃", sensors["gpu_temp"]), 3)
	drawText(fmt.Sprintf("GPU 风扇: %-5sRPM", sensors["gpu_fan"]), 4)
}

func raspPiStageHandler() {
	memInfo, _ := mem.VirtualMemory()
	tempInfo, _ := host.SensorsTemperatures()

	drawTitle("树莓派")
	drawText(fmt.Sprintf("CPU 温度: %-5.1f℃", tempInfo[0].Temperature), 1)
	drawText(fmt.Sprintf("MEM 总计: %-5dＭ", memInfo.Total/1024/1024), 2)
	drawText(fmt.Sprintf("MEM 可用: %-5dＭ", memInfo.Free/1024/1024), 3)
	drawText(fmt.Sprintf("MEM 已用: %-5.1f％", memInfo.UsedPercent), 4)
}

func roomTempStageHandler() {
	temp := redisClient.HGet("nio_sensors", "room_temp").Val()
	rh := redisClient.HGet("nio_sensors", "room_rh").Val()

	drawTitle("室温")
	drawText(fmt.Sprintf("温度: %-5s℃", temp), 1)
	drawText(fmt.Sprintf("湿度: %-5sRH", rh), 3)
}
