# adb-proxy client

The adb-proxy client uses a background service to connect to the adb-proxy server.

Establishing a connection to the proxy, allows users with proxy access to connect to the 
specified device address / port on your device without additional authorization or any port being exposed on your phone.


The application release can be built with:
```bash
flutter build apk --release
```

Resulting in `build/app/outputs/flutter-apk/app-release.apk` which can be installed on the device with:
```bash
adb -s <deviceid> install build/app/outputs/flutter-apk/app-release.apk
```

In case of an issue, the logs can be examined with logcat:
```bash
adb -s <deviceid> logcat | grep flutter
```