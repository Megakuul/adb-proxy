# adb-proxy

adb-proxy is a simple application serving as a proof-of-concept for creating a proxy that allows devices (with corresponding client software) to register themselves. 
This enables the end user to access the devices adb ports via the proxy. 


The use case is for scenarios where you want to manage multiple devices adb connections, but the devices ip addresses are either `a)` not directly reachable or `b)` variable. 

Each device provides additional metadata (e.g., Name), which helps the user identify which device corresponds to which IP address. This metadata can be accessed through an HTTP web interface.



**Important**:
As mentioned, this is a proof-of-concept. While it works locally, it should not be used in production. There are no security measures or special authentication mechanisms in place to protect the system.
