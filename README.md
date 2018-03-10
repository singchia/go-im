## 1. Background
This project is a practice of [**pricise-concurrency**](http://www.singchia.com/2018/01/29/Concurrency-Patterns-Summary-And-Implementation/) which is described as **_"Since goroutines and threads are stateless, we can combine them with some certain resources for precise control"._**   

So **go-im** combine each goroutines with a specific channel to get the mapping **< goroutine, connection/channel >**, and this channel represents an user data tunnel. After parsing the data from the tunnel, **go-im** build another mapping **< channel, user >**. Then the whole **instant-messaging** system can be build over those.  

**Client-less** is another original thougth and design. User don't need  to install or build anything else as long as you have had **telnet** or **netcat** installed.  

**NOTE: this project is just a demo and should not be used in product environment.**

## 2. Installation
### 2.1. Go environment
If you don't have the Go development environment installed, visit the [Getting Started](https://golang.org/doc/install) document and follow the instructions. You can alternatively choose one to install.

### 2.2. Intall in go way
After setting **GOPATH** and run command below you will get a binary file named **go-im** at **GOPATH/bin**.

```
go get -u github.com/singchia/go-im
```

### 2.3. Intall to current dir with Makefile

```
git clone https://github.com/singchia/go-im.git
make
```


## 3. How-to-use
### 3.1. Run the server

```
> ./go-im -addr 127.0.0.1:1202
```

### 3.2. Clinet in usings telnet

**User _foo_**

```
> telnet 127.0.0.1 1202
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
signup: foo
[from system] enter passward:passward
[from system] re-enter passward: passward
[from system] auth succeed.
to user: bar
how you doing
[from system] object does not exist.
create group: g1
[from system] group create succeed.
invite group: g1 bar
[from system] object does not exist.
```
**User _bar_**

```
> telnet 127.0.0.1 1202
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
signup: bar
[from system] enter passward:passward
[from system] re-enter passward: passward
[from system] auth succeed.
```
**User _foo_**

```
> telnet 127.0.0.1 1202
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
signup: foo
[from system] enter passward:passward
[from system] re-enter passward: passward
[from system] auth succeed.
to user: bar
how you doing
[from system] object does not exist.
create group: g1
[from system] group create succeed.
invite group: g1 bar
[from system] object does not exist.
to user: bar
how you doing
```

**User _bar_**

```
> telnet 127.0.0.1 1202
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
signup: bar
[from system] enter passward:passward
[from system] re-enter passward: passward
[from system] auth succeed.
[from user foo] how you doing
to user: foo
I'm doing good
```

**User _foo_**

```
> telnet 127.0.0.1 1202
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
signup: foo
[from system] enter passward:passward
[from system] re-enter passward: passward
[from system] auth succeed.
to user: bar
how you doing
[from system] object does not exist.
create group: g1
[from system] group create succeed.
invite group: g1 bar
[from system] object does not exist.
to user: bar
how you doing
[from user bar] I'm doing good
```

### 3.3. Client in using netcat
**User _anonymous_**

```
> nc 127.0.0.1 1202
signup: anonymous
[from system] enter passward:passward
[from system] re-enter passward:passward
[from system] auth succeed
to user: foo
你好
```

## 4. User Interface

| command| function |
|:---|:---:|
|signup:{username}|sign up|
|signin:{username}|sign in|
|signout|sign out|
|to user:{user name}|send message to an user|
|to group:{group name}|send message to a group|
|create group:{group name}|create a group|
|join group:{group name}|join group|
|invite group:{group name} {user name}|invite a user join group|
