package main

import (
	"fmt"
	"hotel/dbhandler"
	"hotel/session"
	"html/template"
	"log"
	"net/http"
)

func isUserConnected(r *http.Request) bool {
	_, _, err := session.GetSessionCookie(r)
	return err == nil
}

func Logout(w http.ResponseWriter) {
	session.ClearSessionCookie(w)
	fmt.Println("✅Déconnexion réussie. Cookie supprimé.")
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		IsConnected bool
	}{
		IsConnected: isUserConnected(r),
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		IsConnected bool
	}{
		IsConnected: isUserConnected(r),
	}

	if r.Method == "POST" {
		Email := r.FormValue("email")
		Password := r.FormValue("password")
		fmt.Println(dbhandler.Login(w, Email, Password))
	}

	_, _, err := session.GetSessionCookie(r)

	tmpl, err := template.ParseFiles("templates/connexion.html")
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		IsConnected bool
	}{
		IsConnected: isUserConnected(r),
	}
	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("mail")
		password := r.FormValue("mdp")

		err := dbhandler.Register(w, username, email, password)
		if err != nil {
			log.Println("❌ Erreur RegisterHandler:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	_, _, err := session.GetSessionCookie(r)

	tmpl, err := template.ParseFiles("templates/inscription.html")
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session.SetSessionCookie(w, -1, "", 0)

	log.Println("✅ Cookie supprimé (v2)")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ParcourirHandler(w http.ResponseWriter, r *http.Request) {
	hotels, err := dbhandler.GetAllHotels()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des hôtels", http.StatusInternalServerError)
		return
	}

	data := struct {
		IsConnected bool
		Hotels      []dbhandler.Hotel
	}{
		IsConnected: isUserConnected(r),
		Hotels:      hotels,
	}

	tmpl, err := template.ParseFiles("templates/parcourir.html")
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func main() {
	dbhandler.ConnectDB()
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/parcourir", ParcourirHandler)


	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fsUploads := http.FileServer(http.Dir("static/uploads"))
	http.Handle("/static/uploads/", http.StripPrefix("/static/uploads/", fsUploads))

	fmt.Println("Serveur démarré sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
