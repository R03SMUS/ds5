Start server, with port `Localhost:42069`
```bash
go run .\Server.go
```

Start as many replicas as you want, where the id flag is the port, default (50000)
```bash
go run .\Backup.go -id 50000
```

Start as many Clients as you want, where id is the client id, defualt (1)
```bash
go run .\Backup.go -id 1
```

in the clients you have 3 options
Request, Request the highest current bid
```bash
/request
```

Bid, where x is an int e.g. 200
```bash
/bid x
```

Exit, exits the program
```bash
/exit
```
