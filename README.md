# Phantom Hosting

[![The SSPL License](https://img.shields.io/badge/license-SSPL-orange.svg?style=flat-square)](https://www.mongodb.com/licensing/server-side-public-license)
[![Travis](https://img.shields.io/travis/carsenk/phantom-hosting.svg?style=flat-square)](https://travis-ci.org/carsenk/phantom-hosting)
[![Go Report Card](https://goreportcard.com/badge/github.com/carsenk/phantom-hosting?style=flat-square)](https://goreportcard.com/report/github.com/carsenk/phantom-hosting)

The all in one self hosted Phantom Masternode solution.

![Landing Page](./screenshots/landing.png)

![Deploy Page](./screenshots/deploy.png)

## Installation and Usage
```
git clone https://github.com/carsenk/phantom-hosting.git
cd phantom-hosting
dep ensure
go run main.go
```

## Standalone Build
```
./build.sh
```

## Docker
Run the Phantom Hosting solution under a Docker container
```
docker build -t phantom-hosting .
docker run -d -p 8000:8000 phantom-hosting

```

#### Docker Compose
```
docker-compose up -d
```

## User Interface
The user interface for the Phantom hosting solution is being served on:
```
http://localhost:8000
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
* BreakCrypto/Carsenk/Jackkdev - [Phantom Node Daemon](https://github.com/carsenk/phantom)