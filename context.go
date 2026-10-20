**“A context carries a cancellation signal and an optional deadline across goroutines and function calls. Calling cancel() or hitting the timeout closes the ctx.Done() channel, and the goroutines listening on it stop. I pass it as the first argument so long-running work like DB calls and HTTP requests can be cancelled.” **


The problem context solves

Say you start a goroutine to do some work, and later you want to tell it: “stop, we don’t need this anymore.” Goroutines can’t be killed from outside in Go. You can only ask them to stop, and they have to listen.

Context is the way to ask. It’s a stop signal you hand to a goroutine.

The analogy: a stop card

You give a worker a card and say: “Check this card now and then. If it says STOP, finish up and go home.”

The context is the card.
cancel() is you writing STOP on it.
ctx.Done() is the worker looking at the card.

Nothing is forced. The worker has to check the card.

Example 1: stopping a worker manually
go
package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done(): // card says STOP
			fmt.Println("worker stopping:", ctx.Err())
			return
		default: // card is blank, keep working
			fmt.Println("working...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go worker(ctx)

	time.Sleep(2 * time.Second)
	cancel() // write STOP on the card

	time.Sleep(time.Second) // give the worker time to print
}

Output:

working...
working...
working...
working...
worker stopping: context canceled

What happened:

WithCancel gave us a ctx (the card) and a cancel function (the pen).
The worker loops, checking ctx.Done() each time.
After 2 seconds, main calls cancel().
ctx.Done() becomes ready, and the worker sees it and returns.
What is ctx.Done()?

It’s just a channel. When someone calls cancel(), Go closes that channel. A receive from a closed channel returns immediately, so case <-ctx.Done() fires. The mailbox idea from earlier applies again.

Example 2: timeout

A timeout is the same card, but it writes STOP on itself after a set time.

go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

go worker(ctx) // same worker as above

After 2 seconds, the card says STOP automatically. The worker prints worker stopping: context deadline exceeded.

This is what your Shutdown code does:

go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
srv.Shutdown(ctx)

Shutdown is the worker. It waits for requests to finish, while also checking ctx.Done(). If 10 seconds pass first, it sees STOP and gives up waiting.

The two reasons a context stops
ctx.Err()	Meaning
context.Canceled	Someone called cancel()
context.DeadlineExceeded	The timeout ran out
Contexts pass downward

You give the same ctx to everything involved in one job:

go
func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := db.QueryContext(ctx, "SELECT ...")
}

r.Context() is cancelled automatically if the client disconnects. Because the database call received that same ctx, the query stops too, instead of wasting work for a user who left. When you cancel a parent context, all contexts created from it are cancelled as well.

4 rules to remember
Pass ctx as the first parameter: func doWork(ctx context.Context, ...).
Always defer cancel() right after WithCancel or WithTimeout.
Workers must check ctx.Done(). Cancelling doesn’t kill anything by itself.
Don’t store a context in a struct. Pass it through function calls.

