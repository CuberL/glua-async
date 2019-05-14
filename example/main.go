package main

import (
	"log"
	"time"

	"github.com/CuberL/glua-async"
	"github.com/yuin/gopher-lua"
)

func Sleep(L *lua.LState) int {
	sec := L.CheckInt(1)
	caller := L.CheckString(2)

	async.AsyncRun(func() []lua.LValue {
		time.Sleep(time.Second * time.Duration(sec))
		return []lua.LValue{lua.LString(caller)}
	}, L)

	return 0
}

func main() {
	script := `
		function main(str)
			-- A
			async(function()
				for i=3, 1, -1 do
					result = await(sleep, 1, "a")
					print("it's "..result)
				end
			end)

			-- B
			async(function()
				for i=3, 1, -1 do
					result = await(sleep, 2, "b")
					print("it's "..result)
				end
			end)
			
			-- C
			async(function()
				for i=3, 1, -1 do
					result = await(sleep, 3, "c")
					print("it's "..result)
				end
			end)

			return str
		end
	`

	L := lua.NewState()
	defer L.Close()

	async.Init(L)

	L.SetGlobal("sleep", L.NewFunction(Sleep))

	err := L.DoString(script)
	if err != nil {
		log.Println(err)
	}

	L.CallByParam(lua.P{
		Fn:      async.WrapAsyncFunc(L, L.GetGlobal("main").(*lua.LFunction)),
		NRet:    1,
		Protect: true,
	}, lua.LString("hello"))

	for i := 1; i <= L.GetTop(); i++ {
		log.Println(L.Get(i).String())
	}
}
