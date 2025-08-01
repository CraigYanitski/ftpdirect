# FTPdirect
---

## Motivation

While collaborating with developers even within the same country, I found that transferring large files was 
rather cumbersome.
Mail and conferencing clients have limit data transfer, and if you are fortunate enough to have access 
to a secure server with port forwarding, maybe it is possible to share a link to a collaborator to download the file.
More often than not, though, these servers are accessible to a private network and you can only 
access them if you can tunnel into it.

This is the reason I developed FTPdirect: to provide a peer-to-peer file transfer with symmetric encryption.
It is intended to be used while in contact with a peer in order to share the P2P address and assure 
a symmetric Diffie-Hellman key exchange.
Currently a direct P2P connection is not established and files are transferred via the server.

## Installation

The best way to test this is to download and run an executable.
This is setup to use the current deployed server.
Any downloaded files will default to a `.ftpd` directory in your home directory.

If you would like to use FTPdirect on your own server, make sure you create a `.env` file setting `FTPD_PORT`
for the port where you want your server to listen for connections.
The client executables are also setup to use the `.env` file to get `FTPD_SCHEME`, `FTPD_URL`, and `FTPD_ENDPOINT`.
Likely you should hardcode these into `cmd/client/main.go` for your build.
Then to build the client `ftpd` and server `ftpd-server` executables run,

```bash
make build
```

## Usage

Running the client will start listening for a TCP connection on your local network, make a websocket 
connection to the server, and begin a REPL.
The prompt will initially be `Headless -> ` since you are not connected to a peer.
Enter the command `connect` to initiate a connection to which your peer can connect.
The server will assign a UUID to your connection, and you should send to a peer.
Your prompt has now updated to your connection UUID.
Your peer should then enter `connect <UUID>` in their REPL to complete the connection.
Now you or your peer can can use the command `send <filename>` to send a file over the connection.

If your peer is in your local network, run the executable with,

```bash
ftpd -i true
```

This will specify that the connection is internal.
The main difference here is that you connect to your peer using `connect <TCPaddr>`, where 
you can enter the TCP IP address of your peer.
This can be used to send a file to your peer, and your peer can do the same by connecting to your TCP IP address.
The issue with this at the moment is that the filename will just be the timestamp, and you will need to 
rename it manually.

## Development

There are a few features that will be developed in the coming months.

**Establish TCP connection rather than using websocket**: sending file data through the server throttles the 
transfer speed and is not technically a *direct* method for FTP.
The primary issue that will need to be overcome is establishing a TCP connection between peers
behind independent Network Address Translation (NAT) implementations.
Therefore some form of NAT traversal is required, which might not work on symmetric NAT implementations and 
the connection should default to the websocket connection transferring files.

**Encrypt file transfer**: this will not be difficult to implement and will come soon.
The client will generate a Diffie-Hellman key to share via the websocket connection to a peer.
Of course this is not more robust against a man-in-the-middle attack over TLS, but it helps.

**General refactoring**: there is too much hardcoding in this project.
This will be removed in future versions of FTPdirect.

## Contributing

Feel free to clone the repo and make a pull request to suggest updates

