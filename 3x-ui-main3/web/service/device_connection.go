package service

import (
	"sync"
	"time"
	"x-ui/database/model"
	"x-ui/logger"
)

type DeviceConnectionService struct {
	connections sync.Map // map[string]map[string]model.DeviceConnection // email -> deviceId -> connection
}

var deviceConnectionService *DeviceConnectionService
var deviceConnectionServiceOnce sync.Once

func GetDeviceConnectionService() *DeviceConnectionService {
	deviceConnectionServiceOnce.Do(func() {
		deviceConnectionService = &DeviceConnectionService{}
	})
	return deviceConnectionService
}

func (s *DeviceConnectionService) AddConnection(email string, deviceId string, ip string) bool {
	// Получаем текущие подключения для email
	connectionsRaw, _ := s.connections.LoadOrStore(email, make(map[string]model.DeviceConnection))
	connections := connectionsRaw.(map[string]model.DeviceConnection)

	// Проверяем лимит устройств
	client, err := GetInboundService().GetClientByEmail(email)
	if err != nil {
		logger.Warning("Failed to get client info:", err)
		return false
	}

	if client.MaxDevices > 0 && len(connections) >= client.MaxDevices {
		// Если достигнут лимит устройств, отклоняем новое подключение
		return false
	}

	// Добавляем новое подключение
	connections[deviceId] = model.DeviceConnection{
		ClientEmail: email,
		DeviceID:    deviceId,
		IP:          ip,
		LastSeen:    time.Now().Unix(),
	}

	return true
}

func (s *DeviceConnectionService) RemoveConnection(email string, deviceId string) {
	if connectionsRaw, ok := s.connections.Load(email); ok {
		connections := connectionsRaw.(map[string]model.DeviceConnection)
		delete(connections, deviceId)
		if len(connections) == 0 {
			s.connections.Delete(email)
		}
	}
}

func (s *DeviceConnectionService) GetConnections(email string) []model.DeviceConnection {
	if connectionsRaw, ok := s.connections.Load(email); ok {
		connections := connectionsRaw.(map[string]model.DeviceConnection)
		result := make([]model.DeviceConnection, 0, len(connections))
		for _, conn := range connections {
			result = append(result, conn)
		}
		return result
	}
	return nil
}

func (s *DeviceConnectionService) UpdateLastSeen(email string, deviceId string) {
	if connectionsRaw, ok := s.connections.Load(email); ok {
		connections := connectionsRaw.(map[string]model.DeviceConnection)
		if conn, ok := connections[deviceId]; ok {
			conn.LastSeen = time.Now().Unix()
			connections[deviceId] = conn
		}
	}
}

// Периодическая очистка старых подключений
func (s *DeviceConnectionService) CleanupOldConnections() {
	for {
		time.Sleep(5 * time.Minute)
		now := time.Now().Unix()
		s.connections.Range(func(key, value interface{}) bool {
			email := key.(string)
			connections := value.(map[string]model.DeviceConnection)
			
			for deviceId, conn := range connections {
				// Удаляем подключения, которые не обновлялись более 5 минут
				if now-conn.LastSeen > 300 {
					delete(connections, deviceId)
				}
			}
			
			if len(connections) == 0 {
				s.connections.Delete(email)
			}
			return true
		})
	}
} 