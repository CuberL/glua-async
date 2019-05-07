package main

import (
	"log"
	"time"

	"github.com/CuberL/glua-async"
	"github.com/yuin/gopher-lua"
)

func Sleep(L *lua.LState) int {
	sec := L.CheckInt(1)
	async.AsyncRun(func() []lua.LValue {
		time.Sleep(time.Second * time.Duration(sec))
		return []lua.LValue{}
	}, L)
	return 0
}

func main() {
	script := `
		function work(str)
			await(sleep, 3)
			async(function()
				for i=3, 1, -1 do
					await(sleep, 1)
					print("it's a")
				end
			end)

			async(function()
				for i=3, 1, -1 do
					await(sleep, 2)
					print("it's b")
				end
			end)
			return str
		end
	`

	// script_2 := `
	// await(sleep, 3)
	// print('hi')
	// `

	L := lua.NewState()
	defer L.Close()

	async.Init(L)
	L.SetGlobal("sleep", L.NewFunction(Sleep))

	err := L.DoString(script)
	if err != nil {
		log.Println(err)
	}

	// th, _ := L.NewThread()
	// _, _, rets := L.Resume(th, L.GetGlobal("__wrap").(*lua.LFunction), L.GetGlobal("work").(*lua.LFunction), lua.LString("input"))

	L.CallByParam(lua.P{
		Fn:      async.WrapFunc(L, L.GetGlobal("work").(*lua.LFunction)),
		NRet:    1,
		Protect: true,
	}, lua.LString("hello"))

	for i := 1; i <= L.GetTop(); i++ {
		log.Println(L.Get(i).String())
	}
}
