package rabbitmq

type rabbitmqHook struct {
	queue   string
	service string
}

func NewRabbitmqHook(queue string, service string) *rabbitmqHook {
	return &rabbitmqHook{
		queue:   queue,
		service: service,
	}
}

func (r *rabbitmqHook) Write(p []byte) (n int, err error) {
	return 0, nil
}
