# twarc (Twitter Archiver)

## Usage

Quit Chrome browser and make the command launch the browser with the CDP port open.

```bash
$ twitter-archive -u user -o /path/to/output.json
```

Alternatively, start your browser with `--remote-debugging-port=9222` (any port number will do) and pass the port number to the command.

```bash
$ /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222
$ twitter-archive -p 9222 -u user -o /path/to/output.json
```

You can also fetch search results like https://x.com/explore by passing the query string.

```bash
$ twitter-archive -q "from:user since:2020-01-01 until:2010-02-01" -o /path/to/output.json
```
