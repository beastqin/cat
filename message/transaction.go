package message

import (
	"sync"
	"time"
)

type TransactionGetter interface {
	GetDuration() time.Duration
}

type Transactor interface {
	Messager
	TransactionGetter
	SetDuration(duration time.Duration)
	SetDurationStart(time time.Time)
	NewEvent(mType, name string) Messager
	LogEvent(mType, name string, args ...string)
	AddChild(m Messager)
	GetChildren() []Messager
	SetRootMessageId(rootMessageId string)
	SetParentMessageId(parentMessageId string)
	SetMessageId(messageId string)
}

type Transaction struct {
	Message

	children []Messager

	isCompleted bool

	mu sync.Mutex

	duration      time.Duration
	durationStart time.Time
}

func (t *Transaction) Complete() {
	if t.isCompleted {
		return
	}
	t.isCompleted = true

	if t.duration == 0 {
		t.duration = time.Now().Sub(t.Message.timestamp)
	}

	if t.Message.flush != nil {
		t.Message.flush(t)
	}
}

func (t *Transaction) GetDuration() time.Duration {
	return t.duration
}

func (t *Transaction) SetDuration(duration time.Duration) {
	t.duration = duration
}
func (t *Transaction) SetDurationStart(time time.Time) {
	t.durationStart = time
}

func (t *Transaction) NewEvent(mtype, name string) Messager {
	var e = NewEvent(mtype, name, nil)
	t.AddChild(e)
	return e
}

func (t *Transaction) LogEvent(mtype, name string, args ...string) {
	var e = t.NewEvent(mtype, name)
	if len(args) > 0 {
		e.SetStatus(args[0])
	}
	if len(args) > 1 {
		e.SetData(args[1])
	}
	e.Complete()
}

func (t *Transaction) AddChild(m Messager) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.children = append(t.children, m)
}

func (t *Transaction) GetChildren() []Messager {
	return t.children
}

func (t *Transaction) SetRootMessageId(rootMessageId string) {
	t.rootMessageId = rootMessageId
}

func (t *Transaction) SetParentMessageId(parentMessageId string) {
	t.parentMessageId = parentMessageId
}

func (t *Transaction) SetMessageId(messageId string) {
	t.messageId = messageId
}

func NewTransaction(mType, name string, flush Flush) *Transaction {
	return &Transaction{
		Message:       NewMessage(mType, name, flush),
		children:      make([]Messager, 0),
		isCompleted:   false,
		mu:            sync.Mutex{},
		duration:      0,
		durationStart: time.Time{},
	}
}
