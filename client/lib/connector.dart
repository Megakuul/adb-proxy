import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_background_service/flutter_background_service.dart';
import 'package:flutter_background_service_android/flutter_background_service_android.dart';

@pragma('vm:entry-point')
void connectProxy(ServiceInstance? service) async {
  if (service==null) {
    return;
  }

  service.on("stopService").listen((event) {
    service.stopSelf();
  });

  if (service is AndroidServiceInstance) {
    service.on('setAsForeground').listen((event) {
      service.setAsForegroundService();
    });

    service.on('setAsBackground').listen((event) {
      service.setAsBackgroundService();
    });
  }

  Socket? proxySocket;
  Socket? localSocket;

  bool state = false;
  String deviceName = "";
  String proxyAddr = "";
  int proxyPort = 0;
  int localPort = 0;
  String localAddr = "";

  service.on("updateRequest").listen((data) async {
    if (data != null) {
      state = data["state"] ?? false;
      deviceName = data["device_name"] ?? "undefined";
      localPort = int.tryParse(data["local_port"]) ?? 5555;
      localAddr = data["local_addr"] ?? "";
      proxyPort = int.tryParse(data["proxy_port"]) ?? 6775;
      proxyAddr = data["proxy_addr"] ?? "";
    }

    try {
      await proxySocket?.close();
    } catch (e) {
      service.invoke("updateResponse", {
        "error": e.toString(),
      });
    }
    proxySocket?.destroy();

    try {
      await localSocket?.close();
    } catch (e) {
      service.invoke("updateResponse", {
        "error": e.toString(),
      });
    }
    localSocket?.destroy();

    if (state) {
      try {
        proxySocket = await Socket.connect(proxyAddr, proxyPort);
        localSocket = await Socket.connect(localAddr, localPort);
        service.invoke("updateResponse", {
          "state": true,
        });
        await startProxyConnection(proxySocket!, localSocket!, service, deviceName);
      } catch (e) {
        service.invoke("updateResponse", {
          "state": false,
          "error": e.toString(),
        });
      }
    } else {
      service.invoke("updateResponse", {
        "state": false,
      });
    }
  });
}

Future<void> startProxyConnection(Socket proxySocket, Socket localSocket, ServiceInstance service, String name) async {
  final header = utf8.encode(jsonEncode({
    "name": name,
  }));
  final headerLength = ByteData(2)..setUint16(0, header.length, Endian.big);
  final payload = headerLength.buffer.asUint8List() + header;

  proxySocket.add(payload);

  localSocket.listen((data) {
    proxySocket.add(data);
  }, onError: (error) {
    service.invoke("updateResponse", {
      "state": false,
      "error": error.toString(),
    });
  }, onDone: () {
    service.invoke("updateResponse", {
      "state": false,
    });
  });
  proxySocket.listen((data) {
    localSocket.add(data);
  }, onError: (error) {
    service.invoke("updateResponse", {
      "state": false,
      "error": error.toString(),
    });
  }, onDone: () {
    service.invoke("updateResponse", {
      "state": false,
    });
  });
}