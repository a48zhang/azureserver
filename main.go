package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var _db map[string]User

type User struct {
	Name     string
	Nickname string
	Pwd      string
	cookie   http.Cookie
}

// Return User's Info from HTTP Body
func getusr(r *http.Request) User {
	len := r.ContentLength
	info := make([]byte, len)
	r.Body.Read(info)
	var usrInfo User
	json.Unmarshal(info, &usrInfo)
	return usrInfo
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
}

func Register(w http.ResponseWriter, r *http.Request) {
	usrInfo := getusr(r)

	if _, ok := _db[usrInfo.Name]; !ok {
		_db[usrInfo.Name] = usrInfo
		fmt.Fprintln(w, "Register successfully")
		fmt.Fprintln(w, usrInfo.Name)
		fmt.Println(usrInfo.Name, "has registered")
		fmt.Println(usrInfo.Pwd, "has registered")
	} else {
		fmt.Fprintln(w, "Username has been occupied")
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	usrInfo := getusr(r)

	if _db[usrInfo.Name].Pwd == usrInfo.Pwd {
		fmt.Println(usrInfo.Name, "logging in")
		fmt.Println(usrInfo.Pwd)
		fmt.Println(_db[usrInfo.Name].Pwd)
		fmt.Println(_db[usrInfo.Name].Pwd == usrInfo.Pwd)
		cookie, err := r.Cookie(usrInfo.Name)
		if cookie == nil && err != nil {
			rand.Seed(time.Now().UnixNano())
			//设置Cookie
			cookie := &http.Cookie{
				Name:     usrInfo.Name,
				Value:    strconv.Itoa(rand.Intn(1000)),
				MaxAge:   60 * 60,
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			http.SetCookie(w, cookie)
			usrInfo.cookie = *cookie
			_db[usrInfo.Name] = usrInfo
			fmt.Fprintf(w, "Welcome,%s\n", usrInfo.Name)
			fmt.Println(usrInfo.Name, "has logged in")
		} else {
			fmt.Fprintf(w, "You have logged in,%s\n", cookie.Name)
		}
	} else {
		fmt.Fprintln(w, "Wrong Username or Password")
	}
}
func ListInfo(w http.ResponseWriter, r *http.Request) {
	usrInfo := getusr(r)

	cookie, err := r.Cookie(usrInfo.Name)
	if cookie.Value == _db[usrInfo.Name].cookie.Value && err == nil {
		usrInfo = _db[cookie.Name]
		fmt.Fprintf(w, "Welcome,%s\n", usrInfo.Name)
		fmt.Fprintf(w, "Nickname: %s\n", usrInfo.Nickname)
		fmt.Fprintf(w, "Password: %s\n", usrInfo.Pwd)
	} else {
		fmt.Fprintf(w, "Please login\n")
	}
}
func ChInfo(w http.ResponseWriter, r *http.Request) {
	usrInfo := getusr(r)

	cookie, err := r.Cookie(usrInfo.Name)
	if cookie.Value == _db[usrInfo.Name].cookie.Value && err == nil {
		fmt.Fprintf(w, "Welcome,%s\n", usrInfo.Name)
		fmt.Fprintf(w, "Origin nickname: %s\n", usrInfo.Nickname)
		_db[cookie.Name] = usrInfo
		fmt.Fprintf(w, "New nickname: %s\n", usrInfo.Nickname)
	} else {
		fmt.Fprintf(w, "Please login\n")
	}
}

func ChPwd(w http.ResponseWriter, r *http.Request) {
	usrInfo := getusr(r)

	cookie, err := r.Cookie(usrInfo.Name)
	if cookie.Value == _db[usrInfo.Name].cookie.Value && err == nil {
		fmt.Fprintf(w, "Welcome,%s\n", usrInfo.Name)
		fmt.Fprintf(w, "Origin password: %s\n", usrInfo.Pwd)
		_db[cookie.Name] = usrInfo
		fmt.Fprintf(w, "New password: %s\n", usrInfo.Pwd)
	} else {
		fmt.Fprintf(w, "Please login\n")
	}
}

func main() {
	_db = make(map[string]User)
	http.HandleFunc("/", Index)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/user", ListInfo)
	http.HandleFunc("/user/changeInfo", ChInfo)
	http.HandleFunc("/user/changePwd", ChPwd)
	fmt.Println("Server online")
	http.ListenAndServe(":9090", nil)
}
