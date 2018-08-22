package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) //klien yang tersambung
var broadcast = make(chan Message)           //channel broadcast
var upgrader = websocket.Upgrader{}          //konfigurasi upgrader

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	//membuat file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)
	//konfigurasi route websocket
	http.HandleFunc("/ws", handleConnections)
	//listening messages
	go handleMessages()

	log.Println("Server berjalan pada port : 8888")
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal("ListenAndServe : ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)

	}
	defer ws.Close()
	clients[ws] = true
	for {
		var msg Message
		// Membaca pesan masuk baru sebagai JSON dan memetakannya kedalam Objek Pesan
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Mengirim pesan yang baru saja diterima ke channel broadcast
		broadcast <- msg
	}
}
func handleMessages() {
	for {
		// Ambil pesan selanjutnya dari broadcast channel
		msg := <-broadcast
		// Kirim ke setiap klien yang terkoneksi
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
