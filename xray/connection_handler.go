package xray

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"

	"x-ui/logger"
	"x-ui/web/service"
)

type ConnectionLog struct {
	Email    string `json:"email"`
	DeviceID string `json:"id"`
	IP       string `json:"ip"`
}

func StartConnectionHandler() {
	go func() {
		deviceService := service.GetDeviceConnectionService()
		go deviceService.CleanupOldConnections()

		for {
			time.Sleep(1 * time.Second)
			
			accessLogPath, err := GetAccessLogPath()
			if err != nil {
				logger.Warning("Failed to get access log path:", err)
				continue
			}

			file, err := os.Open(accessLogPath)
			if err != nil {
				logger.Warning("Failed to open access log:", err)
				continue
			}

			// Перемещаемся в конец файла
			_, err = file.Seek(0, 2)
			if err != nil {
				logger.Warning("Failed to seek to end of file:", err)
				file.Close()
				continue
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				
				// Парсим JSON лог
				var logEntry map[string]interface{}
				err := json.Unmarshal([]byte(line), &logEntry)
				if err != nil {
					continue
				}

				// Проверяем, что это лог подключения
				if logEntry["email"] == nil || logEntry["id"] == nil {
					continue
				}

				email := logEntry["email"].(string)
				deviceId := logEntry["id"].(string)
				ip := ""

				// Извлекаем IP из строки подключения
				if logEntry["dest"] != nil {
					dest := logEntry["dest"].(string)
					parts := strings.Split(dest, ":")
					if len(parts) > 0 {
						ip = parts[0]
					}
				}

				// Добавляем подключение
				deviceService.AddConnection(email, deviceId, ip)
			}

			file.Close()
		}
	}()
} 