# cqut-photo

Just a script can grab students' photoes in education system

# Use
```go
go build -o cqut
./cqut run
```

## More
```
run: start to run script
clean: delete the break file
-----------------------------
config.json
{
	from: int, [the grade of start]
	to : int, [the grade of end]
	peopleCount: int, [the total number of a class]
	duration: float64, [how long does script run once]
	username: string,
	password: string
}
```