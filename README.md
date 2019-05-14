# glua-async

An async/await implement for [gopher-lua](https://github.com/yuin/gopher-lua).

## Installation

``` shell

go get github.com/CuberL/glua-async

```

## Example Usage

First, you need to modify your blocking function by wrapping the blocking part(sleep, i/o, etc) in `async.AsyncRun`. And the return values of the `async.AsyncRun` will be the return values of the `await`.

``` go

func Sleep(L *lua.LState) int {
    sec := L.CheckInt(1)
    caller := L.CheckString(2)

    async.AsyncRun(func() []lua.LValue {
        time.Sleep(time.Second * time.Duration(sec))
        return []lua.LValue{lua.LString(caller)}
    }, L)

    return 0
}

```

And you can use it in your function. Use `async` to start a coroutine and use `await` to wait some blocking result in an async function.

``` lua
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
```

Finally, you can call `main` function, and you also need to wrap your function in `async.WrapAsyncFunc`.

``` go

L.CallByParam(lua.P{
    Fn:      async.WrapAsyncFunc(L, L.GetGlobal("main").(*lua.LFunction)),
    NRet:    1,
    Protect: true,
}, lua.LString("hello"))

for i := 1; i <= L.GetTop(); i++ {
    log.Println(L.Get(i).String())
}

```