/*
	Project : DEPTH-FIRST SEARCH WITH SPECIFIED ROOT FOR CREATION OF A SPANNING TREE
	Author : ANTOINE LESTRADE, GUILLAUME DIGIER
	Date : DECEMBER 2021
*/

package main

//---------------------
// IMPORTS
//---------------------

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

//---------------------
// PREPERATIONS
//---------------------

var PORT string = ":30000"

var MST = make(map[string][]string)

type yamlConfig struct {
	ID      int    `yaml:"id"`
	Address string `yaml:"address"`

	Neighbours []struct {
		ID         int    `yaml:"id"`
		Address    string `yaml:"address"`
		EdgeWeight int    `yaml:"edge_weight"`
	} `yaml:"neighbours"`
}

//-------------------------
// PREPARATING FUNCTIONS
//-------------------------

func initAndParseFileNeighbours(filename string) yamlConfig {

	fullpath, _ := filepath.Abs("./Neighbours/" + filename)
	yamlFile, err := ioutil.ReadFile(fullpath)

	if err != nil {
		panic(err)
	}

	var data yamlConfig

	err = yaml.Unmarshal([]byte(yamlFile), &data)

	if err != nil {
		panic(err)
	}

	return data
}

//-------------------------
// COMMUNICATING FUNCTIONS
//-------------------------

func send(message string, nodeAddress string, neighAddress string) {

	myLog(nodeAddress, "Sending message to "+neighAddress)

	// Il faut bien preciser l'adresse de l'envoyeur si on veut que le recepteur puisse la connaitre.
	d := net.Dialer{LocalAddr: &net.TCPAddr{IP: net.ParseIP(nodeAddress)}}

	outConn, err := d.Dial("tcp", neighAddress+PORT)

	if err != nil {
		log.Fatal(err)

		return
	}

	outConn.Write([]byte(message))
	outConn.Close()
}

func sendToOneNeighbour(message string, node yamlConfig, neighbour string) {

	myLog(node.Address, "Sending message to one neighbour ...")

	go send(message, node.Address, neighbour)
}

//-------------------------
// UTILITY FUNCTIONS
//-------------------------

func myLog(localAdress string, message string) {

	fmt.Printf("[%s] : %s\n", localAdress, message)
}

func findMapIndexByValue(myMap map[int]string, myValue string) int {

	for k, v := range myMap {
		if v == myValue {
			return k
		}
	}

	fmt.Println("ERROR NOT FOUND VALUE IN MAP")
	os.Exit(-1)
	return -1
}

func pickRandomValueInMap(myMap map[int]string) string {

	var keys = make(map[int]int)

	for k := range myMap {
		keys[len(keys)] = k
	}

	var randKey int = rand.Intn(len(keys))

	return myMap[keys[randKey]]
}

//-------------------------
// THE SERVER FUNCTION
//-------------------------

func server(neighboursFilePath string, isStartingPoint bool) {

	// --- PREPARE THE CONFFIGURATION --- //

	var node yamlConfig = initAndParseFileNeighbours(neighboursFilePath)

	myLog(node.Address, "Starting server .... and listening ...")

	ln, err := net.Listen("tcp", node.Address+PORT)

	if err != nil {
		log.Fatal(err)

		return
	}

	rand.Seed(time.Now().UTC().UnixNano())

	// --- PREPARE VARIABLES --- //

	var ni string = node.Address

	var parent string = ""

	var F = make(map[int]string)  //set of children neighbours of ni
	var NF = make(map[int]string) //set of non-children neighbours of ni

	var v = make(map[int]string) //set of neighbours of ni
	for _, neighbour := range node.Neighbours {
		v[len(v)] = neighbour.Address
	}

	var NE = v //set of not yet explored neighbours of ni
	for k, v := range v {
		NE[k] = v
	}

	var terminated bool = false

	// --- THE ALGORITHM --- //

	myLog(node.Address, "Neighbours file parsing ...")
	myLog(node.Address, "Done")

	myLog(node.Address, "Starting algorithm ...")

	if isStartingPoint {
		parent = node.Address
		var nk = pickRandomValueInMap(NE)
		delete(NE, findMapIndexByValue(NE, nk))
		go sendToOneNeighbour("M", node, nk) // Sending M to nk !
	}

	for !terminated {

		conn, _ := ln.Accept() // BLOCKING !
		message, _ := bufio.NewReader(conn).ReadString('\n')
		conn.Close()

		remote_addr := conn.RemoteAddr().String()

		var nj = remote_addr[:len(remote_addr)-6] // Without the port number.

		myLog(node.Address, "Message received : "+message)

		switch message {

		case "M":

			if parent == "" {

				parent = nj
				delete(NE, findMapIndexByValue(NE, nj))

				if len(NE) != 0 {

					var nk = pickRandomValueInMap(NE)
					delete(NE, findMapIndexByValue(NE, nk))
					go sendToOneNeighbour("M", node, nk)
				} else {

					go sendToOneNeighbour("P", node, parent)
					terminated = true
				}
			} else {

				delete(NE, findMapIndexByValue(NE, nj))
				go sendToOneNeighbour("R", node, nj)
				NF[len(NF)] = nj
			}

		case "P", "R":

			if message == "P" {
				F[len(F)] = nj
			} else {
				NF[len(NF)] = nj
			}

			if len(NE) != 0 {

				var nk = pickRandomValueInMap(NE)
				delete(NE, findMapIndexByValue(NE, nk))
				go sendToOneNeighbour("M", node, nk) // Sending M to nk !
			} else {

				if parent != ni {

					go sendToOneNeighbour("P", node, parent)
				}

				terminated = true
			}
		}
	}

	if isStartingPoint {
		fmt.Println()
		myLog(node.Address, "All non-root nodes have been terminated. ")
		myLog(node.Address, "Root node has no more neighbours to explore.")
		myLog(node.Address, "Ending the algorithm ...")
	} else {
		fmt.Println("["+node.Address+"] :", "parent = "+parent, ", fils =", F, ", non-fils =", NF)
		myLog(node.Address, "No more neighbours to explore and parent has been contacted -> terminating node ...")
	}
}

//---------------------
// MAIN
//---------------------

func main() {

	//localadress := "127.0.0.1"

	go server("node-2.yaml", false)
	go server("node-3.yaml", false)
	go server("node-4.yaml", false)
	go server("node-5.yaml", false)
	go server("node-6.yaml", false)
	go server("node-7.yaml", false)
	go server("node-8.yaml", false)

	time.Sleep(2 * time.Second) // Waiting all node to be ready

	server("node-1.yaml", true)

	time.Sleep(1 * time.Second) // Waiting all console return from nodes
}
