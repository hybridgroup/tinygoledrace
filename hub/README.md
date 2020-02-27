# Hub

## Prereqs on build machine

```
go get gobot.io/x/gobot
```

Bulding

```
GOARCH=arm go build -o toyhub ./hub/
```

## Prereqs on raspberry pi

```
sudo apt-get install mpg123
```

Copy the "audio" directory to raspberry pi


Copy file "toyhub" file from build machine to the raspberry pi

## Running on raspberry pi

```
./toyhub localhost:1883
```
