#!/usr/bin/make -f
build:
	sudo curl https://get.ignite.com/cli! | sudo bash
	sudo ignite chain build --release --clear-cache
	sudo tar -xf release/*.tar.gz -C /usr/local/bin
