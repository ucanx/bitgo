# bitgo

Bitcoin Handshake implementation using Go

## Explanation
* Serialization of the version Message: We serialize each field of the version message according to the Bitcoin protocol specification. This includes converting IP addresses to a 16-byte format and ensuring all integers are in little-endian format.
* Handling Incoming Messages: We read the header of each incoming message to determine its type and payload size. Depending on the command (either version or verack), we handle it accordingly. After receiving a version, we send a verack, and upon receiving a verack, we conclude the handshake.
* Sending verack Message: After receiving a version message from the node, we send a verack message to acknowledge it, which completes the handshake from our side.
