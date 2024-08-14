
import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_background_service/flutter_background_service.dart';

void ConnectProxy(ServiceInstance service) async {
  Socket? ProxySocket;
  Socket? LocalSocket;

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
      localAddr = data["local_addr"];
      proxyPort = int.tryParse(data["proxy_port"]) ?? 6775;
      proxyAddr = data["proxy_addr"] ?? "";
    }

    try {
      await ProxySocket?.close();
    } catch (e) {
      service.invoke("updateResponse", {
        "error": e.toString(),
      });
    }
    ProxySocket?.destroy();

    try {
      await LocalSocket?.close();
    } catch (e) {
      service.invoke("updateResponse", {
        "error": e.toString(),
      });
    }
    LocalSocket?.destroy();

    if (state) {
      try {
        ProxySocket = await Socket.connect(proxyAddr, proxyPort);
        LocalSocket = await Socket.connect(localAddr, localPort);
        await startProxyConnection(ProxySocket!, LocalSocket!, service, deviceName);
        service.invoke("updateResponse", {
          "state": true,
        });
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

  service.on("stopService").listen((event) {
    service.stopSelf();
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