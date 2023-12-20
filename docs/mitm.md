# MITM is not evil

Man-in-the-middle (MITM) is a proxy that acts as an intermediary between the client and the origin server engaged in communication. The MITM can read, decrypt, and alter exchanged traffic. As a result, the origin server can receive the request curated by the proxy instead of the original request sent by the client. The same applies to communication in
the reverse direction. The response delivered to the client could have been altered in transit by the proxy. For these reasons, MITM is often associated with a cybersecurity attack where an unwanted MITM proxy could be used to access sensitive information. TLS was designed to prevent it.

Assuming TLS is being used to secure communication, the only way for the proxy to introspect traffic is to perform a TLS handshake with the client. A successful handshake means the client approved the certificate presented by the proxy. In other words, it accepted the MITM.

How can a client prove a TLS handshake with the proxy?

An HTTPS client accepts certificates signed by certified authorities (CA). CA certificates are stored in a so-called trust store, a collection of certificates saved in some protected file system location where write
access is restricted to users with administrator privileges.

An important digression. Due to the way a trust store operates, users who log in as guests to an operating system managed by others (i.e., public library) should exercise increased caution when using a web browser. A system admin can install some additional certificates that will be accepted by a web browser as trusted.

Besides operating system vendors who decide which certificates are included in the pristine trust store, CA certificates are distributed as software libraries and come preinstalled with web browsers. In case a certificate cannot be verified, i.e., signed by an unknown authority, or expired, a user might still demand to proceed with a TLS handshake despite any warnings displayed by the browser.

Once the TLS handshake is established between the client and the proxy, the proxy establishes a TLS handshake with the origin server. Subsequently, the proxy manages both TCP connections. The proxy listens to the incoming traffic, parses HTTP requests and responses, and copies them between connections.

# TCP Tunneling

In contrast to the MITM which requires a client to accept a TLS handshake with the proxy, in TCP tunneling the TLS handshake is established between the client and the origin server.

Once a TLS handshake has been established between the client and the origin server, traffic passing through the communication channel is encrypted. Suppose a strong cipher is being used and no catastrophic bugs are present in the TLS stack, then traffic decryption is practically infeasible. Furthermore, attempts to alter communication by a third party can be immediately detected by either side of the connection.

In TCP tunneling, the proxy is accountable for ensuring the client and the origin server can perform a TLS handshake. It means the proxy has to accept the TCP connection with the client, establish a TCP connection with the origin server, and copy a stream of bytes between both connections.

The primary motivation for setting up a proxy for TCP tunneling is to protect the client's IP address/location from being discovered. As the origin server receives traffic for the IP address of the proxy, it is not
aware of the client's whereabouts. The origin server might not be even aware a proxy is being used. Furthermore, the proxy could perform some throttling activity restricted to the TCP layer, i.e., to control the number of connections opened simultaneously and to keep bandwidth under control. 


