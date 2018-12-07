package thisis

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"log"
)

var DB *sql.DB

type Result int

const (
	Success         Result = 0
	WrongPassword   Result = 1
	DatabaseFailure Result = 2
	NotExist        Result = 3
	Exist           Result = 4
	BcryptFailure   Result = 5
)

type Link struct {
	OriginalURL, ShortUrl, Password string
}

func PasswordHash(password string) []byte {
	phash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal("bcrypt failure")
	}
	return phash
}

func AddShortLink(link Link) Result {
	phash := PasswordHash(link.Password)
	_, err := DB.Query("INSERT INTO URLStorage (ShortedURL, OriginalURL, ChangeKey) VALUES (?,?,?)",
		link.ShortUrl, link.OriginalURL, phash)
	if err != nil {
		return DatabaseFailure
	}
	return Success
}

func IsShortLinkExist(link Link) Result {
	var exists bool
	err := DB.QueryRow("SELECT exists(SELECT OriginalURL FROM URLStorage WHERE ShortedURL=?)",
		link.ShortUrl).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return NotExist
		} else {
			return DatabaseFailure
		}
	} else {
		if exists {
			return Exist
		}
		return NotExist
	}
}

func CheckLinkPassword(link Link) Result {
	var keyhash string

	err := DB.QueryRow("SELECT ChangeKey FROM URLStorage WHERE ShortedURL=?",
		link.ShortUrl).Scan(&keyhash)
	if err != nil {
		if err == sql.ErrNoRows {
			return NotExist
		} else {
			return DatabaseFailure
		}
	}
	res := bcrypt.CompareHashAndPassword([]byte(keyhash), []byte(link.Password))
	if res == nil {
		return Success
	} else if res == bcrypt.ErrMismatchedHashAndPassword {
		return WrongPassword
	} else {
		return BcryptFailure
	}
}

func UpdateShortLink(link Link) Result {
	_, err := DB.Query("UPDATE URLStorage SET OriginalURL=? WHERE ShortedURL=?",
		link.OriginalURL, link.ShortUrl)
	if err != nil {
		return DatabaseFailure
	}
	return Success
}

func DeleteShortLink(link Link) Result {
	_, err := DB.Query("DELETE FROM URLStorage WHERE ShortedURL=?",
		link.ShortUrl)
	if err != nil {
		return DatabaseFailure
	}
	return Success
}

func GetShortLink(link *Link) Result {
	err := DB.QueryRow("SELECT OriginalURL FROM URLStorage WHERE ShortedURL=?",
		link.ShortUrl).Scan(link.OriginalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return NotExist
		} else {
			return DatabaseFailure
		}
	} else {
		return Success
	}
}

func InitDB() {
	DB.Query(`CREATE TABLE IF NOT EXISTS URLStorage(
  ID           INT            NOT NULL    AUTO_INCREMENT,
  ShortedURL   VARCHAR(45),
  OriginalURL  TEXT,
  ChangeKey    VARCHAR(60),
  PRIMARY KEY (ID))`)
}

func ConnectToDB(DSN string) {
	tdb, err := sql.Open("mysql", DSN)
	DB = tdb
	if err != nil {
		log.Fatal("DB Connection Failed")
		panic(err.Error())
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal("DB Ping Failed")
		panic(err.Error())
	}

}
