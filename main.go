package main

import (
	"context"
	"day9/connection"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func handleRequests() {
	route := mux.NewRouter()

	connection.DatabaseConnect()

	// router path folder untuk public
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	// routing
	route.HandleFunc("/", Home).Methods("GET")
	route.HandleFunc("/contact", Contact).Methods("GET")
	route.HandleFunc("/formProject", formProject).Methods("GET")
	route.HandleFunc("/detailProject/{id}", DetailProject).Methods("GET")
	route.HandleFunc("/addProject", addProject).Methods("POST")
	route.HandleFunc("/deleteProject/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/formEditProject/{id}", formEditProject).Methods("GET")
	route.HandleFunc("/editProject/{id}", editProject).Methods("POST")

	route.HandleFunc("/formRegister", formRegister).Methods("GET")
	route.HandleFunc("/register", Register).Methods("POST")

	route.HandleFunc("/formLogin", formLogin).Methods("GET")
	route.HandleFunc("/login", Login).Methods("POST")

	route.HandleFunc("/logout", logout).Methods("GET")


	fmt.Println("Go Running on Port 5000")
	http.ListenAndServe(":5000", route)
}

type Project struct {
	Title 					string
	StartDate 			time.Time
	EndDate 				time.Time
	StartDateFormat string
	EndDateFormat 	string
	DescFormat			string
	Duration				string
	Desc 						string
	Id							int
	Tech						[]string
	IsLogin					bool
}


type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

type SessionData struct {
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = SessionData{}

func Home(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Contect-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, f1 := range fm {
			
			flashes = append(flashes, f1.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	var result []Project
	data, _ := connection.Conn.Query(context.Background(), "SELECT id, name, start_date, end_date, description, technologies FROM tb_projects")
	for data.Next() {
		each := Project{}
		err := data.Scan(&each.Id, &each.Title, &each.StartDate, &each.EndDate, &each.Desc, &each.Tech)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		each.Duration = ""

		day :=  24 //in hours
		month :=  24 * 30 // in hours
		year :=  24 * 365 // in hours
		differHour := each.EndDate.Sub(each.StartDate).Hours()
		var differHours int = int(differHour)
		days := differHours / day
		months := differHours / month
		years := differHours / year
		if differHours < month {
			each.Duration = strconv.Itoa(int(days)) + " Days"
		} else if differHours < year {
			each.Duration = strconv.Itoa(int(months)) + " Months"
		} else if differHours > year {
			each.Duration = strconv.Itoa(int(years)) + " Years"
		}

		result = append(result, each)
	}

	response := map[string]interface{}{
		"Projects": result,
		"DataSession": Data,
	}

	if err == nil {
		tmpl.Execute(w, response)
	} else {
		w.Write([]byte("Message: "))
		w.Write([]byte(err.Error()))
	}
}


func Contact(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Contect-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/contact.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func formProject(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Contect-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/addProject.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func DetailProject(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Contect-Type", "text/html; charset=utf-8")
	
	var tmpl, err = template.ParseFiles("views/detailProject.html")
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	ProjectDetail := Project{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, start_date, end_date, description, technologies FROM tb_projects WHERE id=$1", id).Scan(&ProjectDetail.Id, &ProjectDetail.Title, &ProjectDetail.StartDate, &ProjectDetail.EndDate, &ProjectDetail.Desc, &ProjectDetail.Tech)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	
	ProjectDetail.StartDateFormat = ProjectDetail.StartDate.Format("02 Jan 2006")
	ProjectDetail.EndDateFormat = ProjectDetail.EndDate.Format("02 Jan 2006")
	ProjectDetail.Duration = ""

	day :=  24 //in hours
	month :=  24 * 30 // in hours
	year :=  24 * 365 // in hours
	differHour := ProjectDetail.EndDate.Sub(ProjectDetail.StartDate).Hours()
	var differHours int = int(differHour)
	days := differHours / day
	months := differHours / month
	years := differHours / year
	if differHours < month {
		ProjectDetail.Duration = strconv.Itoa(int(days)) + " Days"
	} else if differHours < year {
		ProjectDetail.Duration = strconv.Itoa(int(months)) + " Months"
	} else if differHours > year {
		ProjectDetail.Duration = strconv.Itoa(int(years)) + " Years"
	}

	response := map[string]interface{}{
		"Details" : ProjectDetail,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, response)
}


func addProject(w http.ResponseWriter, r *http.Request) {
	
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var inputTitle			string
	var inputStartDate 	string
	var inputEndDate 		string
	var inputDesc 			string
	var inputTech				[]string

	for i, values := range r.PostForm {
		for _ , value := range values {
			if i == "inputTitle" {
				inputTitle = value
			}
			if i == "inputStartDate" {
				inputStartDate = value
			}
			if i == "inputEndDate" {
				inputEndDate = value
			}
			if i == "inputDesc" {
				inputDesc = value
			}
			if i == "inputTech" {
				inputTech = append(inputTech, value)
			}
		}
	}

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_projects(name, start_date, end_date, description, technologies) VALUES ($1, $2, $3, $4, $5) ", inputTitle, inputStartDate, inputEndDate, inputDesc, inputTech)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
	}

	http.Redirect(w,r, "/", http.StatusMovedPermanently)
}


func editProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var inputTitle 			= r.PostForm.Get("inputTitle")
	var inputStartDate 	= r.PostForm.Get("inputStartDate")
	var inputEndDate 		= r.PostForm.Get("inputEndDate")
	var inputDesc 			= r.PostForm.Get("inputDesc")
	var inputTech 			[]string
	inputTech = r.PostForm["inputTech"]

	// for i, values := range r.PostForm {
	// 	for _, value := range values {
	// 		if i == "inputTitle"{
	// 			inputTitle = value
	// 		}
	// 		if i == "inputStartDate"{
	// 			inputStartDate = value
	// 		}
	// 		if i == "inputEndDate"{
	// 			inputEndDate = value
	// 		}
	// 		if i == "inputDesc" {
	// 			inputDesc = value
	// 		}
	// 		if i == "inputTech" {
	// 			inputTech = append(inputTech, value)
	// 		}
	// 	}

	// }
		_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_projects SET name=$1, start_date=$2, end_date=$3, description=$4, technologies=$5 WHERE id=$6", inputTitle, inputStartDate, inputEndDate, inputDesc, inputTech, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func formEditProject(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/editMyProject.html")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	ProjectEdit := Project{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, start_date, end_date, description, technologies FROM tb_projects WHERE id=$1", id).Scan(&ProjectEdit.Id, &ProjectEdit.Title, &ProjectEdit.StartDate, &ProjectEdit.EndDate, &ProjectEdit.Desc, &ProjectEdit.Tech)

	ProjectEdit.StartDateFormat = ProjectEdit.StartDate.Format("2006-01-02")
	ProjectEdit.EndDateFormat = ProjectEdit.EndDate.Format("2006-01-02")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message :" + err.Error()))
	}

	response := map[string]interface{}{
		"Project": ProjectEdit,
	}

		if err == nil {
		tmpl.Execute(w, response)
	} else {
		panic(err)
	}
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message :" + err.Error()))
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("views/register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func Register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := r.PostForm.Get("inputName")
	email := r.PostForm.Get("inputEmail")
	password := r.PostForm.Get("inputPassword")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	fmt.Println(passwordHash)

	_ , err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_users(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/formLogin", http.StatusMovedPermanently)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("views/login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	
	store := sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, f1 := range fm {
			flashes = append(flashes, f1.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")
	tmpl.Execute(w, Data)
}

func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	email := r.PostForm.Get("inputEmail")
	password := r.PostForm.Get("inputPassword")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(),
		"SELECT * FROM tb_users WHERE email=$1", email).Scan(&user.ID, &user.Name,&user.Email, &user.Password)

	if err != nil {
		store := sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")

		session.AddFlash("Email belum terdaftar!", "message")
		session.Save(r, w)

		http.Redirect(w, r, "/formLogin", http.StatusMovedPermanently)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		store := sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")

		session.AddFlash("Password Salah!", "message")
		session.Save(r, w)

		http.Redirect(w, r, "/formLogin", http.StatusMovedPermanently)

		return
	}

	store := sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email
	session.Values["ID"] = user.ID
	session.Values["IsLogin"] = true
	session.Options.MaxAge = 10800 // 3 JAM expred

	session.AddFlash("LOGIN BERHASIL", "message")
	
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {
	store := sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/formLogin", http.StatusSeeOther)
}

func main() {
	handleRequests() 
}