package database

import (
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) string {
	pass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(pass)
}

func Seed() {
	queryAddAdmins := "INSERT INTO admins (first_name, last_name, email, password, role, status) VALUES (?,?,?,?,?,?)"
	queryAddClients := "INSERT INTO clients (first_name, last_name, email, password, status) VALUES (?,?,?,?,?)"
	queryAddClientData := "INSERT INTO client_details (client_id, phone, company, company_phone, company_address, company_website, hours_of_operation, notes) VALUES (?,?,?,?,?,?,?,?)"

	type admin struct {
		firstName string
		lastName  string
		email     string
		password  string
		role      string
		status    string
	}

	type client struct {
		firstName        string
		lastName         string
		email            string
		phone            string
		company          string
		companyPhone     string
		companyAddress   string
		companyWebsite   string
		hoursOfOperation string
		notes            string
		password         string
		status           string
	}

	var admins = []admin{
		admin{
			firstName: "John",
			lastName:  "Doe",
			email:     "super@test.com",
			password:  "password",
			role:      "superadmin",
			status:    "active",
		},
		admin{
			firstName: "Jane",
			lastName:  "Doe",
			email:     "admin@test.com",
			password:  "password",
			role:      "admin",
			status:    "active",
		},
	}

	var clients = []client{
		client{
			firstName:        "Joe",
			lastName:         "Blow",
			email:            "client@test.com",
			phone:            "867-5309",
			company:          "acme goods",
			companyPhone:     "123-4567",
			companyAddress:   "123 Main st, Anytown, Anystate 12345",
			companyWebsite:   "www.testing.com",
			hoursOfOperation: "Open every day that ends in Y",
			notes:            "test note",
			password:         "password",
			status:           "active",
		},
	}

	for _, adminData := range admins {
		stmt, err := Client.Prepare(queryAddAdmins)
		if err != nil {
			log.Error("error when trying to insert admin data", err)
		}
		_, saveErr := stmt.Exec(adminData.firstName, adminData.lastName, adminData.email, hashPassword(adminData.password), adminData.role, adminData.status)
		if saveErr != nil {
			log.Error("error when trying to save admin", saveErr)
		}
	}

	for _, clientData := range clients {
		stmt, err := Client.Prepare(queryAddClients)
		if err != nil {
			log.Error("error when trying to insert client", err)
		}
		res, saveErr := stmt.Exec(clientData.firstName, clientData.lastName, clientData.email, hashPassword(clientData.password), clientData.status)
		if saveErr != nil {
			log.Error("error when trying to save client", saveErr)
		}

		lid, err := res.LastInsertId()
		if err != nil {
			log.Error("error when trying to save client data", err)
		}

		stmt, err = Client.Prepare(queryAddClientData)
		if err != nil {
			log.Error("error when trying to insert client data", err)
		}
		_, saveErr = stmt.Exec(lid, clientData.phone, clientData.company, clientData.companyPhone, clientData.companyAddress, clientData.companyWebsite, clientData.hoursOfOperation, clientData.notes)

		if saveErr != nil {
			log.Error("error when trying to save client", saveErr)
		}
	}
}
