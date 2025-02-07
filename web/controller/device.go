package controller

import (
	"x-ui/web/service"

	"github.com/gin-gonic/gin"
)

type DeviceController struct {
	deviceService service.DeviceConnectionService
}

func NewDeviceController(g *gin.RouterGroup) *DeviceController {
	a := &DeviceController{}
	a.initRouter(g)
	return a
}

func (a *DeviceController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/device")

	g.POST("/list/:email", a.getDevices)
	g.POST("/disconnect/:email/:deviceId", a.disconnectDevice)
}

func (a *DeviceController) getDevices(c *gin.Context) {
	email := c.Param("email")
	devices := service.GetDeviceConnectionService().GetConnections(email)
	jsonObj(c, devices, nil)
}

func (a *DeviceController) disconnectDevice(c *gin.Context) {
	email := c.Param("email")
	deviceId := c.Param("deviceId")
	service.GetDeviceConnectionService().RemoveConnection(email, deviceId)
	jsonMsg(c, "Device disconnected", nil)
} 