package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"

	"main/cache"
	"main/logger"
	"main/models"
	"main/msg_broker"
	"main/tools"
)

type Worker struct {
	id                   int
	waitGroup            *sync.WaitGroup
	cancelTasksReceiving context.CancelFunc
	cancelSignsReceiving context.CancelFunc
	taskRecepients       int
	broker               msg_broker.IMessageBroker
	cache                cache.ICache
	browser              *rod.Browser
	log                  *logger.Logger
	group                string
}

func New() *Worker {
	log := logger.Instance()
	lau := ""

	if os.Getenv("SEARCH_CHROMIUM") == "yes" {
		path, exists := launcher.LookPath()

		if !exists {
			log.Fatal("Chromium is not found")
		}

		lau = launcher.New().Bin(path).Headless(true).MustLaunch()
	} else {
		lau = launcher.New().Headless(true).MustLaunch()
	}

	id, err := strconv.Atoi(os.Getenv("ID"))

	if err != nil {
		id = 1
	}

	taskRecepients, err := strconv.Atoi(os.Getenv("TASK_RECEPIENTS"))

	if err != nil {
		taskRecepients = 1
	}

	return &Worker{
		id:             id,
		waitGroup:      &sync.WaitGroup{},
		taskRecepients: taskRecepients,
		broker:         msg_broker.NewKafkaBroker(os.Getenv("KAFKA_HOSTS")),
		cache:          cache.NewRedisCache(os.Getenv("REDIS_HOST")),
		browser:        rod.New().ControlURL(lau).MustConnect(),
		log:            log,
		group:          "Workers",
	}
}

func (w *Worker) runTaskRecipients(ctx context.Context) {
	w.waitGroup.Add(w.taskRecepients)

	for i := 0; i < w.taskRecepients; i++ {
		go w.broker.ReceivingWithWaitGroup(w.waitGroup, ctx, models.FindWordsTopic, w.group, w.handler)
	}
}

func (w *Worker) signsHandler(data []byte) {
	msg := &models.SignMsg{}
	err := json.Unmarshal(data, &msg)

	if err != nil {
		w.log.Error("Incorrect signal error: %v", err)
		return
	}

	if id, ok := msg.Services["worker"]; ok {

		if msg.Sign == models.Signals.Shutdown {
			if id == w.id {
				w.cancelTasksReceiving()
				w.cancelTasksReceiving = nil
				w.cancelSignsReceiving()
				w.log.Info("Closed with signal 'Shutdown'")
			}
			return
		}

		if msg.Sign == models.Signals.ScaleUpdate {

			if w.id <= id && w.cancelTasksReceiving == nil {
				ctx := context.Background()
				ctx, w.cancelTasksReceiving = context.WithCancel(ctx)

				w.runTaskRecipients(ctx)
				w.log.Info("Start with signal 'ScaleUpdate'")

			} else if w.id > id && w.cancelTasksReceiving != nil {
				w.cancelTasksReceiving()
				w.cancelTasksReceiving = nil
				w.log.Info("Pause with signal 'ScaleUpdate'")
			}
			return

		}

		if msg.Sign == models.Signals.Kill && id == w.id {
			w.log.Info("Closed with signal 'Kill'")
			os.Exit(0)
		}

	}
}

func (w *Worker) Run() {
	fw_ctx, s_ctx := context.Background(), context.Background()
	fw_ctx, w.cancelTasksReceiving = context.WithCancel(fw_ctx)
	s_ctx, w.cancelSignsReceiving = context.WithCancel(s_ctx)

	w.runTaskRecipients(fw_ctx)
	w.waitGroup.Add(1)
	go w.broker.ReceivingWithWaitGroup(w.waitGroup, s_ctx, models.SignalsTopic, w.group, w.signsHandler)

	w.waitGroup.Wait()
}

func (w *Worker) retry(msg *models.FindWordsMsg) (bool, error) {
	if msg.Attempts > 2 {
		return false, nil
	}

	msg.Attempts++
	return true, w.broker.Send(models.FindWordsTopic, msg)
}

func (w *Worker) callAggregator(msg *models.FindWordsMsg) error {
	AggMsg := models.AggregateResultMsg{RequestID: msg.RequestID}
	return w.broker.Send(models.AggregateResultTopic, AggMsg)
}

func (w *Worker) getPage(msg *models.FindWordsMsg) (*rod.Page, error) {
	page := w.browser.MustPage()

	page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36",
		AcceptLanguage: "en-US,en;q=0.9",
	})

	err := page.Navigate(msg.URL)

	if err != nil {
		page.Close()

		w.log.Error(
			"Request %d -> Getting document from %s navigation error: %v",
			msg.RequestID, msg.URL, err,
		)

		ok, err := w.retry(msg)

		if err != nil {
			return nil, err
		}

		if !ok {
			err = w.cache.SetURLToCompleteds(msg.RequestID, msg.URL, "fail", []string{})

			if err != nil {
				return nil, err
			}

			err = w.callAggregator(msg)

			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	}

	err = page.WaitLoad()

	if err != nil {
		w.log.Error(
			"Request %d -> Getting document from %s loading error: %v",
			msg.RequestID, msg.URL, err,
		)

		err = w.cache.SetURLToCompleteds(msg.RequestID, msg.URL, "unreaded", []string{})

		if err != nil {
			return nil, err
		}

		err = w.callAggregator(msg)

		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	return page, nil
}

func (w *Worker) findURLs(page *rod.Page, currUrl string, sameDomainOnly bool) []string {
	domain := tools.GetDomain(currUrl)
	urls := []string{}

	for _, el := range page.MustElements("a") {
		_url, err := el.Attribute("href")

		if err != nil || _url == nil || *_url == "" {
			continue
		}

		url := *_url

		if !strings.Contains(url, "://") {
			if url[0] == '/' {
				url = domain + url
			} else {
				url = strings.TrimRight(url, "/")

				if strings.HasSuffix(currUrl, url) {
					continue
				}
				url = fmt.Sprintf("%s/%s", currUrl, url)
			}
		} else if sameDomainOnly && !strings.HasPrefix(url, domain) {
			continue
		}

		index := strings.Index(url, "/#")

		if index != -1 {
			url = url[:index]
		}

		index = strings.LastIndex(url, "?")

		if index != -1 {
			url = url[:index]
		}

		urls = append(urls, strings.TrimRight(url, "/"))
	}

	return urls
}

func (w *Worker) findWords(page *rod.Page, words []string) []string {
	findedWords := make([]string, 0, len(words))
	listFinded := make([]bool, len(words))

	for _, el := range page.MustElements("*") {
		text := strings.ToLower(el.MustText())

		if text == "" {
			continue
		}

		for i := 0; i < len(words); i++ {
			if !listFinded[i] && strings.Contains(text, words[i]) {
				listFinded[i] = true
				findedWords = append(findedWords, words[i])

				if len(findedWords) == len(words) {
					return findedWords
				}
			}
		}
	}

	return findedWords
}

func (w *Worker) urlsToFindWordsMsgs(requestID int64, urls []string) []any {
	return tools.Select(urls, func(url string) any {
		return &models.FindWordsMsg{
			RequestID: requestID,
			URL:       url,
		}
	})
}

func (w *Worker) handler(data []byte) {
	w.log.Info("Start 'handler'")

	// parse message

	msg := models.FindWordsMsg{}
	err := json.Unmarshal(data, &msg)

	if err != nil || msg.URL == "" {
		w.log.Error("Incorrect message error: %v", err)
		return
	}

	isFinded, err := w.cache.URLIsCompleted(msg.RequestID, msg.URL)

	if err != nil {
		return
	}

	if isFinded {
		w.log.Info("Has been already completed 'handler' for request %d, URL %s", msg.RequestID, msg.URL)
		w.callAggregator(&msg)
		return
	}

	// get page from URL

	page, err := w.getPage(&msg)

	if err != nil || page == nil {
		return
	}

	defer page.Close()

	// get request data and find words

	requestData, err := w.cache.GetRequestData(msg.RequestID)

	if err != nil {
		return
	}

	// find and send anchecked urls

	urls := w.findURLs(page, msg.URL, requestData.SameDomainOnly)
	urls, err = w.cache.GetNotProcessedURLs(msg.RequestID, urls)

	if err != nil {
		return
	}

	messages := w.urlsToFindWordsMsgs(msg.RequestID, urls)
	err = w.broker.Send(models.FindWordsTopic, messages...)

	if err != nil {
		return
	}

	err = w.cache.SetURLsToAllURLs(msg.RequestID, urls)

	if err != nil {
		return
	}

	// find words and send result tp completed

	findedWords := w.findWords(page, requestData.Words)

	err = w.cache.SetURLToCompleteds(msg.RequestID, msg.URL, "success", findedWords)

	if err != nil {
		return
	}

	if len(urls) == 0 {
		w.log.Info("Completed 'handler' for request %d: Potentially last URL %s", msg.RequestID, msg.URL)
		w.callAggregator(&msg)
		return
	}

	w.log.Info("Completed 'handler' for request %d, URL %s", msg.RequestID, msg.URL)
}
