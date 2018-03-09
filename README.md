# 1. Background
This project is a pratice of [**pricise-concurrency**](http://www.singchia.com/2018/01/29/Concurrency-Patterns-Summary-And-Implementation/) which is described as **_"Since goroutines and threads are stateless, we can combine them with some certain resources for precise control"._**   

So **go-im** combine each goroutines with a specific channel to get the mapping **< goroutine, connection/channel >**, and this channel represents an user data tunnel. After parsing the data from channel, **go-im** build another mapping **< channel, user >**. Then the whole **instant-messaging** system can be build on the those.  

**Client-less** is another original thougth and design. User don't need  to install or build anything else as long as you have had **telnet** or **netcat** installed.  

# 2. Installation
## 2.1. go environment
If you don't have the Go development environment installed, visit the [Getting Started](https://golang.org/doc/install) document and follow the instructions. You can alternatively choose one way 

## 2.2 intall in go way
After setting **GOPATH** and run command below you will get a binary file named **go-im** at GOPATH/bin.

```
go get -u github.com/singchia/go-im
```

## 2.3 intall with make

# How-to-use

## Start 