# I2C

`dtparam=i2c_arm=on` activa el controlador y el kernel incluye `I2C_CHARDEV`. El plugin configura `I2C_SLAVE`, `I2C_TIMEOUT` e `I2C_RETRIES`, serializa cada transaccion y rechaza direcciones fuera de 7 bits.

API Joss:

```joss
$devices = ClimaSenseHardware::i2cScan(1)
$id = ClimaSenseHardware::i2cReadRegister(1, 0x76, 0xD0, 1, 1000)
$reading = ClimaSenseHardware::bmp280Read(1, "auto", {"pressure_oversampling": 4, "filter": 2})
```

El usuario `climasense` pertenece al grupo `i2c`; la aplicacion completa no corre como root.
