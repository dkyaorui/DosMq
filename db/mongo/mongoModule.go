package mongo

import "time"

type RequestMessage struct {
    RecvRequest []byte `bson:"receive_request"`
    Created time.Time `bson:"created"`
}
