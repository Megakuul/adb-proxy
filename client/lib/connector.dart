
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

  service.on("updateRequest").listen((data) async {
    if (data != null) {
      state = data["state"] ?? false;
      deviceName = data["device_name"] ?? "undefined";
      localPort = data["local_port"] ?? 0;
      proxyPort = data["proxy_port"] ?? 0;
      proxyAddr = data["proxy_addr"] ?? "";
    }
    await ProxySocket?.close();
    ProxySocket?.destroy();
    await LocalSocket?.close();
    LocalSocket?.destroy();
    if (state) {
      try {
        ProxySocket = await Socket.connect(proxyAddr, proxyPort);
        LocalSocket = await Socket.connect("127.0.0.1", localPort);
        await startProxyConnection(ProxySocket!, LocalSocket!, service, deviceName);
      } catch (e) {
        service.invoke("updateResponse", {
          "state": true,
          "error": e,
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
  localSocket.listen((data) {
    proxySocket.write(data);
  }, onError: (error) {
    service.invoke("updateResponse", {
      "error": error,
    });
  });
  proxySocket.listen((data) {
    localSocket.write(data);
  }, onError: (error) {
    service.invoke("updateResponse", {
      "error": error,
    });
  });

  final header = utf8.encode(jsonEncode({
    "name": name,
  }));

  final headerLength = ByteData(2)..setUint16(0, header.length, Endian.big);

  final payload = headerLength.buffer.asUint8List() + header;

  proxySocket.write(payload);
}