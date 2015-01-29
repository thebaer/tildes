~
=

Various scripts and programs for the tildeverse.

## tildelog
tildelog has moved to **[squigglelog](https://github.com/thebaer/squigglelog)**.

## code
Use **code** to generate a list of files contained within a given directory under their home folder. This was originally made to see who had a **Code** directory (this is the default), but you can specify whatever common directory you'd like to find. Do this:

```bash
go build code.go
./code -d bin
```

This outputs an HTML file in your `public_html/` folder based on the _templates/code.html_ template. The HTML file will be named after whatever directory you're scanning, so in this example, it'd be `public_html/bin.html`.

#### multiple sources
You can optionally supply multiple folders by passing a comma-separated list of folder to the `-d` flag. The first folder in the list will be used as the output file name.

```bash
go build code.go
./code -d Code,code,projects
```

This lists files in `/home/*/{Code,code,projects}/*`.
