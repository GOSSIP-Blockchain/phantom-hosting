# Masternode Hosting Solution

[![The MIT License](https://img.shields.io/badge/license-MIT-orange.svg?style=flat-square)](http://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/jackkdev/phantom-hosting?style=flat-square)](https://goreportcard.com/report/github.com/jackkdev/phantom-hosting)

This is a solution which allows for "fake" masternode hosting.

## Installation and Usage
```
git clone https://github.com/jackkdev/phantom-hosting.git
cd phantom-hosting-api
dep ensure
go run main.go
```

## Making requests
### Generate a masternode configuration string

```http request
POST http://localhost:8000/api/generatemasternodestring
```
#### Request Body
```json
{
  "port": 9998,
  "genkey": "75eqvNfaEfkd3YTwQ3hMwyxL2BgNSrqHDgWc6jbUh4Gdtnro2Wo",
  "txid": "f8a3e39da2d13e10736a77940a2a78823e30e3ac40140f0a0b1ec31d07989aef",
  "tx_index": 1
}
```
#### Request Response
```json
{
    "success": true,
    "data": "331720b1-6d69-404c-b84e-932642c93e92 [5a67:ae46:afa1:fd29:35a:2b37:dd1d:b138]:9998 75eqvNfaEfkd3YTwQ3hMwyxL2BgNSrqHDgWc6jbUh4Gdtnro2Wo f8a3e39da2d13e10736a77940a2a78823e30e3ac40140f0a0b1ec31d07989aef 1 1555938586",
    "error": ""
}
```

### Generate a masternode.txt file

```http request
POST http://localhost:8000/api/generateconfigfile
```
#### Request Response
```json
{
    "success": true,
    "data": "Configuration file created",
    "error": ""
}
```
The **masternode.txt** will be generated/stored in the project directory.

### Add a masternode to the configuration file

```http request
POST http://localhost:8000/api/addmasternode
```
#### Request Response
```json
{
    "success": true,
    "data": "Masternode added successfully to configuration file",
    "error": ""
}
```
The most recent masternode string created, which is stored in memory, will be appended to the **masternodes.txt** configuration file.

## Credits
* BreakCrypto - [Phantom Node Daemon](https://github.com/breakcrypto/phantom)