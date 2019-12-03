package mq

var MessageQueueMap map[string]LFQueue

/*
count the num of topic in mongodb.And creat the lfQueue for all of them.
So this function must run after db.Init().
*/
func Init() {
}
