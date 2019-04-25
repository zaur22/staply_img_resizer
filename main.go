package main

import (
	serv "staply_img_resizer/server"
)

func main() {
	serv.NewServer().Serve()
}
