package tcp

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

var ErrPoolExhausted = errors.New("connection pool exhausted")

type Pool struct {
	// 设置连接函数
	Dial func() (net.Conn, error)
	// 设置连接的上下文
	DialContext func(ctx context.Context) (net.Conn, error)
	// 池中的最大的空闲连接数
	MaxIdle int
	// 池中的最大的非空闲连接数
	MaxActive int
	// 空闲连接的存活时间，如果长久不用就关闭了
	IdleTimeout time.Duration
	// 如果连接池的非空闲连接达到最大值, 是否会等待新的空闲连接,如果 设置成true的话就会一直等待一个连接空闲
	Wait bool
	// 连接的最大存活时间
	MaxConnLifeTime time.Duration
	// 锁 为了保护下面的字段的更改
	mu sync.Mutex
	// 确认连接池是否关闭
	closed bool
	// 当前的非空闲的连接数量
	active int
	// init函数 全局就一次
	initOnce sync.Once
	// 如果Wait为true时会使用到
	ch chan struct{}
	// 空闲连接链表
	idle idleList
	// 当前在等待的连接数量
	waitCount int64
	// 等待新的连接的时间
	waitDuration time.Duration
}

func NewPool(newFn func() (conn net.Conn, err error), maxIdle int) *Pool {
	return &Pool{Dial: newFn, MaxIdle: maxIdle}
}

func (p *Pool) lazyInit() {
	p.initOnce.Do(func() {
		p.ch = make(chan struct{}, p.MaxActive)
		if p.closed {
			close(p.ch)
		} else {
			for i := 0; i < p.MaxActive; i++ {
				p.ch <- struct{}{}
			}
		}
	})
}

func (p *Pool) Get() net.Conn {
	c, _ := p.GetContext(context.Background())
	return c
}

func (p *Pool) GetContext(ctx context.Context) (net.Conn, error) {
	waited, err := p.waitVacantConn(ctx)
	if err != nil {
		return errorConn{err}, err
	}

	p.mu.Lock()

	if waited > 0 {
		p.waitCount++
		p.waitDuration += waited
	}

	// 清理下没用的连接
	if p.IdleTimeout > 0 {
		n := p.idle.count
		for i := 0; i < n && p.idle.back != nil && p.idle.back.t.Add(p.IdleTimeout).Before(time.Now()); i++ {
			pc := p.idle.back
			p.idle.popBack()
			p.mu.Unlock()
			pc.c.Close()
			p.mu.Lock()
			p.active--
		}
	}

	// 获取空闲连接
	for p.idle.front != nil {
		pc := p.idle.front
		p.idle.popFront()
		p.mu.Unlock()
		if p.MaxConnLifeTime == 0 || time.Now().Sub(pc.created) < p.MaxConnLifeTime {
			return &activeConn{p: p, pc: pc}, nil
		}
		pc.c.Close()
		p.mu.Lock()
		p.active--
	}

	// 检查该池是否已经关闭了
	if p.closed {
		p.mu.Unlock()
		err := errors.New("get on closed pool")
		return errorConn{err}, err
	}

	if !p.Wait && p.MaxActive > 0 && p.active >= p.MaxActive {
		p.mu.Unlock()
		return errorConn{ErrPoolExhausted}, ErrPoolExhausted
	}

	p.active++
	p.mu.Unlock()
	c, err := p.dial(ctx)
	if err != nil {
		p.mu.Lock()
		p.active--
		if p.ch != nil && !p.closed {
			p.ch <- struct{}{}
		}
		p.mu.Unlock()
		return errorConn{err}, err
	}
	return &activeConn{p: p, pc: &poolConn{c: c, created: time.Now()}}, nil
}

func (p *Pool) waitVacantConn(ctx context.Context) (waited time.Duration, err error) {
	if !p.Wait || p.MaxActive <= 0 {
		return 0, nil
	}

	p.lazyInit()

	wait := len(p.ch) == 0
	var start time.Time
	if wait {
		start = time.Now()
	}

	select {
	case <-p.ch:
		select {
		case <-ctx.Done():
			p.ch <- struct{}{}
			return 0, ctx.Err()
		default:
		}
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	if wait {
		return time.Since(start), nil
	}
	return 0, nil
}

func (p *Pool) dial(ctx context.Context) (net.Conn, error) {
	if p.DialContext != nil {
		return p.DialContext(ctx)
	}
	if p.Dial != nil {
		return p.Dial()
	}
	return nil, errors.New("must pass Dial or DialContext to pool")
}

func (p *Pool) put(pc *poolConn, forceClose bool) error {
	p.mu.Lock()
	if !p.closed && !forceClose {
		pc.t = time.Now()
		p.idle.pushFront(pc)
		if p.idle.count > p.MaxIdle {
			pc = p.idle.back
			p.idle.popBack()
		} else {
			pc = nil
		}
	}

	if pc != nil {
		p.mu.Unlock()
		pc.c.Close()
		p.mu.Lock()
		p.active--
	}

	if p.ch != nil && !p.closed {
		p.ch <- struct{}{}
	}
	p.mu.Unlock()
	return nil
}

type idleList struct {
	count int
	// 最前端以及最后端的连接
	front, back *poolConn
}

type poolConn struct {
	c       net.Conn
	t       time.Time
	created time.Time
	// 双向链表
	next, prev *poolConn
}

func (l *idleList) pushFront(pc *poolConn) {
	// 将当前连接的下个连接接到当前最前端的连接
	pc.next = l.front
	// 因为当前连接就是最前端的连接 所以没有前端的连接了
	pc.prev = nil
	// 如果现在没有空闲连接 那么说明该连接既是最前端也是最后端的连接 否则就把最前端的连接的前一个连接设置成当前链接
	if l.count == 0 {
		l.back = pc
	} else {
		l.front.prev = pc
	}
	// 将当前最前端连接置成新连接
	l.front = pc
	l.count++
}

func (l *idleList) popFront() {
	pc := l.front
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.next.prev = nil
		l.front = pc.next
	}
	pc.next, pc.prev = nil, nil
}

func (l *idleList) popBack() {
	pc := l.back
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.prev.next = nil
		l.back = pc.prev
	}
	pc.next, pc.prev = nil, nil
}

type activeConn struct {
	p     *Pool
	pc    *poolConn
	state int
}

func (a *activeConn) Read(b []byte) (n int, err error) {
	return a.pc.c.Read(b)
}

func (a *activeConn) Write(b []byte) (n int, err error) {
	return a.pc.c.Write(b)
}

func (a *activeConn) Close() error {
	return a.pc.c.Close()
}

func (a *activeConn) LocalAddr() net.Addr {
	return a.pc.c.LocalAddr()
}

func (a *activeConn) RemoteAddr() net.Addr {
	return a.pc.c.RemoteAddr()
}

func (a *activeConn) SetDeadline(t time.Time) error {
	return a.pc.c.SetDeadline(t)
}

func (a *activeConn) SetReadDeadline(t time.Time) error {
	return a.pc.c.SetReadDeadline(t)
}

func (a *activeConn) SetWriteDeadline(t time.Time) error {
	return a.pc.c.SetWriteDeadline(t)
}

type errorConn struct{ err error }

func (e errorConn) Read(b []byte) (n int, err error) {
	return 0, e.err
}

func (e errorConn) Write(b []byte) (n int, err error) {
	return 0, e.err
}

func (e errorConn) Close() error {
	return e.err
}

func (e errorConn) LocalAddr() net.Addr {
	return nil
}

func (e errorConn) RemoteAddr() net.Addr {
	return nil
}

func (e errorConn) SetDeadline(t time.Time) error {
	return e.err
}

func (e errorConn) SetReadDeadline(t time.Time) error {
	return e.err
}

func (e errorConn) SetWriteDeadline(t time.Time) error {
	return e.err
}
