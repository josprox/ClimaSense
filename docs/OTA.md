# OTA

El modelo de datos de releases y estados existe, pero el instalador OTA con Ed25519, checksum, health check y rollback aun no esta implementado. No publique releases como actualizables hasta completar esa ruta.

Actualmente las actualizaciones se distribuyen volviendo a ejecutar el instalador Raspberry Pi OS; este conserva `/data/climasense`, incluida identidad, token, Wi-Fi y cola. Ese mecanismo no es OTA desatendida ni sustituye firma y rollback.

La evolucion prevista descarga a un archivo temporal, verifica firma/arquitectura/version, conserva la version anterior, realiza un reemplazo atomico y revierte si el servicio no alcanza salud. Una fase posterior debe usar particiones A/B para el sistema completo.
