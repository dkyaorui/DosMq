package mq

import (
	Mongodb "DosMq/db/mongo"
	myRedis "DosMq/db/redis"
	"DosMq/modules"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var MessageQueueMap = make(map[string]*LFQueue)
var NewQueueChannel = make(chan string, 8)
var MessageProcessWorkers *threadSafeWorkers
var globalQuit = make(chan struct{})
var DelTopicChannel = make(chan string, 4)

type QueWorker struct {
	topicIDHex string
	source     chan string
	quit       chan struct{}
}

type threadSafeWorkers struct {
	sync.Mutex
	workers []*QueWorker
}

func (q *QueWorker) Start() {
	q.source = make(chan string)
	mongoUtils := Mongodb.Utils
	mongoUtils.OpenConn()
	mongoUtils.SetDB(mongoUtils.DBName)
	topicID, _ := primitive.ObjectIDFromHex(q.topicIDHex)
	findResult, err := mongoUtils.FindOne(modules.DbTopic, bson.M{"_id": topicID})
	if err != nil {
		log.Errorf("err:%+v", errors.WithMessage(err, "queue init wrong"))
		return
	}
	var topic modules.Topic
	if err = findResult.Decode(&topic); err != nil {
		log.Errorf("err:%+v", errors.WithMessage(err, "queue init wrong"))
		return
	}
	MessageQueueMap[q.topicIDHex] = NewQue(1024)
	if strings.Compare(strings.ToLower(topic.ProcessMessageType), "push") == 0 {
		go pushMessageToSubscriber(q.topicIDHex)
	}
	go redisToQue(q.topicIDHex)

	mongoUtils.CloseConn()
	for {
		select {
		case <-q.quit:
			log.Infof("%s que is quit", q.topicIDHex)
			return
		case msg := <-q.source:
			if strings.Compare(msg, q.topicIDHex) == 0 {
				log.Infof("%s que is quit", q.topicIDHex)
				return
			}
		}
	}
}

func (t *threadSafeWorkers) push(w *QueWorker) {
	t.Lock()
	defer t.Unlock()

	t.workers = append(t.workers, w)
}

func (t *threadSafeWorkers) notice(msg string) {
	for _, worker := range t.workers {
		worker.source <- msg
	}
}

func redisToQue(topicId string) {
	redisClient := &myRedis.RDbClient
	redisKey := modules.GetMqKeyByTopicId(topicId)
	msgQue := MessageQueueMap[topicId]
	for {
		redisMqLen, err := redisClient.LLen(redisKey)
		if err != nil {
			log.Errorf("[find error]:%+v", errors.WithMessage(err, "redis error"))
		}
		if redisMqLen == 0 {
			time.Sleep(1000)
		} else {
			for i := int64(0); i < redisMqLen; i++ {
				reply, err := redisClient.RPop(redisKey)
				if err != nil {
					log.Errorf("[find error]:%+v", errors.WithMessage(err, "redis error"))
				} else {
					message := reply.(modules.Message)
					for {
						mqErr := msgQue.Push(message)
						if mqErr != nil {
							log.Errorf("[push error]:%+v", errors.WithMessage(err, "redis error"))
							continue
						} else {
							break
						}
					}
				}
			}
		}
	}
}

func pushMessageToSubscriber(topicIdHex string) {
	// init
	msgQue := MessageQueueMap[topicIdHex]
	mongoUtils := Mongodb.Utils
	mongoUtils.OpenConn()
	mongoUtils.SetDB(mongoUtils.DBName)
	topicIdObj, err := primitive.ObjectIDFromHex(topicIdHex)
	if err != nil {
		log.Errorf("err:%+v", errors.WithMessage(err, "topic id is wrong"))
	}
	findResult, err := mongoUtils.FindOne(modules.DbTopic, bson.M{"_id": topicIdObj})
	if err != nil {
		log.Errorf("err:%+v", errors.WithMessage(err, "[find err] find the topic wrong"))
		return
	}
	var topic modules.Topic
	_ = findResult.Decode(&topic)
	mongoUtils.CloseConn()
	// main
	for {
		item, msgErr := msgQue.Pop()
		if msgErr != nil {
			if strings.Contains(msgErr.Error(), "lockFreeQueue is empty") {
				continue
			}
			log.Error("queue's item is not message")
			continue
		}
		message, ok := item.(modules.Message)
		if ok {
			subscribers, _ := topic.GetAllSubscribers()
			var wg = &sync.WaitGroup{}
			for _, subscriber := range subscribers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var req *http.Request
					var err error
					switch strings.ToUpper(subscriber.Method) {
					case http.MethodGet:
						req, err = methodGetRequest(message, subscriber)
					case http.MethodPost:
						req, err = methodPostRequest(message, subscriber)
					}
					if err != nil {
						log.Errorf("err:%+v", errors.WithMessage(err, "new request fail"))
						return
					}
					c := &http.Client{}
					_, err = c.Do(req)
				}()
			}
			wg.Wait()
		}
	}
}

func methodGetRequest(message modules.Message, subscriber modules.Subscriber) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, subscriber.Host+subscriber.Api, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Query().Add("message", message.Value)
	req.URL.Query().Add("message", message.Value)
	return req, err
}

func methodPostRequest(message modules.Message, subscriber modules.Subscriber) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, subscriber.Host+subscriber.Api, nil)
	if err != nil {
		return nil, err
	}
	req.Form.Add("message", message.Value)
	req.Form.Add("auth_key", subscriber.Key)
	return req, err
}

/*
count the num of topic in mongodb.And creat the lfQueue for all of them.
So this function must run after db.Init().
*/

/*
   start n goroutine to get message in redis and then push into queue. N is topic's number.
   start n goroutine to get message in queue and then push to subscriber. N is topic's number.
*/
func StartProcess() {
	mongoUtils := Mongodb.Utils
	mongoUtils.OpenConn()
	mongoUtils.SetDB(mongoUtils.DBName)
	defer mongoUtils.CloseConn()
	MessageProcessWorkers = &threadSafeWorkers{}
	result, err := mongoUtils.Distinct(modules.DbTopic, "_id", bson.M{})
	if err != nil {
		log.Errorf("[distinct err]:%+v", errors.WithMessage(err, "db error"))
		panic("mq init error")
	}
	topicArray := result.([]interface{})
	// start message in que send to subscribe and start message in redis send to que
	for _, item := range topicArray {
		topicId := item.(primitive.ObjectID)
		var worker = QueWorker{
			topicIDHex: topicId.Hex(),
			quit:       globalQuit,
		}
		MessageProcessWorkers.push(&worker)
		go worker.Start()
	}
	log.Info("mq init success")
	// listen new que
	for {
		select {
		case topicIdHex := <-NewQueueChannel:
			var worker = QueWorker{
				topicIDHex: topicIdHex,
				quit:       globalQuit,
			}
			MessageProcessWorkers.push(&worker)
			go worker.Start()
		case delTopicIdHex := <-DelTopicChannel:
			MessageProcessWorkers.notice(delTopicIdHex)
		}
	}
}
