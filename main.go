package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

var wg sync.WaitGroup

func checkUrl(address string, ch chan bool) {

	_, err := url.Parse(address)
	if err != nil {
		ch <- true
	}

	resp, err := http.Get(address)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error fetching url %s: %v", address, err)
	}
	log.Println(string(body))
	wg.Done()

}

// /["https://api.chucknorris.io/jokes/random", "https://jsonplaceholder.typicode.com/users"] rested
func main() {
	//todo: realize func, hadnle by ropute, []url,
	//check http.Get, if 1 url broken, send error
	//else -> download resource
	type T struct {
		Urls []string
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("Error reading", err)
			}
			u := T{}
			json.Unmarshal(reqBody, &u.Urls)

			wg.Add(len(u.Urls))

			ch := make(chan bool)

			for i := 0; i < len(u.Urls); i++ {
				go checkUrl(u.Urls[i], ch)
				if status := <-ch; !status {
					//return client -> error message
					log.Println("Error url")
				}
			}
			wg.Wait()
		}
	})
	http.ListenAndServe(":6969", nil)
	log.Println("start server..")
}

// приложение представляет собой http-сервер с одним хендлером+
// хендлер на вход получает POST-запрос со списком url в json-формате+
// сервер запрашивает данные по всем этим url и возвращает результат клиенту в json-формате+
// если в процессе обработки хотя бы одного из url получена ошибка, обработка всего списка прекращается и клиенту возвращается текстовая ошибка Ограничения:+

// использовать можно только компоненты стандартной библиотеки Go
// сервер не принимает запрос если количество url в в нем больше 20
// сервер не обслуживает больше чем 100 одновременных входящих http-запросов
// для каждого входящего запроса должно быть не больше 4 одновременных исходящих
// таймаут на запрос одного url - секунда
// обработка запроса может быть отменена клиентом в любой момент, это должно повлечь за собой остановку всех операций связанных с этим запросом
// сервис должен поддерживать 'graceful shutdown'
// результат должен быть выложен на github
