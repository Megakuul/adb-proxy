import 'dart:io';

import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'dart:async';

import 'package:flutter_background_service/flutter_background_service.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';

import 'connector.dart';



void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final service = FlutterBackgroundService();

  // this function is called because any weird android optimizer or obfuscater
  // (probably r8) deletes the function if it's just set as a callback on the service.
  connectProxy(null);

  // notification channel is created and registered because the newer android
  // apis require a notification for foreground services.
  const AndroidNotificationChannel channel = AndroidNotificationChannel(
    'adb_proxy_connector_channel',
    'adb_proxy',
    description: 'adb_proxy notification channel.',
    importance: Importance.high,
  );

  final FlutterLocalNotificationsPlugin flutterLocalNotificationsPlugin = FlutterLocalNotificationsPlugin();

  if (Platform.isAndroid) {
    await flutterLocalNotificationsPlugin.initialize(
      const InitializationSettings(
        android: AndroidInitializationSettings('ic_bg_service_small'),
      ),
    );
  }

  await flutterLocalNotificationsPlugin
      .resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>()
      ?.createNotificationChannel(channel);

  await service.configure(
    androidConfiguration: AndroidConfiguration(
      onStart: connectProxy,
      isForegroundMode: true,
      autoStart: true,
      notificationChannelId: 'adb_proxy_connector_channel',
      initialNotificationTitle: 'ADB Proxy Launched',
      initialNotificationContent: 'ADB Proxy Connector is running in the background.',
      foregroundServiceNotificationId: 420,
      foregroundServiceType: AndroidForegroundType.dataSync,
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

  List<DropdownMenuItem<String>> deviceAddressList = [];
  String? deviceAddress;

  Future<void> _initDeviceAddressList() async {
    deviceAddressList = [];
    try {
      for (var interface in await NetworkInterface.list()) {
        for (var addr in interface.addresses) {
          deviceAddressList.add(DropdownMenuItem<String>(
            value: addr.address,
            alignment: Alignment.center,
            child: Text(addr.address, style: GoogleFonts.ubuntu(color: Colors.white60))
          ));
        }
      }
    } catch (err) {
      errorMessage = err.toString();
    }
  }

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
    _initDeviceAddressList();
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
          body: SingleChildScrollView(
            child: Container(
              padding: const EdgeInsets.all(12),
              height: MediaQuery.of(context).size.height,
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
                      style: GoogleFonts.ubuntu(color: Colors.white60),
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
                    child: Align(
                      alignment: Alignment.center,
                      child: Container(
                        decoration: BoxDecoration(
                          border: Border.all(color: Colors.white60.withOpacity(0.4)),
                          borderRadius: BorderRadius.circular(12)
                        ),
                        child: DropdownButton<String>(
                          hint: Text("Device Address", style: GoogleFonts.ubuntu(color: Colors.white60)),
                          borderRadius: BorderRadius.circular(12),
                          icon: const Icon(Icons.arrow_drop_down_circle_outlined, color: Colors.white60),
                          alignment: Alignment.center,
                          padding: const EdgeInsets.all(10),
                          dropdownColor: Colors.black87.withOpacity(0.6),
                          isExpanded: true,
                          underline: const SizedBox(),
                          value: deviceAddress,
                          onChanged: (String? newAddr) {
                            setState(() {
                              deviceAddress = newAddr;
                            });
                          },
                          items: deviceAddressList,
                        ),
                      )
                    ),
                  ),
                  const SizedBox(height: 5),
                  Expanded(
                    child: TextField(
                      keyboardType: TextInputType.number,
                      style: GoogleFonts.ubuntu(color: Colors.white60),
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
                  Row(
                    children: [
                      Expanded(
                        flex: 2,
                        child: TextField(
                          style: GoogleFonts.ubuntu(color: Colors.white60),
                          decoration: const InputDecoration(
                            hintText: "Proxy Address",
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.all(Radius.circular(12))
                            )
                          ),
                          controller: proxyAddrController,
                        ),
                      ),
                      const SizedBox(width: 5),
                      Expanded(
                        flex: 1,
                        child: TextField(
                          keyboardType: TextInputType.number,
                          style: GoogleFonts.ubuntu(color: Colors.white60),
                          decoration: const InputDecoration(
                            hintText: "Proxy Port",
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.all(Radius.circular(12))
                            )
                          ),
                          controller: proxyPortController,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),
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
                          "local_addr": deviceAddress,
                          "local_port": devicePortController.text,
                          "proxy_addr": proxyAddrController.text,
                          "proxy_port": proxyPortController.text,
                        });
                      },
                      child: Text("Apply", style: GoogleFonts.ubuntu(color: Colors.white70, fontSize: 30))
                    ),
                  ),
                  const SizedBox(height: 20),
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
                      height: 100,
                      width: double.infinity,
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
            ),
          )
        ),
      )
    );
  }
}
