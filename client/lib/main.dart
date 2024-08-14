import 'package:flutter/material.dart';
import 'package:flutter_background_service/flutter_background_service.dart';
import 'package:google_fonts/google_fonts.dart';

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
      home: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            colors: [Colors.black, Colors.purple],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          )
        ),
        child: Scaffold(
          backgroundColor: Colors.transparent,
          body: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.center,
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const SizedBox(height: 25),
                Expanded(
                  flex: 2,
                  child: Text(
                    "adb-proxy",
                    style: GoogleFonts.ubuntu(fontSize: 60, color: Colors.white70)
                  ),
                ),
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
                const SizedBox(height: 5),
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
                const SizedBox(height: 5),
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
                const SizedBox(height: 5),
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
                const SizedBox(height: 5),
                SizedBox(
                  width: double.infinity,
                  height: 50,
                  child: OutlinedButton(
                    style: OutlinedButton.styleFrom(
                      backgroundColor: Colors.white24.withOpacity(0.05),
                      shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12)
                      ),
                    ),
                    onPressed: () {
                      widget.service.invoke("updateRequest", {
                        "state":  true,
                        "device_name": deviceNameController.text,
                        "local_port": devicePortController.text,
                        "proxy_addr": proxyAddrController.text,
                        "proxy_port": proxyPortController.text,
                      });
                    },
                    child: Text("Update Proxy Values", style: GoogleFonts.ubuntu(color: Colors.white70, fontSize: 30))
                  ),
                ),
                const SizedBox(height: 25),
                Container(
                  height: 50,
                  width: double.infinity,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    gradient: LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: connectionState ? [
                        Colors.green.withOpacity(0.4),
                        Colors.lightGreen.withOpacity(0.7)
                      ] : [
                        Colors.red.withOpacity(0.4),
                        Colors.deepOrange.withOpacity(0.7)
                      ]
                    ),
                  ),
                  child: Text(connectionState ? "ON" : "OFF", style: GoogleFonts.ubuntu(color: Colors.white70, fontSize: 40)),
                ),
                const SizedBox(height: 15),
                SingleChildScrollView(
                  child: Container(
                    padding: const EdgeInsets.all(5),
                    alignment: AlignmentDirectional.bottomStart,
                    width: double.infinity,
                    height: 100,
                    decoration: BoxDecoration(
                      borderRadius: BorderRadius.circular(12),
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          Colors.white54.withOpacity(0.3),
                          Colors.white60.withOpacity(0.2)
                        ]
                      )
                    ),
                    child: Text(
                      errorMessage,
                      style: GoogleFonts.ubuntu(color: Colors.red, fontSize: 15)
                    ),
                  ),
                ),
              ],
            ),
          )
        ),
      )
    );
  }
}
