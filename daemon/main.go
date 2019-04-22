/**
*    Copyright (C) 2019-present C2CV Holdings, LLC.
*
*    This program is free software: you can redistribute it and/or modify
*    it under the terms of the Server Side Public License, version 1,
*    as published by C2CV Holdings, LLC.
*
*    This program is distributed in the hope that it will be useful,
*    but WITHOUT ANY WARRANTY; without even the implied warranty of
*    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
*    Server Side Public License for more details.
*
*    You should have received a copy of the Server Side Public License
*    along with this program. If not, see
*    <http://www.mongodb.com/licensing/server-side-public-license>.
*
*    As a special exception, the copyright holders give permission to link the
*    code of portions of this program with the OpenSSL library under certain
*    conditions as described in each individual source file and distribute
*    linked combinations including the program with the OpenSSL library. You
*    must comply with the Server Side Public License in all respects for
*    all of the code used other than as permitted herein. If you modify file(s)
*    with this exception, you may extend this exception to your version of the
*    file(s), but you are not obligated to do so. If you do not wish to do so,
*    delete this exception statement from your version. If you delete this
*    exception statement from all source files in the program, then also delete
*    it in the license file.
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"log"
	"net"
	"phantom/socket/wire"
	"strconv"
	"strings"
	"sync"
	"time"
)

var maxConnections uint

var magicBytes uint32
var defaultPort uint
var protocolNumber uint32
var magicMessage string
var bootstrapIPs string
var bootstrapHash chainhash.Hash
var bootstrapExplorer string
var sentinelVersion uint32
var daemonVersion uint32

const VERSION = "0.0.1"

func main() {

	//disable all logging
	//log.SetOutput(ioutil.Discard)

	var magicHex string
	var magicMsgNewLine bool
	var protocolNum uint
	var bootstrapHashStr string

	flag.UintVar(&maxConnections, "max_connections", 10, "the number of peers to maintain")
	flag.StringVar(&magicHex, "magicbytes", "", "a hex string for the magic bytes")
	flag.UintVar(&defaultPort, "port", 0, "the default port number")
	flag.UintVar(&protocolNum, "protocol_number", 0, "the protocol number to connect and ping with")
	flag.StringVar(&magicMessage, "magic_message", "", "the signing message")
	flag.BoolVar(&magicMsgNewLine, "magic_message_newline", true, "add a new line to the magic message")
	flag.StringVar(&bootstrapIPs, "bootstrap_ips", "", "IP address to bootstrap the network")
	flag.StringVar(&bootstrapHashStr, "bootstrap_hash", "", "Hash to bootstrap the pings with ( top - 12 )")
	flag.StringVar(&bootstrapExplorer, "bootstrap_url", "", "Explorer to bootstrap from.")

	var sentinelString string
	var daemonString string

	flag.StringVar(&sentinelString, "sentinel_version", "0.0.0", "The string to use for the sentinel version number (i.e. 1.20.0)")
	flag.StringVar(&daemonString, "daemon_version", "0.0.0.0", "The string to use for the sentinel version number (i.e. 1.20.0)")

	flag.Parse()

	magicMsgNewLine = true

	magicBytes64, _ := strconv.ParseUint(magicHex, 16, 32)
	magicBytes = uint32(magicBytes64)

	protocolNumber = uint32(protocolNum)

	if sentinelString != "" {
		//fmt.Println("ENABLING SENTINEL.")
		sentinelVersion = convertVersionStringToInt(sentinelString)
	}

	if daemonString != "" {
		//fmt.Println("ENABLING SENTINEL.")
		sentinelVersion = convertVersionStringToInt(daemonString)
	}

	if magicMsgNewLine {
		magicMessage = magicMessage + "\n"
	}

	var connectionSet = make(map[string]*PingerConnection)
	var peerSet = make(map[string]wire.NetAddress)

	var waitGroup sync.WaitGroup

	addresses := splitAddressList(bootstrapIPs)

	if uint(len(addresses)) > maxConnections {
		log.Fatal("The number of bootstrap IPs is larger than the number of max connections, please try again.")
	}

	for _, address := range addresses {
		peerSet[address.IP.String()] = address
	}

	addrProcessingChannel := make(chan wire.NetAddress, 1500)
	hashProcessingChannel := make(chan chainhash.Hash, 1500)

	hashQueue := NewQueue(12)

	if bootstrapExplorer != "" {
		bootstrapper := Bootstrapper{bootstrapExplorer}
		var err error
		bootstrapHash, err = bootstrapper.LoadBlockHash()

		empthHash := chainhash.Hash{}
		if err != nil {
			log.Fatal("Unable to bootstrap using the explorer url provided. ", err)
		}

		if bootstrapHash == empthHash {
			log.Fatal("Unable to bootstrap using the explorer url provided. Invalid result returned.")
		}
	} else {
		chainhash.Decode(&bootstrapHash,bootstrapHashStr)
		hashQueue.Push(&bootstrapHash)
	}

	Preamble()

	time.Sleep(10 * time.Second)

	fmt.Println("--USING THE FOLLOWING SETTINGS--")
	fmt.Println("Magic Bytes: ", magicHex)
	fmt.Println("Magic Message: ", magicMessage)
	fmt.Println("Magic Message Newline: ", magicMsgNewLine)
	fmt.Println("Protocol Number: ", protocolNumber)
	fmt.Println("Bootstrap IPs: ", bootstrapIPs)
	fmt.Println("Default Port: ", defaultPort)
	fmt.Println("Hash: ", bootstrapHash)
	fmt.Println("Sentinel Version: ", sentinelVersion)
	fmt.Println("Sentinel Version: ", daemonVersion)
	fmt.Println("\n\n")

	for ip := range peerSet {
		//make the ping channel
		pingChannel := make(chan MasternodePing, 1500)

		waitGroup.Add(1)

		pinger := PingerConnection{
			MagicBytes: magicBytes,
			IpAddress: ip,
			Port: uint16(defaultPort),
			ProtocolNumber: protocolNumber,
			BootstrapHash: bootstrapHash,
			PingChannel: pingChannel,
			AddrChannel: addrProcessingChannel,
			HashChannel: hashProcessingChannel,
			Status: 0,
			WaitGroup: &waitGroup,
		}

		//make a client
		connectionSet[pinger.IpAddress] = &pinger

		go pinger.Start()
	}

	pingGeneratorChannel := make(chan MasternodePing, 1500)

	waitGroup.Add(1)

	go sendPings(connectionSet, peerSet, pingGeneratorChannel, addrProcessingChannel, hashProcessingChannel, waitGroup)
	go generatePings(pingGeneratorChannel, hashQueue, magicMessage)
	go processNewAddresses(addrProcessingChannel, peerSet)
	go processNewHashes(hashProcessingChannel, hashQueue)

	waitGroup.Wait()
}

func splitAddressList(bootstraps string) (addresses []wire.NetAddress) {
	for _, bootstrap := range strings.Split(bootstraps, ",") {
		ipPort := strings.Split(bootstrap, ":")
		ip := ipPort[0]
		port, _ := strconv.Atoi(ipPort[1])
		addresses = append(addresses, wire.NetAddress{time.Now(),
			0,
			net.ParseIP(ip),
			uint16(port)})
	}
	return addresses
}

func generatePings(pingChannel chan MasternodePing, queue *Queue, magicMessage string) {
	for {

		fmt.Println("Loading settings from masternode.txt")
		GeneratePingsFromMasternodeFile("./masternode.txt", pingChannel, queue, magicMessage, sentinelVersion)
		time.Sleep(time.Minute * 10)
	}
}

func processNewHashes(hashChannel chan chainhash.Hash, queue *Queue) {
	for {
		hash := <-hashChannel

		//log.Println("Adding hash to queue: ", hash.String(), "(", queue.count, ")")

		queue.Push(&hash)
		for queue.count > 12 { //clear the queue until we're at 12 entries
			queue.Pop()
			//log.Println("Removing hash from queue: ", popped.String(), "(", queue.count, ")")
		}
	}
}

func processNewAddresses(addrChannel chan wire.NetAddress, peerSet map[string]wire.NetAddress) {
	for {
		addr := <-addrChannel

		if addr.IP.To4() == nil {
			continue
		}

		peerSet[addr.IP.String()] = addr
	}
}

func getNextPeer(connectionSet map[string]*PingerConnection, peerSet map[string]wire.NetAddress) (returnValue wire.NetAddress, err error) {
	for peer := range peerSet {
		if _, ok := connectionSet[peer]; !ok {
			//we have a peer that isn't in the conncetion list return it
			returnValue = peerSet[peer]

			//remove the peer from the connection list
			delete(peerSet, peer)

			log.Println("Found new peer: ", peer)

			return returnValue, nil
		}
	}
	return returnValue, errors.New("No peers found.")
}

func sendPings(connectionSet map[string]*PingerConnection, peerSet map[string]wire.NetAddress, pingChannel chan MasternodePing, addrChannel chan  wire.NetAddress, hashChannel chan chainhash.Hash, waitGroup sync.WaitGroup) {

	time.Sleep(10 * time.Second) //hack to work around .Wait() race condition on fast start-ups

	for {
		ping := <-pingChannel

		sleepTime := ping.PingTime.Sub(time.Now())

		log.Println(time.Now().UTC())
		log.Println(ping.Name, ping.PingTime.UTC())

		if sleepTime > 0 {
			fmt.Println("Sleeping for ", sleepTime.String())
			//log.Println("SLEEPING FOR: " + sleepTime.String())
			time.Sleep(sleepTime)
		}

		//send the ping
		// Iterate through list and print its contents.
		var newConnectionSet = make(map[string]*PingerConnection)

		for _, pinger := range connectionSet {
			status := pinger.GetStatus()

			if status < 0 || len(pinger.PingChannel) > 10 { //the pinger has had an error, close the channel
				fmt.Println("There's been an error, closing connection to ", pinger.IpAddress)
				pinger.SetStatus(-1)

				//log.Printf("%s : Closing down the ping channel.\n", pinger.IpAddress )
				close(pinger.PingChannel) // don't add the closed pinger to the connectionArray

				//remove the peer from the peerSet
				delete(peerSet, pinger.IpAddress)
			} else {
				if status > 0 {
					//log.Printf("%s : Pinging.", pinger.IpAddress)
					pinger.PingChannel <- ping //only ping on connected pingers (1)
				}
				// this filters out bad connections, re-add unconnected peers just to be safe
				log.Printf("Re-added %s to the queue (channel #: %d).\n", pinger.IpAddress, len(pinger.PingChannel))
				newConnectionSet[pinger.IpAddress] = pinger
			}
		}

		//replace the pointer
		connectionSet = newConnectionSet

		fmt.Println("Current number of connections to network: (", len(connectionSet), " / ", maxConnections, ")")

		//spawn off extra nodes here if we don't have enough
		if len(connectionSet) <  int(maxConnections) {

			log.Println("Under the max connection count, spawning new peer (", len(connectionSet), " / ", maxConnections, ")")

			for i := 0; i < int(maxConnections) - len(connectionSet); i++ {

				//spawn off a new connection
				peer, err := getNextPeer(connectionSet, peerSet)

				if err != nil {
					log.Println("No new peers found.")
					continue
				}

				newPingChannel := make(chan MasternodePing, 1500)

				// intentionally don't provide a bootstraphash to prevent
				// duplicate data downloads for unneeded blocks
				newPinger := PingerConnection{
					MagicBytes: 	magicBytes,
					IpAddress:      peer.IP.String(),
					Port:           peer.Port,
					ProtocolNumber: protocolNumber,
					PingChannel:    newPingChannel,
					AddrChannel: 	addrChannel,
					HashChannel: 	hashChannel,
					Status:         0,
					WaitGroup:      &waitGroup,
				}

				//make a client
				newConnectionSet[newPinger.IpAddress] = &newPinger
				//connectionList = nil //release for the GC

				waitGroup.Add(1)
				go newPinger.Start()

				fmt.Println("Opened a new connection to ", newPinger.IpAddress)
			}
		}
		log.Println(time.Now().UTC())
		log.Println(ping.Name, ping.PingTime.UTC())
	}

	waitGroup.Done()
}

func convertVersionStringToInt(str string) uint32 {
	version := 0
	parts := strings.Split(str, ".")
	for _, part := range parts {
		version <<= 8
		value, _ := strconv.Atoi(part)
		version |= value
	}
	return uint32(version)
}