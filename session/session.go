package session

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func SetSessionCookie(w http.ResponseWriter, duration time.Duration, username string, userID int) {
	expiration := time.Now().Add(duration)
	cookie := http.Cookie{
		Name:     "cookie",
		Value:    fmt.Sprintf("%s|%d", username, userID),
		Expires:  expiration,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
}

func GetSessionCookie(r *http.Request) (string, int, error) {
	cookie, err := r.Cookie("cookie")
	if err != nil {
		fmt.Println(err, "❌ Erreur lors de la récupération du cookie")
		return "", 0, err
	}
	fmt.Println("✅ Cookie reçu :", cookie.Value)

	var username string
	var userID int
	parts := strings.Split(cookie.Value, "|")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("❌ Format de cookie invalide")
	}

	username = parts[0]
	userID, err = strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("❌ Erreur de conversion de l'ID utilisateur")
	}

	fmt.Println("🎉 Session récupérée → Utilisateur :", username, ", ID :", userID)
	return username, userID, nil
}

func ClearSessionCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     "cookie",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
	fmt.Println("✅ Cookie supprimé.")
}
