# secureFileTransfer
Simple file transfer client and server with few basic security features

## Security features:
- tls connection
- constant time password comparison

## Server

Run in folder where you want to save files.  
Optionally password for connections can be set with ```--pass``` argument.

## Client

Server address with port should be specified with ```--server``` argument.   
Password for server can be set with ```--pass``` argument.  
Files to transfer should be listed in the end.
