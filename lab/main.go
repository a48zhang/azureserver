// Go connection Sample Code:
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

var db *sql.DB

func openDB() {
	var server = "azuredbvlad.database.windows.net"
	var port = 1433
	var user = "vlad"
	var password = "Windows@azure"
	var database = "test"
	// Build connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		server, user, password, port, database)
	var err error
	// Create connection pool
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("Connected!")
}

var session map[string]string

type User struct {
	id       int
	Name     string
	Pwd      string
	Nickname string
}

// Return User's Info from HTTP Body
func getUsr(r *http.Request) (User, error) {
	len := r.ContentLength
	info := make([]byte, len)
	_, err := r.Body.Read(info)
	var usrInfo User
	json.Unmarshal(info, &usrInfo)
	return usrInfo, err
}

// Return User's Info from Database
func find(name string) (User, error) {
	res := db.QueryRow("SELECT id,Pwd,Nickname FROM usrinfo WHERE name = ?", name)
	var info User
	err := res.Scan(&info.id, &info.Pwd, &info.Nickname)
	info.Name = name
	return info, err
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
}

func Register(w http.ResponseWriter, r *http.Request) {
	usrInfo, _ := getUsr(r)

	if _, ok := find(usrInfo.Name); ok == sql.ErrNoRows {
		db.Exec("INSERT INTO usrinfo (`Name`,`Pwd`,`Nickname`) VALUE (?,?,?)", usrInfo.Name, usrInfo.Pwd, usrInfo.Nickname)
		fmt.Fprintln(w, "Register successfully")
		fmt.Fprintln(w, usrInfo.Name)
	} else {
		fmt.Fprintln(w, "Username has been occupied")
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	usrInfo, _ := getUsr(r)

	var pwd string
	res := db.QueryRow("SELECT Pwd FROM usrinfo WHERE name = ?", usrInfo.Name)
	err := res.Scan(&pwd)
	if err == sql.ErrNoRows {
		fmt.Fprintln(w, "Wrong Username or Password")
		return
	}
	if pwd == usrInfo.Pwd {
		rand.Seed(time.Now().UnixNano())
		//Setting Cookie
		cookie := &http.Cookie{
			Name:     usrInfo.Name,
			Value:    strconv.Itoa(rand.Intn(1000)),
			MaxAge:   60 * 60,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		session[usrInfo.Name] = cookie.Value
		fmt.Fprintf(w, "Welcome,%s\n", usrInfo.Name)
		fmt.Println(usrInfo.Name, "has logged in")

	} else {
		fmt.Fprintln(w, "Wrong Username or Password")
		return
	}
}

func ListInfo(w http.ResponseWriter, r *http.Request) {
	usrInfo, _ := getUsr(r)

	cookie, _ := r.Cookie(usrInfo.Name)
	if cookie.Value == session[usrInfo.Name] {
		usrInfo, _ = find(usrInfo.Name)
		fmt.Fprintf(w, "Welcome,%s\n", usrInfo.Name)
		fmt.Fprintf(w, "id:%d\n", usrInfo.id)
		fmt.Fprintf(w, "Nickname: %s\n", usrInfo.Nickname)
		fmt.Fprintf(w, "Password: %s\n", usrInfo.Pwd)
	} else {
		fmt.Fprintf(w, "Please login\n")
	}
}

func ChInfo(w http.ResponseWriter, r *http.Request) {
	usrInfo, _ := getUsr(r)

	cookie, _ := r.Cookie(usrInfo.Name)
	if cookie.Value == session[usrInfo.Name] {
		fmt.Fprintf(w, "Welcome,%s\n", usrInfo.Name)
		info, _ := find(usrInfo.Name)
		fmt.Fprintf(w, "Origin nickname: %s\n", info.Nickname)
		db.Exec("UPDATE usrinfo SET Nickname = ? WHERE id = ?", usrInfo.Nickname, info.id)
		fmt.Fprintf(w, "New nickname: %s\n", usrInfo.Nickname)
	} else {
		fmt.Fprintf(w, "Please login\n")
	}
}

func ChPwd(w http.ResponseWriter, r *http.Request) {
	usrInfo, _ := getUsr(r)

	cookie, _ := r.Cookie(usrInfo.Name)
	if cookie.Value == session[usrInfo.Name] {
		fmt.Fprintf(w, "Welcome,%s\n", usrInfo.Name)
		info, _ := find(usrInfo.Name)
		fmt.Fprintf(w, "Origin password: %s\n", info.Pwd)
		db.Exec("UPDATE usrinfo SET Pwd = ? WHERE id = ?", usrInfo.Pwd, info.id)
		fmt.Fprintf(w, "New password: %s\n", usrInfo.Pwd)
	} else {
		fmt.Fprintf(w, "Please login\n")
	}
}

func main() {
	session = make(map[string]string)
	http.HandleFunc("/", Index)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/user", ListInfo)
	http.HandleFunc("/user/changeInfo", ChInfo)
	http.HandleFunc("/user/changePwd", ChPwd)
	openDB()
	http.ListenAndServe(":9090", nil)
}
