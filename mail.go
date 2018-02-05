package main

import (
	"fmt"
	"net/smtp"
)

func SendEmail(to, body string) {
	from := "test0120181@gmail.com"
	pass := "qwertyAsdfg"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Bad metrics param" + "\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		fmt.Println("error sending")
	}
}
