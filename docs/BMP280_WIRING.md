# Cableado BMP280

| BMP280 | Raspberry Pi Zero 2 W |
|---|---|
| VCC/VIN compatible con 3.3 V | 3.3 V |
| GND | GND |
| SDA | GPIO 2 / SDA |
| SCL | GPIO 3 / SCL |

Verifique la hoja del modulo concreto: no todos incorporan regulador o adaptacion de nivel y no debe asumirse tolerancia a 5 V. El driver acepta BMP280 ID `0x58` en `0x76` o `0x77`; identifica y rechaza BME280 ID `0x60`.
