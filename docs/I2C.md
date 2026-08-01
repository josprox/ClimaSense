# I2C

`dtparam=i2c_arm=on` activa el controlador y Raspberry Pi OS expone `/dev/i2c-1`. El plugin configura `I2C_SLAVE`, `I2C_TIMEOUT` e `I2C_RETRIES`, serializa cada transaccion y rechaza direcciones fuera de 7 bits.

API Joss:

```joss
$devices = ClimaSenseHardware::i2cScan(1)
$id = ClimaSenseHardware::i2cReadRegister(1, 0x76, 0xD0, 1, 1000)
$reading = ClimaSenseHardware::bmp280Read(1, "auto", {"pressure_oversampling": 4, "filter": 2})
```

El servicio se ejecuta con el usuario dedicado `climasense`, que pertenece al grupo `i2c`; la aplicacion completa no corre como root. Un resultado `76` o `77` de `i2cdetect -y 1` confirma la direccion, pero el chip ID `0x58`/`0x60` determina si es BMP280/BME280.
