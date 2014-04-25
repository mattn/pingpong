# pingpong

Ping Pong Task Managing

![](https://raw.githubusercontent.com/mattn/pingpong/master/etc/pingpong.png)

## Description

In generally, most of server processes are running forever. But in some use cases, you may want to terminate server process when you don't need.
This utilities are separeted `pping` and `ppong`. `pping` order to start process on server, and runs your command with ping to server. If `ppong` received ping, start server process and watching thems. If ping is not received while few seconds, `ppong` terminate the process.

## Usage

Start ppong

```
$ cat webcam.json
{
	"name": "ffserver",
	"args": ["-f", "webcam.conf"],
	"timeout": 30
}

$ ppong
```

Start ffmpeg command with ping

```
$ pping -n webcam ffmpeg -f vfwcap -r 25 http://127.0.0.1:8081/feed1.ffm
```

Then `ppong` start ffserver. While running ffmpeg, `pping` send pings, If ffmpeg is terminated, `pping` will exit.
After some seconds, `ppong` stop the ffserver.

## Configuration

* name: path to command
* args: arguments for the command
* timeout: specify timeout senconds. if omitted, 60 seconds.

## Options

`pping`

* `-n`: name of task. `ppong` read `name.json` on current director.json` on current directory.
* `-p`: interval seconds to ping.

`ppong`

* `-i`: interval seconds to check process alives.
* `-s`: server address.
* `-t`: default timeout value.

## Installation

pping on client

```
go get github.com/mattn/pingpong/pping
```

ppong on server
```
go get github.com/mattn/pingpong/ppong
```

# License

MIT

# Author

Yasuhiro Matsumoto (a.k.a mattn)

