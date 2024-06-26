package flow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	errs "github.com/pkg/errors"
)

var (
	ErrClosed  = errors.New("work flow is closed")
	ErrTimeout = errors.New("send timeout")
	ErrRetry   = errors.New("please retry")
	ErrNoFlow  = errors.New("no flows available")
	ErrAbort   = errors.New("abort")
)

func Retry(err error) error {
	return errs.Wrap(ErrRetry, err.Error())
}

func Abort(err error) error {
	return errs.Wrap(ErrAbort, err.Error())
}

type Goto string

func (s Goto) Error() string {
	return string(s)
}

type WorkOpt struct {
	Name      string
	CacheSize uint
}

// 创建一个工作流
func New[T any](name string) WorkFlow[T] {
	return &workFlow[T]{
		name: name,
		done: make(chan struct{}),
	}
}

// 工作流：表示一个工作流水线
//
// Usage:
//
//	var flw WorkFlow[string]
//	flw.AddFlow("upper", func(_ context.Context, s string) (string, error) {
//		return strings.ToUpper(s), nil
//	})
//	flw.AddFlow("trim", func(_ context.Context, s string) (string, error) {
//		return strings.TrimSpace(s), nil
//	})
//	flw.Run()
type WorkFlow[T any] interface {
	// 添加一个工作流，默认一个线程，出口缓存容量为1
	AddFlow(string, func(context.Context, T) (T, error), ...FlowOpt)
	// 向工作流发送一个数据，该方法会阻塞，直到提交成功，
	// 如果工作流已关闭，返回[ErrClosed]错误
	Send(T, ...RunOpt) error
	// 设置失败处理函数
	OnFail(func(context.Context, string, T, error))
	// 设置末端处理函数
	OnFinish(func(context.Context, T))
	// 启动工作流
	Run()
	// 关闭工作流
	ShutDown()
}

// 流程：若干个流程组成一个工作流
// 暂时不开放自定义实现
type stepFlow[T any] interface {
	// 流程名称
	// name() string
	// 运行流程
	run(out chan packet[T]) (in chan packet[T])
}

type FlowOpt interface {
	apply(*flowOpt)
}

type flowOpt struct {
	cacheSize int // channel大小
	workCount int // 并发量
	maxRetry  int // 最大重试次数
}

type CacheSize int // 设置channel缓存大小
type WorkCount int // 设置工作线程数量
type MaxRetry int  // 设置最大重试次数

func (c CacheSize) apply(opt *flowOpt) {
	n := int(c)
	if n < 0 {
		return
	}
	opt.cacheSize = int(c)
}

func (c WorkCount) apply(opt *flowOpt) {
	n := int(c)
	if n < 1 {
		return
	}
	opt.workCount = n
}

func (m MaxRetry) apply(opt *flowOpt) {
	n := int(m)
	if n < 0 {
		return
	}
	opt.maxRetry = n
}

type RunOpt interface {
	apply(*runOpt)
}

type runOpt struct {
	timeout   time.Duration
	startStep string
}

type Timeout string
type StartFrom string

func (t Timeout) apply(opt *runOpt) {
	var err error
	opt.timeout, err = time.ParseDuration(string(t))
	if err != nil {
		opt.timeout = time.Second * 5
	}
}

func (s StartFrom) apply(opt *runOpt) {
	opt.startStep = string(s)
}

// 工作流水线
type workFlow[T any] struct {
	lock      sync.RWMutex                            // 同步锁
	name      string                                  // 工作流名称
	input     chan packet[T]                          // 输入channel
	output    chan packet[T]                          // 输出channel
	done      chan struct{}                           // 结束channel
	steps     []stepFlow[T]                           // 工作流步骤
	finalFunc func(context.Context, T)                // 末端处理函数
	failFunc  func(context.Context, string, T, error) // 失败处理函数
	closed    bool                                    // 工作流关闭标志
}

// 添加工作流程
func (w *workFlow[T]) AddFlow(name string, worker func(context.Context, T) (T, error), opts ...FlowOpt) {
	opt := flowOpt{
		cacheSize: 1,
		workCount: 1,
		maxRetry:  3,
	}
	for i := range opts {
		opts[i].apply(&opt)
	}
	// 添加流程
	step := step[T]{
		name:      name,
		cacheSize: opt.cacheSize,
		workCount: opt.workCount,
		maxRetry:  opt.maxRetry,
		worker:    worker,
		flow:      w,
	}
	w.steps = append(w.steps, &step)
}

// 向工作流发送一条数据
func (w *workFlow[T]) Send(t T, opts ...RunOpt) error {
	var opt runOpt
	for i := range opts {
		opts[i].apply(&opt)
	}

	w.lock.RLock()
	defer w.lock.RUnlock()

	if w.closed {
		return ErrClosed // 已关闭
	}
	if len(w.steps) == 0 {
		return ErrNoFlow // 没有可用流程
	}
	// 发送数据到工作流
	p := packet[T]{
		ctx:       context.Background(),
		startStep: opt.startStep,
		data:      t,
	}
	if opt.timeout == 0 {
		w.input <- p
		return nil
	}
	select {
	case w.input <- p:
		return nil
	case <-time.After(opt.timeout):
		return ErrTimeout
	}
}

// 设置失败处理函数
func (w *workFlow[T]) OnFail(fn func(context.Context, string, T, error)) {
	w.failFunc = fn
}

// 设置末端处理函数
func (w *workFlow[T]) OnFinish(fn func(context.Context, T)) {
	w.finalFunc = fn
}

// 启动工作流
func (w *workFlow[T]) Run() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.closed {
		panic(ErrClosed)
	}

	w.output = make(chan packet[T], 1)
	var out = w.output
	// 依次启动每个流程
	for i := len(w.steps) - 1; i >= 0; i-- {
		out = w.steps[i].run(out)
	}
	w.input = out

	go func() {
		for p := range w.output {
			if w.finalFunc != nil && p.startStep == "" {
				w.handle(p)
			}
		}
		close(w.done) // 所有流程都已结束
	}()

	slog.Info("work flow start", "name", w.name)
}

// 关闭工作流
func (w *workFlow[T]) ShutDown() {
	w.lock.Lock()
	w.closed = true // 设置关闭状态
	close(w.input)  // 关闭输入channel
	w.lock.Unlock()

	<-w.done // 等待所有工作流结束
	slog.Info("work flow shutdown", "name", w.name)
}

func (w *workFlow[T]) handle(t packet[T]) {
	if w.finalFunc == nil {
		return
	}
	// 防止panic
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recover final", "error", r)
		}
	}()

	w.finalFunc(t.ctx, t.data)
}

func (w *workFlow[T]) isClosed() bool {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.closed
}

// 工作步骤
type step[T any] struct {
	wg        sync.WaitGroup
	flow      *workFlow[T] // 工作流
	name      string       // 步骤名称
	cacheSize int          // channel大小
	workCount int          // 并发量
	maxRetry  int          // 最大重试次数
	worker    func(context.Context, T) (T, error)
}

// func (s *step[T]) name() string {
// 	return name
// }

func (s *step[T]) handle(t packet[T]) (r packet[T], err error) {
	// 防止panic
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	// 处理业务
	r.data, err = s.worker(t.ctx, t.data)
	r.ctx = t.ctx
	return
}

func (s *step[T]) handleError(t packet[T], err error) {
	// 防止panic
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recover onFail", "name", s.name, "output", r)
		}
	}()

	slog.Error("work flow failed", "step", s.name, "error", err)
	// 错误处理
	if s.flow.failFunc != nil {
		s.flow.failFunc(t.ctx, s.name, t.data, err)
	}
}

// 运行处理流程
func (s *step[T]) run(out chan packet[T]) (in chan packet[T]) {
	in = make(chan packet[T], s.cacheSize) // 输入channel
	s.wg.Add(s.workCount)

	go func(in, out chan packet[T]) {
		for i := 0; i < s.workCount; i++ {
			// 启动工作线程
			go func(in, out chan packet[T], n int) {
				defer s.wg.Done()
				slog.Info("start work step", "name", s.name, "number", n)

				// 开始流程循环
				for input := range in {
					// slog.Debug("start handle", "name", s.name, "input", input)
					if input.startStep != "" && input.startStep != s.name {
						// 起始步骤与当前步骤不匹配
						out <- input // goto next step
						continue
					}
				RETRY:
					output, err := s.handle(input)
					// slog.Debug("handle result", "name", s.name, "output", output, "err", err)

					// handle error
					gotoStep, ok := err.(Goto)
					switch {
					case err == nil:
						out <- output // goto next step
						continue
					case ok:
						output.startStep = string(gotoStep)
						out <- output
						continue
					case errors.Is(err, ErrAbort):
						// s.flow.output <- output
						continue
					case errors.Is(err, ErrRetry):
						if input.retry < s.maxRetry {
							input.retry++
							time.Sleep(time.Millisecond)
							goto RETRY // 原地重试
						}
						// 重试失败
						fallthrough
					default:
						s.handleError(output, err)
					}
					// handle next packet
				}

				// 流程退出
				slog.Info("work step stoped", "name", s.name, "number", n)
			}(in, out, i)
		}
		s.wg.Wait() // 等待所有工作线程退出
		close(out)  // 关闭下一 步骤
	}(in, out)
	return
}

type packet[T any] struct {
	ctx       context.Context
	retry     int    // 重试次数
	startStep string // 当前步骤
	data      T
}
