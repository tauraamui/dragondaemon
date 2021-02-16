# Dragon Daemon - An RTSP based streaming server

Connect to multiple RTSP based streams (IP cameras) and save timestamped clips to a local directory as specified within the configuration. Following the UNIX philosophy of one program to one thing well, the idea is to be able to run this concurrently with a separate media server which will serve the clips to a DASH based player within a web interface. Future features include: motion detection, facial detection, object categorization, zones. It records everything live initially to then go back and remove clips which don't match certain event criteria after analysis.

![terminalexample](/doc/screenshots/terminal.png)

## Example dd.config file
```
{
    "debug": true,
    "cameras": [
        {
            "disabled": false,
            "title": "Back2",
            "fps": 30,
            "address": "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov",
            "seconds_per_clip": 2,
            "persist_location": "/Users/adam/Movies/clips"
        }
    ]
}
```

## Getting Started

At this time this project will need to be built before being able to run.

### Prerequisites

- Go
- GoCV library which has additional setup instructions


### Installing

Install Go

Download this repo via go get
```
go get github.com/tauraamui/dragondaemon
```

Run the setup process for GoCV to install and build OpenCV deps on your host
```
cd $GOPATH/gocv.io/x/gocv
make install
```

## Running the tests

Running the tests just as normal
```
go test -v ./...
```

## Running

Run the main.go or build and run compiled version. By default it will look for a local dd.config file.

Compile
```
go build -o dragondaemon
```

Run
```
./dragondaemon
```

## Deployment

You can install the built binary as a service by running

```
./dragond install
```

Run it via
```
./dragond start
```

Stop it via
```
./dragond stop
```

Check the status with
```
./dragond status
```

Uninstall with
```
./dragond remove
```

## Built With

* [Go mod]() - Dependency Management
* [logging](https://github.com/tacusci/logging) - Logging library
* [GoCV](https://gocv.io/x/gocv/) - Go wrapper for OpenCV

## Contributing

~~Please read [CONTRIBUTING.md]() for details on our code of conduct, and the process for submitting pull requests to us.~~

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/tauraamui/dragondaemon/tags). 

## Authors

* **Adam Prakash Stringer** - *Initial work* - [tauraamui](https://github.com/tauraamui)

See also the list of [contributors](https://github.com/tauraamui/dragondaemon/contributors) who participated in this project.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgments

* Many thanks to the GoCV team without whom this would not have been possible
* My wife for putting up with me spending evenings on this instead of with her
