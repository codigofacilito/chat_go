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
    //"io/ioutil"
    "encoding/json"
)

var Users = make(map[string]User)
var UsersRWMutex sync.RWMutex

type User struct {
    websocket *websocket.Conn
    user_name string
}

type Request struct{
    user_name string  `json:"user_name"`
}

type Response struct{
    valid  bool `json:"valid"`
}

func main() {
    mux := mux.NewRouter()
    cssHandler := http.FileServer(http.Dir("./css/"))
    jsHandler := http.FileServer(http.Dir("./js/"))
    
    mux.HandleFunc("/", HomeHandler).Methods("GET")
    mux.HandleFunc("/ws/{user_name}", web_socket)
    mux.HandleFunc("/validate", validate).Methods("POST")

    http.Handle("/", mux)
    http.Handle("/css/", http.StripPrefix("/css/", cssHandler))
    http.Handle("/js/", http.StripPrefix("/js/", jsHandler))

    log.Println("Server running on :8000")
    log.Fatal(http.ListenAndServe(":8000", nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "index.html")
}

func validate(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    user_name := r.FormValue("user_name")

    response := Response{}
    if validate_user_name(user_name){
        response.valid = true
    }else{
        response.valid = false
    }
    json.NewEncoder(w).Encode(response)
}

func web_socket(w http.ResponseWriter, r *http.Request){
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if err != nil {
        log.Println(err)
        return
    }
    vars := mux.Vars(r)
    user := create_user(ws, vars["user_name"])
    add_user(user)
    for{
        type_message, message, err := ws.ReadMessage()
        if err != nil {
            remove_cliente(user.user_name)
            return
        }
        response_message := create_final_message(message, user.user_name)
        send_echo(type_message, response_message)
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

func create_final_message(message []byte, user_name string) []byte{
    message_string := string(message[:])
    return []byte(user_name + " : " + message_string) 
}

func send_echo(messageType int, message []byte) {
    UsersRWMutex.RLock()
    defer UsersRWMutex.RUnlock()

    for _, user := range Users {
        if err := user.websocket.WriteMessage(messageType, message); err != nil {
            return
        }
    }
}
