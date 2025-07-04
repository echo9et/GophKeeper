package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"time"

	"GophKeeper.ru/internal/client/gcrypto"
	"GophKeeper.ru/internal/entities"
	"GophKeeper.ru/internal/utils"
	"golang.org/x/exp/slog"
)

type GophKeeper struct {
	Client      *http.Client
	user        string
	pass        string
	addr        string
	key         string
	data        []entities.Record
	countUpdate int
}

// NewGophKeeper - конструктор
func NewGophKeeper(user, pass, addr, secret_key, certFile string) (*GophKeeper, error) {
	cert, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		log.Fatalf("unable to parse cert from %s", certFile)
		return nil, err
	}

	gophKeeper := GophKeeper{
		addr: addr,
		user: user,
		pass: pass,
		key:  secret_key,
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	gophKeeper.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
		Jar: jar,
	}

	return &gophKeeper, nil
}

// auth - авторизация по лоигину и паролю
func (g *GophKeeper) auth() error {
	user := entities.User{
		Login:    g.user,
		Password: g.pass,
	}
	user.Password = utils.Sha256hash(user.Password)
	jsonValue, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "https://"+g.addr+"/api/auth", bytes.NewBuffer(jsonValue))

	if err != nil {
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) auth() error")
		return err
	}

	resp, err := g.Client.Do(req)
	if err != nil {
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) auth() error")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("%v", resp.StatusCode)
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) auth() error")
		return err
	}

	return nil
}

// UpdateRecord - добавление\обновление данных
func (g *GophKeeper) UpdateRecord(key, value string) error {

	eKey, err := gcrypto.Encrypt(key, g.key, "GophKeeper")
	if err != nil {
		err = errors.New("шифрование ключа завершилось с ошибкой")
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) UpdateRecord(key, value string) error")
		return err
	}

	eValue, err := gcrypto.Encrypt(value, g.key, "GophKeeper")
	if err != nil {
		err = errors.New("шифрование значения завершилось с ошибкой")
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) UpdateRecord(key, value string) error")
		return err
	}

	record := entities.Record{
		Key:   eKey,
		Value: eValue,
	}

	data, err := json.Marshal(record)
	if err != nil {
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) UpdateRecord(key, value string) error")
		return err
	}

	body := bytes.NewReader(data)
	req, err := http.NewRequest(http.MethodPost, "https://"+g.addr+"/api/data", body)
	if err != nil {
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) UpdateRecord(key, value string) error")
		return err
	}

	_, err = g.Client.Do(req)
	if err != nil {
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) UpdateRecord(key, value string) error")
		return err
	}

	return nil
}

// Remove - удаляет данные их хранилища
func (g *GophKeeper) Remove(key string) error {

	eKey, err := gcrypto.Encrypt(key, g.key, "GophKeeper")
	if err != nil {
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) Remove(key string) error")
		return errors.New("шифрование ключа завершилось с ошибкой")
	}

	record := entities.Record{
		Key: eKey,
	}

	data, err := json.Marshal(record)
	if err != nil {
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) Remove(key string) error")
		return err
	}

	body := bytes.NewReader(data)
	req, err := http.NewRequest(http.MethodDelete, "https://"+g.addr+"/api/data", body)
	if err != nil {
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) Remove(key string) error")
		return err
	}

	_, err = g.Client.Do(req)
	if err != nil {
		slog.Error(err.Error(), "method:", "func (g *GophKeeper) Remove(key string) error")
		return err
	}

	return nil
}

// sync - синхронзиация с сервером
func (g *GophKeeper) sync() (int, error) {
	req, err := http.NewRequest(http.MethodGet, "https://"+g.addr+"/api/data?update="+strconv.Itoa(g.countUpdate), nil)
	if err != nil {
		slog.Error(err.Error(), "method:", "(g *GophKeeper) sync() (int, error)")
		return 0, err
	}

	resp, err := g.Client.Do(req)
	if err != nil {
		slog.Error(err.Error(), "method:", "(g *GophKeeper) sync() (int, error)")
		return 0, err
	}

	if resp.StatusCode > http.StatusNoContent {
		return resp.StatusCode, nil
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(err.Error(), "method:", "(g *GophKeeper) sync() (int, error)")
		return resp.StatusCode, err
	}

	if resp.StatusCode == http.StatusNoContent {
		return resp.StatusCode, nil
	}

	var update entities.Update
	err = json.Unmarshal(bytes, &update)

	if err != nil {
		slog.Error(err.Error(), "method:", "(g *GophKeeper) sync() (int, error)")
		return 0, err
	}

	g.applyUpdate(update)

	return resp.StatusCode, nil
}

// applyUpdate приминение обнолвения и новых данных
func (g *GophKeeper) applyUpdate(update entities.Update) bool {
	if g.countUpdate == update.Value {
		return false
	}

	newData := make([]entities.Record, len(update.Data))

	for idx, record := range update.Data {
		key, err := gcrypto.Decrypt(record.Key, g.key, "GophKeeper")
		if err != nil {
			slog.Error(err.Error(), "method:", "(g *GophKeeper) applyUpdate(update entities.Update) bool")
			return false
		}
		value, err := gcrypto.Decrypt(record.Value, g.key, "GophKeeper")
		if err != nil {
			slog.Error(err.Error(), "method:", "(g *GophKeeper) applyUpdate(update entities.Update) bool")
			return false
		}

		newData[idx].Key = key
		newData[idx].Value = value
	}

	g.countUpdate = update.Value
	g.data = newData
	g.Print()
	return true
}

// syncData - запускает синхронизацию данных
func (g *GophKeeper) syncData() {
	for {
		handlerSync(g)
		time.Sleep(2 * time.Second)
	}
}

// Print - вывод данных
func (g *GophKeeper) Print() {
	fmt.Println("\n========= BEGIN DATA =========")
	for _, record := range g.data {
		fmt.Printf("%s:%s\n", record.Key, record.Value)
	}
	fmt.Println("=========  END DATA  =========")
}

// handlerRequest
func handlerSync(g *GophKeeper) error {
	statusCode, err := g.sync()
	if err != nil {
		slog.Error(err.Error(), "method:", "handlerRequest(g *GophKeeper) error")
		return err
	}

	if statusCode == http.StatusOK {
		return nil
	}

	if statusCode == http.StatusUnauthorized {
		err = g.auth()
		if err != nil {
			slog.Error(err.Error(), "method:", "handlerRequest(g *GophKeeper) error")
			return err
		}

		statusCode, err = g.sync()

		if err != nil || statusCode != http.StatusOK {
			slog.Error(err.Error(), "method:", "handlerRequest(g *GophKeeper) error")
			return err
		}
	}

	return nil
}
