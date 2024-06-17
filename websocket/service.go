package websocket

func Run() error {
	go RunProxy()

	select {}
}
