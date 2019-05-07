package async

import (
	"log"
	"sync"

	"github.com/yuin/gopher-lua"
)

type AsyncState struct {
	channel chan *AsyncResult
	wg      *sync.WaitGroup
}

type AsyncResult struct {
	co     *lua.LState
	result []lua.LValue
}

func AsyncRun(fn func() []lua.LValue, L *lua.LState) {
	_channel := L.GetGlobal("__state").(*lua.LUserData)
	s := _channel.Value.(*AsyncState)
	s.wg.Add(1)

	go func() {
		ret := fn()
		s.channel <- &AsyncResult{
			co:     L,
			result: ret,
		}
	}()
}

func WrapFunc(L *lua.LState, fn *lua.LFunction) *lua.LFunction {
	return L.NewFunction(func(L *lua.LState) int {
		co, _ := L.NewThread()
		args := []lua.LValue{}
		for i := 1; i <= L.GetTop(); i++ {
			args = append(args, L.Get(i))
		}
		_, _, rets := L.Resume(co, fn, args...)
		if len(rets) > 0 && !(len(rets) == 1 && rets[0] == lua.LNil) {
			Schedule(L)
		} else {
			rets = Schedule(L)
		}
		for _, ret := range rets {
			L.Push(ret)
		}
		return len(rets)
	})
}

func Init(L *lua.LState) {
	awaitScript := `
	function await(fn, ...)
		fn(...)
		return coroutine.yield()
	end
	
	function async(fn, ...)
		coroutine.resume(coroutine.create(fn), ...)
	end
	`
	tmpL := lua.NewState()
	tmpL.NewThread()
	tmpL.DoString(awaitScript)
	awaitFunc := tmpL.GetGlobal("await").(*lua.LFunction).Proto
	asyncFunc := tmpL.GetGlobal("async").(*lua.LFunction).Proto

	s := &AsyncState{}
	s.channel = make(chan *AsyncResult)
	s.wg = &sync.WaitGroup{}
	ud := L.NewUserData()
	ud.Value = s

	L.SetGlobal("__state", ud)
	L.SetGlobal("await", L.NewFunctionFromProto(awaitFunc))
	L.SetGlobal("async", L.NewFunctionFromProto(asyncFunc))
	// L.SetGlobal("__wrap", L.NewFunction(WrapFunc))
}

func Schedule(L *lua.LState) []lua.LValue {
	_channel := L.GetGlobal("__state").(*lua.LUserData)
	s := _channel.Value.(*AsyncState)
	var vals []lua.LValue

	go func() {
		s.wg.Wait()
		close(s.channel)
	}()

	for {
		select {
		case a := <-s.channel:
			if a == nil {
				return vals
			}
			_, _, _vals := L.Resume(a.co, nil, a.result...)
			log.Println(_vals)
			if len(_vals) > 0 && !(len(_vals) == 1 && _vals[0] == lua.LNil) {
				vals = _vals
			}
			s.wg.Done()
		}
	}
	return []lua.LValue{}
}
