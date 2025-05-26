```
  __ _  ___ _ __ ___  ___  ___ _ ____   _____ 
 / _` |/ _ | '_ ` _ \/ __|/ _ | '__\ \ / / _ \
| (_| |  __| | | | | \__ |  __| |   \ V |  __/
 \__, |\___|_| |_| |_|___/\___|_|    \_/ \___|
 |___/          
```

Gemserve is a simple Gemini server written in Go.

Run tests and build:

```shell
make test #run tests only
make #run tests and build
```

Run:

```shell
./dist/gemserve
```

You'll need TLS keys, you can use `certs/generate.sh`
for quick generation.

## TODO
- [ ] Make TLS keys path configurable
- [ ] Fix slowloris (proper response timeouts)
