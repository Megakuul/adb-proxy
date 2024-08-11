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

  TextEditingController proxyAddrController = TextEditingController();
  TextEditingController proxyPortController = TextEditingController();

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
                  hintText: "Proxy Address"
                ),
                controller: proxyAddrController,
              ),
            ),
            Expanded(
              child: TextField(
                decoration: const InputDecoration(
                    hintText: "Proxy Port"
                ),
                controller: proxyPortController,
              ),
            ),
            OutlinedButton(
              onPressed: () {

              },
              child: const Text("Update Proxy Values")
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                OutlinedButton(
                  onPressed: () => widget.service.startService(),
                  child: const Text("Start"),
                ),
                OutlinedButton(
                  onPressed: () => widget.service.invoke("stopService"),
                  child: const Text("Stop"),
                )
              ],
            )
          ],
        )
      ),
    );
  }
}
