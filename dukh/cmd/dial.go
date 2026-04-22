package main

import dukhclient "github.com/vhula/grazhda/dukh/client"

func dial() (*dukhclient.Client, error) {
	return dukhclient.NewDefault()
}
