package flow

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestFlow(t *testing.T) {
	w := New[string]("test work flow")
	var r string

	w.AddFlow("upper", func(_ context.Context, s string) (string, error) {
		return strings.ToUpper(s), nil
	})
	w.AddFlow("trim", func(_ context.Context, s string) (string, error) {
		return strings.TrimSpace(s), nil
	})
	w.OnFinish(func(_ context.Context, s string) {
		r = s
	})

	w.Run()
	w.Send("   anb   ")
	w.ShutDown()

	if r != "ANB" {
		t.FailNow()
	}
}

func TestFlowThrough(t *testing.T) {
	steps := []string{"first", "second", "third"}
	var r []string

	w := New[[]string]("test work through")
	for i := range steps {
		step := steps[i]
		w.AddFlow(step, func(_ context.Context, s []string) ([]string, error) {
			return append(s, step), nil
		})
	}
	w.OnFinish(func(_ context.Context, s []string) {
		r = s
	})

	w.Run()
	w.Send(nil)
	w.ShutDown()

	if !reflect.DeepEqual(r, steps) {
		t.Fatalf("got: %v", r)
	}
}

func TestStartStep(t *testing.T) {
	steps := []string{"first", "second", "third"}
	var r []string

	w := New[[]string]("test start step")
	for i := range steps {
		step := steps[i]
		w.AddFlow(step, func(_ context.Context, s []string) ([]string, error) {
			return append(s, step), nil
		})
	}
	w.OnFinish(func(_ context.Context, s []string) {
		r = s
	})

	w.Run()
	w.Send(nil, StartFrom(steps[1]))
	w.ShutDown()

	if !reflect.DeepEqual(r, steps[1:]) {
		t.Fatalf("got: %v", r)
	}
}

func TestNoFlow(t *testing.T) {
	w := New[struct{}]("test no flow")
	w.Run()
	defer w.ShutDown()

	err := w.Send(struct{}{})
	if err != ErrNoFlow {
		t.FailNow()
	}
}

func TestClosed(t *testing.T) {
	w := New[struct{}]("test closed")
	w.Run()
	w.ShutDown()

	err := w.Send(struct{}{})
	if err != ErrClosed {
		t.FailNow()
	}
}

func TestRetry(t *testing.T) {
	retry := 3
	var r int
	w := New[int]("test retry")
	w.AddFlow("retry", func(_ context.Context, i int) (int, error) {
		t.Log("retry:", r)
		r += 1
		return i, ErrRetry
	}, MaxRetry(retry))

	w.Run()
	w.Send(0)
	w.ShutDown()

	if r != retry+1 {
		t.Fatalf("got: %d", r)
	}
}

func TestAbort(t *testing.T) {
	steps := []string{"first", "second", "third"}
	var r []string

	w := New[[]string]("test start step")
	w.AddFlow(steps[0], func(_ context.Context, s []string) ([]string, error) {
		r = append(r, steps[0])
		return s, nil
	})
	w.AddFlow(steps[1], func(_ context.Context, s []string) ([]string, error) {
		r = append(r, steps[1])
		return s, Abort(nil)
	})
	w.AddFlow(steps[2], func(_ context.Context, s []string) ([]string, error) {
		r = append(r, steps[2])
		return s, nil
	})

	w.Run()
	w.Send(nil)
	w.ShutDown()

	if !reflect.DeepEqual(r, steps[:2]) {
		t.Fatalf("got: %v", r)
	}
}

func TestGoto(t *testing.T) {
	steps := []string{"first", "second", "third"}
	var r []string

	w := New[[]string]("test start step")
	w.AddFlow(steps[0], func(_ context.Context, s []string) ([]string, error) {
		r = append(r, steps[0])
		return s, Goto(steps[2])
	})
	w.AddFlow(steps[1], func(_ context.Context, s []string) ([]string, error) {
		r = append(r, steps[1])
		return s, nil
	})
	w.AddFlow(steps[2], func(_ context.Context, s []string) ([]string, error) {
		r = append(r, steps[2])
		return s, nil
	})

	w.Run()
	w.Send(nil)
	w.ShutDown()

	if !reflect.DeepEqual(r, []string{steps[0], steps[2]}) {
		t.Fatalf("got: %v", r)
	}
}

func TestFail(t *testing.T) {
	want := errors.New("some error")
	var got error
	w := New[struct{}]("test fail")
	w.AddFlow("fail", func(_ context.Context, s struct{}) (struct{}, error) {
		return struct{}{}, want
	})
	w.OnFail(func(_ context.Context, s1 string, s2 struct{}, err error) {
		got = err
	})
	w.Run()
	w.Send(struct{}{})
	w.ShutDown()

	if got != want {
		t.FailNow()
	}
}

func TestWrapRetry(t *testing.T) {
	src := errors.New("some errro")
	err := Retry(src)
	if !errors.Is(err, ErrRetry) {
		t.FailNow()
	}
}

func TestWrapAbort(t *testing.T) {
	src := errors.New("some error")
	err := Abort(src)
	if !errors.Is(err, ErrAbort) {
		t.Fatalf("%v\n", err)
	}
}
