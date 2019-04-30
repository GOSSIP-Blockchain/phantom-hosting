package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/carsenk/phantom-hosting/utils"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

var db bolt.DB

func Register(w http.ResponseWriter, r *http.Request) {
	var account Account

	// Create db
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		fmt.Println(err)
	}

	defer db.Close()

	err = json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		utils.Respond(w, nil, err)
		return
	}

	if len(account.Password) < 6 {
		err := errors.New("Password needs to be more than 6 characters")
		utils.Respond(w, nil, err)
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("usersBucket"))
		if err != nil {
			return err
		}

		encoded, err := json.Marshal(account)
		if err != nil {
			return err
		}

		return b.Put([]byte(account.Username), encoded)
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	utils.Respond(w, account, nil)
}
