package client

import (
	"bufio"
	"fmt"
	"os"

	"github.com/celestix/gotgproto"
)

type Conversator struct{}

func (Conversator) AskPhoneNumber() (string, error) {
	fmt.Println("phone number:")
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return text, err
}

func (Conversator) AskCode() (string, error) {
	fmt.Println("code:")
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return text, err
}

func (Conversator) AskPassword() (string, error) {
	fmt.Println("password:")
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return text, err
}
func (Conversator) AuthStatus(authStatus gotgproto.AuthStatus) {
	fmt.Printf("%+v\n", authStatus.Event)
}
