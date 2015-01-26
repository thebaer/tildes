~
=

Various scripts and programs for the tildeverse.

## tildelog
Use **tildelog** to easily create a simple log for your tilde. [Create new posts](https://github.com/thebaer/tildes/tree/master/tildelog/entries#tildelog-entries) in the `tildelog/entries/` folder, then run this:

```bash
cd tildelog
go build tildelog.go
./tildelog -template mytildelog
```

This will use any template in `tildelog/templates/` defined with _mytildelog_ (see below) to generate your full tildelog page.

#### templates

Your template should look like this.

```html
{{define "mytildelog"}}
<html>
	<head>
		<title>My ~log!</title>
	</head>
	<body>
		<h1>~log</h1>
		<p>Welcome to my ~log.</p>
		{{template "log" .}}
	</body>
</html>
{{end}}
```

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
