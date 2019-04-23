package main

import (
	serv "staply_img_resize/server"
)

func main() {
	serv.NewServer().Serve()
}
