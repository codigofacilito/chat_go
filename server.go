package main

/*
	Instalación 
    go get github.com/gorilla/mux
    go get github.com/gorilla/websocket
*/
import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
    "sync"
)

var Users = make(map[string]User)
var UsersRWMutex sync.RWMutex

type User struct {
    websocket *websocket.Conn
    user_name string
}

type request struct {
    user_name string `json:"user_name"`
}

func main() {
    mux := mux.NewRouter()
    cssHandler := http.FileServer(http.Dir("./css/"))
    jsHandler := http.FileServer(http.Dir("./js/"))
    
    mux.HandleFunc("/", HomeHandler).Methods("GET")
    mux.HandleFunc("/ws/{user_name}", web_socket)

    http.Handle("/", mux)
    http.Handle("/css/", http.StripPrefix("/css/", cssHandler))
    http.Handle("/js/", http.StripPrefix("/js/", jsHandler))

    log.Println("Server running on :8000")
    log.Fatal(http.ListenAndServe(":8000", nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "index.html")
}


func web_socket(w http.ResponseWriter, r *http.Request){
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if err != nil {
        log.Println(err)
        return
    }
    vars := mux.Vars(r)
    if !validate_user_name(vars["user_name"]){
        log.Println("El nombre ya esta en uso ")
        http.Error(w, "User name in use", 202)
        return
    }
    user := create_user(ws, vars["user_name"])
    add_user(user)
    for{
        type_message, message, err := ws.ReadMessage()
        if err != nil {
            remove_cliente(user.user_name)
            return
        }
        send_echo(type_message, message, user.user_name)
    }
}

func validate_user_name(user_name string) bool{
    UsersRWMutex.Lock()
    defer UsersRWMutex.Unlock()
    if _, ok := Users[user_name]; ok {
        return false
    }
    return true
}

func remove_cliente(user_name string) {
    UsersRWMutex.Lock()
    delete(Users, user_name)
    UsersRWMutex.Unlock()
}

func create_user(ws *websocket.Conn, usuario string) User{
    return User{ websocket: ws, user_name: usuario}
}

func add_user(user User){
    UsersRWMutex.Lock()
    defer UsersRWMutex.Unlock()
    Users[user.user_name] = user
}

func send_echo(messageType int, message []byte, user_name string) {
    UsersRWMutex.RLock()
    defer UsersRWMutex.RUnlock()

    origin_message := string(message[:])
    final_message:= user_name + " : " + origin_message
    final_bite := []byte(final_message)

    for _, user := range Users {
        if err := user.websocket.WriteMessage(messageType, final_bite); err != nil {
            return
        }
    }
}
