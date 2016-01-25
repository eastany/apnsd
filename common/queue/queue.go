package queue

type Queue interface {
	Pop(qName string) []byte
	Push(qName string, data []byte)
}
