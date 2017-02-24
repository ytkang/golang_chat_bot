package main

import (
	"golang.org/x/net/websocket"
	"html/template"
	"log"
	"net/http"
	"os"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/ytkang/golang_chat_bot/jarvis"
	"github.com/ytkang/golang_chat_bot/network"
	"strings"
	"net"
	"crypto/tls"
)

const (
	Host     = "cluster0-shard-00-00-rgvbm.mongodb.net:27017"
	//Host 	 = "mongodb://cluster0-shard-00-00-rgvbm.mongodb.net:27017,cluster0-shard-00-01-rgvbm.mongodb.net:27017,cluster0-shard-00-02-rgvbm.mongodb.net:27017/test?replicaSet=Cluster0-shard-0"
	Username = "mongo"
	Password = "mongo123"
	Database = "admin"
)

var (
	pwd, _        = os.Getwd()
	RootTemp      = template.Must(template.ParseFiles(pwd + "/chat.html"))
	JSON          = websocket.JSON           // codec for JSON
	Message       = websocket.Message        // codec for string, []byte
	ActiveClients = make(map[network.ClientConn]int) // map containing clients
	mongo	      *mgo.Session = nil
	jarvis	*SmartJarvis.Jarvis
	listenAddr = "0.0.0.0:"+os.Getenv("PORT") // server address
)

// Initialize handlers and websocket handlers
func init() {
	http.HandleFunc("/", RootHandler)
	http.Handle("/sock", websocket.Handler(SockServer))
}

// WebSocket server to handle chat between clients
func SockServer(ws *websocket.Conn) {
	var err error
	var clientMessage string
	// use []byte if websocket binary type is blob or arraybuffer
	// var clientMessage []byte

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			log.Println("Websocket could not be closed", err.Error())
		}
	}()

	client := ws.Request().RemoteAddr
	log.Println("Client connected:", client)
	sockCli := network.ClientConn{ws, client}
	ActiveClients[sockCli] = 0
	log.Println("Number of clients connected ...", len(ActiveClients))
	// for loop so the websocket stays open otherwise
	// it'll close after one Receieve and Send
	for {
		if err = Message.Receive(ws, &clientMessage); err != nil {
			// If we cannot Read then the connection is closed
			log.Println("Websocket Disconnected waiting", err.Error())
			// remove the ws client conn from our active clients
			delete(ActiveClients, sockCli)
			log.Println("Number of clients still connected ...", len(ActiveClients))
			return
		}

		sendingMessage := sockCli.ClientIP + " Said: " + clientMessage
		clientMessage = strings.Replace(clientMessage, "\n", "", -1)
		if len(clientMessage) == 0 {
			continue
		}

		if !strings.Contains(clientMessage, "A:") && !strings.Contains(clientMessage, "a:") {
			for cs, _ := range ActiveClients {
				// go Message.Send(cs.websocket, clientMessage) // DO NOT THIS! This handler is already called from go routine
				if mongo != nil {
					var msg network.Msg = network.Msg{}
					c := mongo.DB("chat").C("messages")
					change := mgo.Change{
						Update: bson.M{"text": clientMessage},
						Upsert: true,
					}
					info, err := c.Find(bson.M{"text": clientMessage}).Apply(change, &msg)

					//err = c.Insert(&Msg{Text:clientMessage})
					if err != nil {
						log.Println("Error: could not insert new message!")
						log.Panic(err)
					}else if info.UpsertedId != nil {
						jarvis.SetPrevMessageId(info.UpsertedId.(bson.ObjectId))
					} else {
						jarvis.SetPrevMessageId(msg.ID)
					}
				} else {
					log.Println("mongo db is nil")
				}

				log.Println("will send broadcast message to clients")
				if err = Message.Send(cs.Websocket, sendingMessage); err != nil {
					// we could not send the message to a peer
					log.Println("Could not send message to ", cs.ClientIP, err.Error())
				} else {
					log.Println("broadcast done")
				}
			}
		}

		if mongo != nil {
			go jarvis.Answer(&Message, ActiveClients, clientMessage, mongo)
		}
	}
}

// RootHandler renders the template for the root page
func RootHandler(w http.ResponseWriter, req *http.Request) {
	err := RootTemp.Execute(w, listenAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	log.Println("MongoDB connection..")
	session, m_err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{Host},
		Username: Username,
		Password: Password,
		Database: Database,
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{})
		},
	})

	if m_err != nil {
		panic(m_err)
		log.Println("[Error] MongoDB connecting failed")
	} else {
		log.Println("MongoDB connected")
	}
	mongo = session
	mongo.SetMode(mgo.Monotonic, true)
	jarvis = SmartJarvis.NewJarvis()
	log.Println("Starting..", listenAddr)
	
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
