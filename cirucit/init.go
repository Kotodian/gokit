package cirucit

import (
	//"github.com/cep21/circuit"
	"time"

	circuitLib "github.com/edwardhey/circuit"
)

type CircuitBase struct {
	Name string
}

func (c CircuitBase) GetName() string {
	return c.Name
}

func (c *CircuitBase) Success(now time.Time, duration time.Duration) {
}

func (c *CircuitBase) Closed(now time.Time) {
}

func (c *CircuitBase) Opened(now time.Time) {
}

func (c *CircuitBase) ErrBadRequest(now time.Time, duration time.Duration) {
}

func (c *CircuitBase) ErrInterrupt(now time.Time, duration time.Duration) {
}

func (c *CircuitBase) ErrConcurrencyLimitReject(now time.Time) {
}

func (c *CircuitBase) ErrShortCircuit(now time.Time) {
}

func (c *CircuitBase) ErrFailure(now time.Time, duration time.Duration) {
}

func (c *CircuitBase) ErrTimeout(now time.Time, duration time.Duration) {
}

type ICircuitCloser interface {
	//hystrix.Opener
	//Circuit
	GetName() string
	circuitLib.OpenToClosed
}

var circuits *circuitLib.Manager

func init() {
	circuits = &circuitLib.Manager{}
}

func GetCM() *circuitLib.Manager {
	return circuits
}

//
//func NewCircuitWithOpenCloserAndMetrics(openerCloser ICircuit) *circuitLib.Circuit {
//	//openerCloser := NewCircuitOpenerCloser(NewCircuitReopenHealthCheck())
//	c := circuits.MustCreateCircuit(openerCloser.GetName(), circuitLib.Config{
//		//Execution: circuit.ExecutionConfig{
//		//	Timeout: 3 * time.Second,
//		//},
//		Metrics: circuitLib.MetricsCollectors{
//			Run: []circuitLib.RunMetrics{
//				openerCloser.(circuitLib.RunMetrics),
//			},
//			Circuit: []circuitLib.Metrics{
//				openerCloser.(circuitLib.Metrics),
//			},
//			//Fallback: []circuitLib.FallbackMetrics{
//			//	openerCloser.(circuitLib.FallbackMetrics),
//			//},
//		},
//	})
//	c.OpenToClose = openerCloser
//	c.ClosedToOpen = openerCloser
//	return c
//}
