package auth

import (
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type authstore interface {
	GetUser(id string) (*Account, error)
	LoadAccounts() ([]Account, error)
}

// func GetUser(id string) (*Account, error) {

// 	content, err := os.ReadFile("./backend/accounts/" + id + ".json")
// 	if err != nil {
// 		if errors.Is(err, os.ErrNotExist) {
// 			return nil, ErrUserNotFound
// 		}
// 		return nil, err
// 	}
// 	// We create another instance of `Credentials` to store the credentials we get from the database
// 	var account Account
// 	err = json.Unmarshal(content, &account)
// 	if err != nil {

// 		return nil, err
// 	}
// 	return &account, nil
// }

// func StoreToken(token *hi.Token) error {
// 	bytes, err := json.MarshalIndent(token, "", "    ")
// 	if err != nil {
// 		return err
// 	}
// 	err = os.WriteFile("../backend/hitokens/"+token.Alias+".json", bytes, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func LoadTokens() map[string]*hi.Token {

// 	m := map[string]*hi.Token{}
// 	files, _ := os.ReadDir("../backend/hitokens/")
// 	for _, file := range files {
// 		if !strings.HasSuffix(file.Name(), ".json") {
// 			continue
// 		}
// 		content, err := os.ReadFile("../backend/hitokens/" + file.Name())
// 		if err != nil {
// 			fmt.Println("could not read hitoken file", file.Name(), err.Error())
// 			continue
// 		}
// 		var token hi.Token
// 		err = json.Unmarshal(content, &token)
// 		if err != nil {
// 			fmt.Println("could not parse file", file.Name())
// 			continue
// 		}
// 		m[token.Alias] = &token
// 	}

// 	return m
// }

// func LoadAccounts() []Account {
// 	var accounts []Account
// 	files, _ := os.ReadDir("../backend/accounts/")
// 	for _, file := range files {
// 		if !strings.HasSuffix(file.Name(), ".json") {
// 			continue
// 		}

// 		content, err := os.ReadFile("../backend/accounts/" + file.Name())
// 		if err != nil {
// 			fmt.Println("could not read account file", file.Name(), err.Error())
// 			continue
// 		}

// 		// We create another instance of `Credentials` to store the credentials we get from the database
// 		var account Account
// 		err = json.Unmarshal(content, &account)
// 		if err != nil {
// 			fmt.Println("could not parse account file", file.Name())
// 			continue
// 		}
// 		accounts = append(accounts, account)
// 	}

// 	return accounts
// }
