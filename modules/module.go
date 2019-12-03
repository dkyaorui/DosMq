package modules

import myMongoModule "DosMq/modules/mongo"

type AuthKey struct {
    Value string `json:"auth_key" binding:"required"`
}

type MessageChannel struct {
    ch   chan myMongoModule.Message
    name string
    size int
}

func (m *MessageChannel) Init(name string, size int) {
    m.ch = make(chan myMongoModule.Message, size)
    m.name = name
    m.size = size
}
