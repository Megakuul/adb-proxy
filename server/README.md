# adb-proxy server

You can start the proxy server using three arguments:

```bash
./adb-proxy-server 7000 6775 9000-10000 1h
```


The first argument specifies the port for the HTTP interface. You can access it with:

```bash
curl http://127.0.0.1:7000
```


The second argument sets the port for the proxy, which is used by clients (devices) to connect to the proxy.

The proxy supports only one connection at a time. To connect from another device, disconnect the current connection first.



The third argument defines the port range allocated for proxy connections.



The fourth argument defines the duration after which a device is removed from the proxy if no client connected to it.



### Device Discovery


Devices can register themselves on the adb-proxy server by initiating a TCP connection to the proxy. After the connection is established, the server expects the following format:

| **Size of Segment** | **Purpose**   | **Content**                              |
|---------------------|---------------|------------------------------------------|
| `2 byte uint16`     | Header length | Specifies the length of the JSON header. |
| header length       | JSON header   | `{ "name": "devicename" }`               |
| Variable            | Body          | TCP stream redirected from adb.          |
