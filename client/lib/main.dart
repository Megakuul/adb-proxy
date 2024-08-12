import 'package:flutter/material.dart';
import 'package:flutter_background_service/flutter_background_service.dart';

import 'connector.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final service = FlutterBackgroundService();

  await service.configure(
    androidConfiguration: AndroidConfiguration(
      onStart: ConnectProxy,
      isForegroundMode: false,
      autoStart: true
    ),
    // Empty IOS configuration.
    iosConfiguration: IosConfiguration(
      onForeground: (service) => true,
      onBackground: (service) async => true,
    )
  );

  service.startService();

  runApp(ProxyClient(service));
}

class ProxyClient extends StatefulWidget {
  const ProxyClient(this.service, {super.key});

  final FlutterBackgroundService service;

  @override
  State<ProxyClient> createState() => _ProxyClientState();
}

class _ProxyClientState extends State<ProxyClient> {

  TextEditingController deviceNameController = TextEditingController();
  TextEditingController devicePortController = TextEditingController();
  TextEditingController proxyAddrController = TextEditingController();
  TextEditingController proxyPortController = TextEditingController();

  bool connectionState = false;
  String errorMessage = "";

  @override
  void initState() {
    super.initState();
    widget.service.on("updateResponse").listen((data) {
      if (data != null) {
        setState(() {
          connectionState = data["state"] ?? false;
          errorMessage = data["error"] ?? "";
        });
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: "adb-proxy client",
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      home: Scaffold(
        body: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Text("adb-proxy"),
            Expanded(
              child: TextField(
                decoration: const InputDecoration(
                    hintText: "Device Name",
                    border: OutlineInputBorder(
                        borderRadius: BorderRadius.all(Radius.circular(12))
                    )
                ),
                controller: deviceNameController,
              ),
            ),
            Expanded(
              child: TextField(
                decoration: const InputDecoration(
                    hintText: "Device Port",
                    border: OutlineInputBorder(
                        borderRadius: BorderRadius.all(Radius.circular(12))
                    )
                ),
                controller: devicePortController,
              ),
            ),
            Expanded(
              child: TextField(
                decoration: const InputDecoration(
                    hintText: "Proxy Address",
                    border: OutlineInputBorder(
                        borderRadius: BorderRadius.all(Radius.circular(12))
                    )
                ),
                controller: proxyAddrController,
              ),
            ),
            Expanded(
              child: TextField(
                decoration: const InputDecoration(
                    hintText: "Proxy Port",
                    border: OutlineInputBorder(
                        borderRadius: BorderRadius.all(Radius.circular(12))
                    )
                ),
                controller: proxyPortController,
              ),
            ),
            Text(connectionState ? "ON" : "OFF"),
            Text(errorMessage),
            OutlinedButton(
              onPressed: () {
                print("I pressed from isolate 1");
                widget.service.invoke("updateRequest", {
                  "state":  true,
                  "device_name": deviceNameController.text,
                  "local_port": devicePortController.text,
                  "proxy_addr": proxyAddrController.text,
                  "proxy_port": proxyPortController.text,
                });
              },
              child: const Text("Update Proxy Values")
            ),
          ],
        )
      ),
    );
  }
}
