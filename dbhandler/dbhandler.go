package dbhandler

import (
	"database/sql"
	"errors"
	"fmt"
	"hotel/session"
	"log"
	"net/http"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

type Hotel struct {
	ID            int
	Name          string
	Image         string
	Description   string
	PricePerNight float64
}

func GetAllHotels() ([]Hotel, error) {
	query := "SELECT id, name, image, description, price_per_night FROM hotels"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotels []Hotel

	for rows.Next() {
		var h Hotel
		err := rows.Scan(&h.ID, &h.Name, &h.Image, &h.Description, &h.PricePerNight)
		if err != nil {
			return nil, err
		}
		hotels = append(hotels, h)
	}

	return hotels, nil
}

// ConnectDB initialise la connexion à la base de données MySQL et configure les paramètres de connexion.
func ConnectDB() {
	// Format: "user:password@tcp(host:port)/dbname"
	dsn := "root:@tcp(localhost:3306)/quadrivago?parseTime=true"

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌Erreur de connexion à la base de données:", err)
	}

	DB.SetConnMaxLifetime(5 * time.Minute) // connecter pour 5 min
	DB.SetMaxOpenConns(10)                 // 10 personnes connecter en même temps
	DB.SetMaxIdleConns(5)                  // 5 connection innactive ouverte en même temps

	if err = DB.Ping(); err != nil {
		log.Fatal("❌Impossible de pinger la base de données:", err)
	}

	fmt.Println("Connexion réussie à MySQL !")

}

// ---------------------------------------------------------------inscription + connection sécurisé----------------------------------------------------
// Register enregistre un nouvel utilisateur après avoir vérifié l'email, haché le mot de passe, et crée une session.
func Register(w http.ResponseWriter, username, email, password string) error {
	if DB == nil {
		return errors.New("❌ La base de données n’est pas connectée")
	}

	if !valideEmail(email) {
		return errors.New("❌Format d'email invalide")
	}

	existe, err := verifemail(DB, email)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existe {
		return errors.New("❌Cet email est déjà utilisé")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := "INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)"
	res, err := DB.Exec(query, username, email, string(hash))
	if err != nil {
		return errors.New("❌Erreur lors de l'inscription: " + err.Error())
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return errors.New("❌Erreur lors de la récupération de l'ID utilisateur")
	}

	session.SetSessionCookie(w, 24*time.Hour, username, int(userID))
	log.Println("✅ Utilisateur inscrit avec succès :", username)
	return nil
}

// utiliser pour verifier si l'email existe deja dans la BDD

func verifemail(db *sql.DB, email string) (bool, error) {
	var num int
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&num)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return true, err
}

// valideEmail vérifie si une adresse email a un format valide à l’aide d’une expression régulière.
func valideEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)

	return re.MatchString(email)
}

// Login permet à un utilisateur de se connecter en vérifiant son email et son mot de passe, puis crée une session.
func Login(w http.ResponseWriter, email, password string) error {
	var userID int
	var hashpass string
	var username string

	if !valideEmail(email) {
		return errors.New("❌Format d'email invalide")
	}

	query := "SELECT id, password_hash, name FROM users WHERE email = ?"
	err := DB.QueryRow(query, email).Scan(&userID, &hashpass, &username)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("❌Utilisateur non trouvé")
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	if err != nil {
		return errors.New("❌Mot de passe incorrect")
	}

	session.SetSessionCookie(w, 24*time.Hour, username, userID)
	fmt.Println("✅Connexion réussie ! Cookie de session créé.")
	return nil
}
